package service

import (
	"context"
	"fmt"
	"time"

	customerrors "github.com/alexalbu001/iguanas-jewelry/internal/customErrors"
	"github.com/alexalbu001/iguanas-jewelry/internal/models"
	"github.com/alexalbu001/iguanas-jewelry/internal/transaction"
	"github.com/alexalbu001/iguanas-jewelry/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type ShippingInfo struct {
	Name         string `json:"name" binding:"required"`
	Email        string `json:"email" binding:"required,email"`
	Phone        string `json:"phone" binding:"required"`
	AddressLine1 string `json:"address_line1" binding:"required"`
	AddressLine2 string `json:"address_line2"` // Optional
	City         string `json:"city" binding:"required"`
	State        string `json:"state" binding:"required"`
	PostalCode   string `json:"postal_code" binding:"required"`
	Country      string `json:"country" binding:"required"`
}

type OrdersStore interface {
	InsertOrder(ctx context.Context, order models.Order) error
	InsertOrderTx(ctx context.Context, order models.Order, tx pgx.Tx) error
	InsertOrderItem(ctx context.Context, orderItem models.OrderItem) error
	InsertOrderItemTx(ctx context.Context, orderItem models.OrderItem, tx pgx.Tx) error
	InsertOrderItemBulk(ctx context.Context, orderItems []models.OrderItem) error
	InsertOrderItemBulkTx(ctx context.Context, orderItems []models.OrderItem, tx pgx.Tx) error
	GetOrderByID(ctx context.Context, orderID string) (models.Order, error)
	GetOrderItems(ctx context.Context, orderID string) ([]models.OrderItem, error)
	GetOrderItemsBatch(ctx context.Context, orderID []string) (map[string][]models.OrderItem, error)
	GetUsersOrders(ctx context.Context, userID string) ([]models.Order, error)
	GetAllOrders(ctx context.Context) ([]models.Order, error)
	GetOrdersByStatus(ctx context.Context, status string) ([]models.Order, error)
	UpdateOrderStatus(ctx context.Context, status, orderID string) error
}

type OrdersService struct {
	orderStore    OrdersStore
	productsStore ProductsStore
	cartsStore    CartsStore
	TxManager     transaction.TxManager
}

func NewOrderService(orderStore OrdersStore, productStore ProductsStore, cartsStore CartsStore, TxManager transaction.TxManager) *OrdersService {
	return &OrdersService{
		orderStore:    orderStore,
		productsStore: productStore,
		cartsStore:    cartsStore,
		TxManager:     TxManager,
	}
}

type OrderOperationResult struct {
	OrderSummary OrderSummary
	Error        string
}

type StatusOrderResult struct {
	OrderID string
	Status  string
	Message string
	Success bool
}

type OrderSummary struct {
	ID              string
	UserID          string
	OrderItems      []OrderItemSummary
	Total           float64
	Status          string
	ShippingName    string
	ShippingAddress ShippingAddress
	// PaymentMethod   string
	CreatedDate time.Time
}
type ShippingAddress struct {
	AddressLine1 string
	AddressLine2 string
	City         string
	State        string
	PostalCode   string
	Country      string
	Email        string
	Phone        string
}

type OrderItemSummary struct {
	ID          string
	ProductID   string
	ProductName string
	Price       float64
	Quantity    int
	Subtotal    float64
}

// Helper functions - add these to your service
func (o *OrdersService) buildOrderItemSummary(orderItem models.OrderItem, product models.Product) OrderItemSummary {
	return OrderItemSummary{
		ID:          orderItem.ID,
		ProductID:   product.ID,
		ProductName: product.Name,
		Price:       orderItem.Price,
		Quantity:    orderItem.Quantity,
		Subtotal:    orderItem.Price * float64(orderItem.Quantity),
	}
}

func (o *OrdersService) buildShippingAddress(order models.Order) ShippingAddress {
	return ShippingAddress{
		AddressLine1: order.ShippingAddressLine1,
		AddressLine2: order.ShippingAddressLine2,
		City:         order.ShippingCity,
		State:        order.ShippingState,
		PostalCode:   order.ShippingPostalCode,
		Country:      order.ShippingCountry,
		Email:        order.ShippingEmail,
		Phone:        order.ShippingPhone,
	}
}

func (o *OrdersService) buildOrderSummary(order models.Order, orderItems []models.OrderItem, productMap map[string]models.Product) (OrderSummary, error) {
	var orderItemsSummary []OrderItemSummary

	for _, orderItem := range orderItems {
		product, exists := productMap[orderItem.ProductID]
		if !exists {
			return OrderSummary{}, fmt.Errorf("product not found: %s", orderItem.ProductID)
		}

		orderItemSummary := o.buildOrderItemSummary(orderItem, product)
		orderItemsSummary = append(orderItemsSummary, orderItemSummary)
	}

	return OrderSummary{
		ID:              order.ID,
		UserID:          order.UserID,
		OrderItems:      orderItemsSummary,
		Total:           order.TotalAmount,
		Status:          order.Status,
		ShippingName:    order.ShippingName,
		ShippingAddress: o.buildShippingAddress(order),
		CreatedDate:     order.CreatedAt,
	}, nil
}

func (o *OrdersService) CreateOrderFromCart(ctx context.Context, userID string, shippingInfo ShippingInfo) (OrderSummary, error) {
	cart, err := o.cartsStore.GetOrCreateCartByUserID(ctx, userID) //Get cart
	if err != nil {
		return OrderSummary{}, fmt.Errorf("Error fetching cart from user %s: %w", userID, err)
	}
	cartItems, err := o.cartsStore.GetCartItems(ctx, cart.ID) // Get cart items
	if err != nil {
		return OrderSummary{}, fmt.Errorf("Error fetching cart items %w", err)
	}
	if len(cartItems) == 0 { // check if empty
		return OrderSummary{}, &customerrors.ErrCartEmpty
	}
	orderID := uuid.NewString()
	var orderItemsSummary []OrderItemSummary //create empty order items slice
	var orderItems []models.OrderItem

	subtotal := 0.0

	productIDs, err := utils.ExtractProductIDs(cartItems)
	if err != nil {
		return OrderSummary{}, fmt.Errorf("Failed to extract product ids: %w", err)
	}
	productMap, err := o.productsStore.GetByIDBatch(ctx, productIDs)
	if err != nil {
		return OrderSummary{}, fmt.Errorf("Failed to retrieve products by ids: %w", err)
	}

	for _, item := range cartItems { //  traverse cart items
		product, exists := productMap[item.ProductID] // get each product from cart items
		if !exists {
			return OrderSummary{}, fmt.Errorf("Error fetching products: %w", err)
		}

		if product.StockQuantity < item.Quantity { //check stock
			return OrderSummary{}, customerrors.NewErrInsufficientStock(item.ProductID, item.Quantity, product.StockQuantity, item.Quantity)
		}
		subtotal += float64(item.Quantity) * product.Price //subtotal
		orderItemID := uuid.NewString()
		orderItemSummary := OrderItemSummary{ // create the order item
			ID:          orderItemID,
			ProductID:   item.ProductID,
			ProductName: product.Name,
			Price:       product.Price,
			Quantity:    item.Quantity,
			Subtotal:    product.Price * float64(item.Quantity),
		}

		orderItem := models.OrderItem{ // create the order item
			ID:        orderItemID,
			OrderID:   orderID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     product.Price,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		orderItemsSummary = append(orderItemsSummary, orderItemSummary)
		orderItems = append(orderItems, orderItem) // append to order items
	}
	orderStatus := "pending"

	shippingAdress := ShippingAddress{
		AddressLine1: shippingInfo.AddressLine1,
		AddressLine2: shippingInfo.AddressLine2,
		City:         shippingInfo.City,
		State:        shippingInfo.State,
		PostalCode:   shippingInfo.PostalCode,
		Country:      shippingInfo.Country,
		Email:        shippingInfo.Email,
		Phone:        shippingInfo.Phone,
	}
	orderSummary := OrderSummary{
		ID:              orderID,
		UserID:          userID,
		OrderItems:      orderItemsSummary,
		Total:           subtotal,
		Status:          orderStatus,
		ShippingName:    shippingInfo.Name,
		ShippingAddress: shippingAdress,
		CreatedDate:     time.Now(),
	}

	order := models.Order{
		ID:                   orderID,
		UserID:               userID,
		TotalAmount:          subtotal,
		Status:               orderStatus,
		ShippingName:         shippingInfo.Name,
		ShippingEmail:        shippingInfo.Email,
		ShippingPhone:        shippingInfo.Phone,
		ShippingAddressLine1: shippingInfo.AddressLine1,
		ShippingAddressLine2: shippingInfo.AddressLine2,
		ShippingCity:         shippingInfo.City,
		ShippingState:        shippingInfo.State,
		ShippingPostalCode:   shippingInfo.PostalCode,
		ShippingCountry:      shippingInfo.Country,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}
	// begin transaction
	err = o.TxManager.WithTransaction(ctx, func(tx pgx.Tx) error {
		err = o.orderStore.InsertOrderTx(ctx, order, tx)
		if err != nil {
			return fmt.Errorf("Error creating order: %w", err)
		}
		err = o.orderStore.InsertOrderItemBulkTx(ctx, orderItems, tx)
		if err != nil {
			return fmt.Errorf("Error inserting order items: %w", err)
		}
		err = o.cartsStore.EmptyCartTx(ctx, userID, tx)
		if err != nil {
			return fmt.Errorf("Error clearing cart: %w", err)
		}
		return nil
	})
	if err != nil {
		return OrderSummary{}, err
	}

	return orderSummary, nil
}

func (o *OrdersService) GetOrdersHistory(ctx context.Context, userID string) ([]OrderSummary, error) {
	orders, err := o.orderStore.GetUsersOrders(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("Failed to return orders :%w", err)
	}

	if len(orders) == 0 {
		return nil, nil
	}

	var orderSummaries []OrderSummary
	orderIDs, err := utils.ExtractOrderIDs(orders)
	if err != nil {
		return nil, fmt.Errorf("Error extracting order ids: %W", err)
	}
	ordersItemsMap, err := o.orderStore.GetOrderItemsBatch(ctx, orderIDs)
	if err != nil {
		return nil, fmt.Errorf("Error mapping orders: %W", err)
	}
	var allProductIDs []string
	for _, orderItems := range ordersItemsMap {
		productIDs, err := utils.ExtractProductIDsFromOrderItems(orderItems)
		if err != nil {
			return nil, fmt.Errorf("Failed to extract product IDs: %w", err)
		}
		allProductIDs = append(allProductIDs, productIDs...)
	}
	productMap, err := o.productsStore.GetByIDBatch(ctx, allProductIDs) //db query has to be out of loop
	if err != nil {
		return nil, fmt.Errorf("Error mapping products: %w", err)
	}

	for _, order := range orders { // no db query in loop
		orderItems := ordersItemsMap[order.ID]
		orderSummary, err := o.buildOrderSummary(order, orderItems, productMap)
		if err != nil {
			return nil, fmt.Errorf("Failed to build order summary: %w", err)
		}
		orderSummaries = append(orderSummaries, orderSummary)
	}
	return orderSummaries, nil
}

func (o *OrdersService) CanBeCancelled(order models.Order) bool {
	return order.Status == "pending" || order.Status == "paid" || order.Status == "cancelled"
}

func (o *OrdersService) CancelOrder(ctx context.Context, userID, orderID string) error {

	order, err := o.orderStore.GetOrderByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("Error fetching order: %w", err)
	}

	if order.UserID != userID {
		return &customerrors.ErrOrderNotOwned
	}

	if !o.CanBeCancelled(order) {
		return &customerrors.ErrCannotCancel
	}
	if err := o.orderStore.UpdateOrderStatus(ctx, "cancelled", order.ID); err != nil {
		return fmt.Errorf("Error updating order status: %w", err)

	}
	return nil
}

func (o *OrdersService) GetOrderInfo(ctx context.Context, userID, orderID string) (OrderSummary, error) {

	order, err := o.orderStore.GetOrderByID(ctx, orderID)
	if err != nil {
		return OrderSummary{}, fmt.Errorf("Error fetching orders: %w", err)
	}
	if userID != order.UserID {
		return OrderSummary{}, &customerrors.ErrOrderNotOwned
	}

	orderItems, err := o.orderStore.GetOrderItems(ctx, order.ID)
	if err != nil {
		return OrderSummary{}, fmt.Errorf("Error fetching order items: %w", err)
	}
	productIDs, err := utils.ExtractProductIDsFromOrderItems(orderItems)
	if err != nil {
		return OrderSummary{}, fmt.Errorf("Error extracting product ids: %w", err)
	}

	productMap, err := o.productsStore.GetByIDBatch(ctx, productIDs)
	if err != nil {
		return OrderSummary{}, fmt.Errorf("Error mapping products to map: %w", err)
	}

	orderSummary, err := o.buildOrderSummary(order, orderItems, productMap)

	return orderSummary, nil
}

// func (o *OrdersService) UpdateShippingInfo(ctx context.Context, orderID string) (OrderOperationResult, error)

func (o *OrdersService) GetAllOrders(ctx context.Context) ([]OrderSummary, error) {
	orders, err := o.orderStore.GetAllOrders(ctx)
	if err != nil {
		return nil, fmt.Errorf("Error retrieving orders: %w", err)
	}

	orderIDs, err := utils.ExtractOrderIDs(orders)
	if err != nil {
		return nil, fmt.Errorf("Error extracting orders ids: %w", err)
	}

	ordersItemsMap, err := o.orderStore.GetOrderItemsBatch(ctx, orderIDs)

	var orderSummaries []OrderSummary

	var allProductIDs []string
	for _, orderItems := range ordersItemsMap {
		productIDs, err := utils.ExtractProductIDsFromOrderItems(orderItems)
		if err != nil {
			return nil, fmt.Errorf("Error extracting product ids: %w", err)
		}
		allProductIDs = append(allProductIDs, productIDs...)
	}
	productMap, err := o.productsStore.GetByIDBatch(ctx, allProductIDs)
	if err != nil {
		return nil, fmt.Errorf("Error getting products: %w", err)
	}

	for _, order := range orders {
		orderItems, exists := ordersItemsMap[order.ID]
		if !exists {
			return nil, fmt.Errorf("Order item with id %s does not exist", order.ID)
		}

		orderSummary, err := o.buildOrderSummary(order, orderItems, productMap)
		if err != nil {
			return nil, fmt.Errorf("Failed to build order summary: %w", err)
		}
		orderSummaries = append(orderSummaries, orderSummary)
	}
	return orderSummaries, nil
}

func (o *OrdersService) GetOrdersByStatus(ctx context.Context, status string) ([]OrderSummary, error) {
	orders, err := o.orderStore.GetOrdersByStatus(ctx, status)
	if err != nil {
		return nil, fmt.Errorf("Error retrieving orders: %w", err)
	}

	orderIDs, err := utils.ExtractOrderIDs(orders)
	if err != nil {
		return nil, fmt.Errorf("Error extracting orders ids: %w", err)
	}

	ordersItemsMap, err := o.orderStore.GetOrderItemsBatch(ctx, orderIDs)

	var orderSummaries []OrderSummary
	var allProductIDs []string

	for _, orderItems := range ordersItemsMap {
		productIDs, err := utils.ExtractProductIDsFromOrderItems(orderItems)
		if err != nil {
			return nil, fmt.Errorf("Error extracting product ids: %w", err)
		}
		allProductIDs = append(allProductIDs, productIDs...)
	}
	productMap, err := o.productsStore.GetByIDBatch(ctx, allProductIDs)
	if err != nil {
		return nil, fmt.Errorf("Error getting products: %w", err)
	}

	for _, order := range orders {
		orderItems, exists := ordersItemsMap[order.ID]
		if !exists {
			return nil, fmt.Errorf("Order item with id %s does not exist", order.ID)
		}

		orderSummary, err := o.buildOrderSummary(order, orderItems, productMap)
		if err != nil {
			return nil, fmt.Errorf("Failed to build order summary: %w", err)
		}
		orderSummaries = append(orderSummaries, orderSummary)
	}
	return orderSummaries, nil
}

func (o *OrdersService) UpdateOrderStatus(ctx context.Context, status, orderID string) error {
	if orderID == "" {
		return &customerrors.ErrOrderNotFound
	}
	order, err := o.orderStore.GetOrderByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("Error fetching order: %w", err)
	}
	if order.Status == "delivered" || order.Status == "cancelled" {
		return &customerrors.ErrCannotChangeStatus
	}

	if err := o.orderStore.UpdateOrderStatus(ctx, status, order.ID); err != nil {
		return fmt.Errorf("Failed to cancel order: %w", err)

	}
	return nil
}

func (o *OrdersService) GetOrderByIDAdmin(ctx context.Context, orderID string) (OrderSummary, error) {

	order, err := o.orderStore.GetOrderByID(ctx, orderID)
	if err != nil {
		return OrderSummary{}, fmt.Errorf("Error fetching orders: %w", err)
	}

	orderItems, err := o.orderStore.GetOrderItems(ctx, order.ID)
	if err != nil {
		return OrderSummary{}, fmt.Errorf("Error fetching order items: %w", err)
	}
	productIDs, err := utils.ExtractProductIDsFromOrderItems(orderItems)
	if err != nil {
		return OrderSummary{}, fmt.Errorf("Error extracting product ids: %w", err)
	}

	productMap, err := o.productsStore.GetByIDBatch(ctx, productIDs)
	if err != nil {
		return OrderSummary{}, fmt.Errorf("Error mapping products to map: %w", err)
	}

	orderSummary, err := o.buildOrderSummary(order, orderItems, productMap)

	return orderSummary, nil
}
