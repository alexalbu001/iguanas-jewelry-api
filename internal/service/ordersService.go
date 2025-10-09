package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	customerrors "github.com/alexalbu001/iguanas-jewelry-api/internal/customErrors"
	"github.com/alexalbu001/iguanas-jewelry-api/internal/models"
	"github.com/alexalbu001/iguanas-jewelry-api/internal/transaction"
	"github.com/alexalbu001/iguanas-jewelry-api/internal/utils"
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
	UpdateOrderStatusTx(ctx context.Context, status, orderID string, tx pgx.Tx) error
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
	ID              string             `json:"id"`
	UserID          string             `json:"user_id"`
	OrderItems      []OrderItemSummary `json:"items"`
	Total           float64            `json:"total_amount"`
	Status          string             `json:"status"`
	ShippingName    string             `json:"shipping_name"`
	ShippingAddress ShippingAddress    `json:"shipping_address"`
	CreatedDate     time.Time          `json:"created_at"`
}
type ShippingAddress struct {
	AddressLine1 string `json:"address_line1"`
	AddressLine2 string `json:"address_line2"`
	City         string `json:"city"`
	State        string `json:"state"`
	PostalCode   string `json:"postal_code"`
	Country      string `json:"country"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
}

type OrderItemSummary struct {
	ID          string  `json:"id"`
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	Price       float64 `json:"price"`
	Quantity    int     `json:"quantity"`
	Subtotal    float64 `json:"subtotal"`
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

func (o *OrdersService) CreateOrderFromCart(ctx context.Context, userID string, rawInfo ShippingInfo) (OrderSummary, error) {

	shippingInfo, err := o.ValidateShippingInfo(ctx, rawInfo)
	if err != nil {
		return OrderSummary{}, &customerrors.ErrInvalidInput
	}

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
			err := o.cartsStore.DeleteCartItem(ctx, item.ID)
			if err != nil {
				return OrderSummary{}, fmt.Errorf("Failed to delete cart item: %w", err)
			}
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
		for _, item := range orderItems {
			err = o.productsStore.UpdateStockTx(ctx, item.ProductID, -item.Quantity, tx)
			if err != nil {
				return fmt.Errorf("Failed to update stock quantity: %w", err)
			}
		}

		// Note: Cart is NOT cleared here - it will be cleared only after successful payment
		return nil
	})
	if err != nil {
		return OrderSummary{}, err
	}

	// After transaction commits successfully, invalidate product cache
	// This ensures the product list API returns updated stock quantities
	o.invalidateProductCache(ctx, orderItems)

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
		return nil, fmt.Errorf("Error extracting order ids: %w", err)
	}
	ordersItemsMap, err := o.orderStore.GetOrderItemsBatch(ctx, orderIDs)
	if err != nil {
		return nil, fmt.Errorf("Error mapping orders: %w", err)
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
		orderItems, exists := ordersItemsMap[order.ID]
		if !exists || len(orderItems) == 0 {
			// Skip orders with no items - these are orphaned orders
			continue
		}

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

	orderItems, err := o.orderStore.GetOrderItems(ctx, order.ID)
	if err != nil {
		return fmt.Errorf("Failed to fetch order items: %w", err)
	}
	err = o.TxManager.WithTransaction(ctx, func(tx pgx.Tx) error {
		if err := o.orderStore.UpdateOrderStatusTx(ctx, "cancelled", order.ID, tx); err != nil {
			return fmt.Errorf("Error updating order status: %w", err)
		}
		for _, item := range orderItems {
			err = o.productsStore.UpdateStockTx(ctx, item.ProductID, item.Quantity, tx)
			if err != nil {
				return fmt.Errorf("Failed to update product stock: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("Transaction failed :%w", err)
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

	// Return empty slice if no orders exist
	if len(orders) == 0 {
		return []OrderSummary{}, nil
	}

	orderIDs, err := utils.ExtractOrderIDs(orders)
	if err != nil {
		return nil, fmt.Errorf("Error extracting orders ids: %w", err)
	}

	ordersItemsMap, err := o.orderStore.GetOrderItemsBatch(ctx, orderIDs)
	if err != nil {
		return nil, fmt.Errorf("Error retrieving order items: %w", err)
	}

	// Initialize with capacity to avoid nil slice
	orderSummaries := make([]OrderSummary, 0, len(orders))

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
		if !exists || len(orderItems) == 0 {
			// Skip orders with no items - these are orphaned orders
			continue
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
		if !exists || len(orderItems) == 0 {
			// Skip orders with no items - these are orphaned orders
			continue
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

	orderItems, err := o.orderStore.GetOrderItems(ctx, order.ID)
	if err != nil {
		return fmt.Errorf("Failed to fetch order items: %w", err)
	}

	err = o.TxManager.WithTransaction(ctx, func(tx pgx.Tx) error {
		if err := o.orderStore.UpdateOrderStatusTx(ctx, status, order.ID, tx); err != nil {
			return fmt.Errorf("Failed to cancel order: %w", err)
		}
		for _, item := range orderItems {
			err = o.productsStore.UpdateStockTx(ctx, item.ProductID, item.Quantity, tx)
			if err != nil {
				return fmt.Errorf("Failed to update product stock: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("Transaction failed :%w", err)
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

func (o *OrdersService) GetOrderStatus(ctx context.Context, orderID string) (string, error) {
	err := uuid.Validate(orderID)
	if err != nil {
		return "", &customerrors.ErrInvalidInput
	}

	order, err := o.orderStore.GetOrderByID(ctx, orderID)
	if err != nil {
		return "", err
	}

	return order.Status, nil
}

func (o *OrdersService) ValidateShippingInfo(ctx context.Context, shippinginfo ShippingInfo) (ShippingInfo, error) {

	info := o.normalizeShippingInfo(shippinginfo)
	if info.Name == "" || info.Email == "" || info.Phone == "" || info.AddressLine1 == "" {
		return ShippingInfo{}, &customerrors.ErrFieldsMissing
	}
	if len(info.Name) > 255 {
		return ShippingInfo{}, &customerrors.ErrShippingNameTooLong
	}
	if len(info.AddressLine1) > 255 {
		return ShippingInfo{}, &customerrors.ErrShippingAddressTooLong
	}
	if !strings.Contains(info.Email, "@") || len(info.Email) < 3 {
		return ShippingInfo{}, &customerrors.ErrInvalidEmail
	}
	return info, nil
}

func (o *OrdersService) normalizeShippingInfo(info ShippingInfo) ShippingInfo {
	// 1. Trim whitespace from all fields
	info.Name = strings.TrimSpace(info.Name)
	info.Email = strings.TrimSpace(info.Email)
	info.Phone = strings.TrimSpace(info.Phone)
	info.AddressLine1 = strings.TrimSpace(info.AddressLine1)
	info.AddressLine2 = strings.TrimSpace(info.AddressLine2)
	info.City = strings.TrimSpace(info.City)
	info.State = strings.TrimSpace(info.State)
	info.PostalCode = strings.TrimSpace(info.PostalCode)
	info.Country = strings.TrimSpace(info.Country)

	// 2. Collapse multiple spaces into one
	space := regexp.MustCompile(`\s+`)
	info.AddressLine1 = space.ReplaceAllString(info.AddressLine1, " ")
	info.AddressLine2 = space.ReplaceAllString(info.AddressLine2, " ")
	info.City = space.ReplaceAllString(info.City, " ")

	// 3. Standardize email
	info.Email = strings.ToLower(info.Email)

	// 4. Standardize country code
	info.Country = strings.ToUpper(info.Country)

	// 5. Clean phone (keep only digits)
	re := regexp.MustCompile(`[^0-9+]`) // Keep + for country code
	info.Phone = re.ReplaceAllString(info.Phone, "")

	// 6. Standardize postal code by country
	if info.Country == "GB" || info.Country == "UK" {
		info.PostalCode = strings.ToUpper(info.PostalCode)
		// Remove spaces from UK postcodes for consistent storage
		info.PostalCode = strings.ReplaceAll(info.PostalCode, " ", "")
	}
	return info
}

// ClearCartAfterPayment clears the user's cart after successful payment
func (o *OrdersService) ClearCartAfterPayment(ctx context.Context, userID string) error {
	err := o.cartsStore.EmptyCart(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to clear cart after payment: %w", err)
	}
	return nil
}

// invalidateProductCache invalidates the cache for products in order items
func (o *OrdersService) invalidateProductCache(ctx context.Context, orderItems []models.OrderItem) {
	if len(orderItems) == 0 {
		return
	}

	productIDs := make([]string, 0, len(orderItems))
	for _, item := range orderItems {
		productIDs = append(productIDs, item.ProductID)
	}

	// Call cache invalidation (ignore errors as it's not critical)
	o.productsStore.InvalidateProductCache(ctx, productIDs)
}

// CancelOrderAndRestoreStock cancels an order and restores the stock
func (o *OrdersService) CancelOrderAndRestoreStock(ctx context.Context, orderID string) error {

	// Get order items
	orderItems, err := o.orderStore.GetOrderItems(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get order items: %w", err)
	}

	// Update order status and restore stock in a transaction
	err = o.TxManager.WithTransaction(ctx, func(tx pgx.Tx) error {
		// Update order status to cancelled
		err = o.orderStore.UpdateOrderStatusTx(ctx, "cancelled", orderID, tx)
		if err != nil {
			return fmt.Errorf("failed to update order status: %w", err)
		}

		// Restore stock for each item
		for _, item := range orderItems {
			err = o.productsStore.UpdateStockTx(ctx, item.ProductID, item.Quantity, tx)
			if err != nil {
				return fmt.Errorf("failed to restore stock for product %s: %w", item.ProductID, err)
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to cancel order and restore stock: %w", err)
	}

	// After transaction commits successfully, invalidate product cache
	// This ensures the product list API returns updated stock quantities
	o.invalidateProductCache(ctx, orderItems)

	return nil
}

// func (o *OrdersService) ScheduleOrderConfirmationEmail(ctx context.Context, orderSummary OrderSummary)error{

// }
