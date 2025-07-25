package service

import (
	"fmt"
	"time"

	"github.com/alexalbu001/iguanas-jewelry/internal/handlers"
	"github.com/alexalbu001/iguanas-jewelry/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx"
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
	InsertOrder(order models.Order) error
	InsertOrderTx(order models.Order, tx pgx.Tx) error
	InsertOrderItem(orderItem models.OrderItem) error
	InsertOrderItemTx(orderItem models.OrderItem, tx pgx.Tx) error
	InsertOrderItemBulk(orderItems []models.OrderItem) error
	InsertOrderItemBulkTx(orderItems []models.OrderItem, tx pgx.Tx) error
	GetOrderByID(orderID string) (models.Order, error)
	GetOrderItems(orderID string) ([]models.OrderItem, error)
	GetUsersOrders(userID string) ([]models.Order, error)
	UpdateOrderStatus(status, orderID string) error
}

type OrdersService struct {
	orderStore    OrdersStore
	productsStore handlers.ProductsStore
	cartsStore    handlers.CartsStore
}

func NewOrderService(orderStore OrdersStore, productStore handlers.ProductsStore, cartsStore handlers.CartsStore) OrdersService {
	return OrdersService{
		orderStore:    orderStore,
		productsStore: productStore,
		cartsStore:    cartsStore,
	}
}

func (o *OrdersService) CreateOrderFromCart(userID string, shippingInfo ShippingInfo) (models.Order, error) {
	cart, err := o.cartsStore.GetOrCreateCartByUserID(userID) //Get cart
	if err != nil {
		return models.Order{}, fmt.Errorf("Error fetching cart from user %s: %w", userID, err)
	}
	cartItems, err := o.cartsStore.GetCartItems(cart.ID) // Get cart items
	if err != nil {
		return models.Order{}, fmt.Errorf("Error fetching cart items %w", err)
	}
	if len(cartItems) == 0 { // check if empty
		return models.Order{}, fmt.Errorf("cart is empty")
	}
	orderID := uuid.NewString()
	var orderItems []models.OrderItem //create empty order items slice
	subtotal := 0.0
	for _, item := range cartItems { //  traverse cart items
		product, err := o.productsStore.GetByID(item.ProductID) // get each product from cart items
		if err != nil {
			return models.Order{}, fmt.Errorf("Error fetching products: %w", err)
		}
		if product.StockQuantity < item.Quantity { //check stock
			return models.Order{}, fmt.Errorf("Not enough products in stock")
		}
		subtotal += float64(item.Quantity) * product.Price //subtotal
		orderItemID := uuid.NewString()
		orderItem := models.OrderItem{ // create the order item
			ID:        orderItemID,
			OrderID:   orderID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     product.Price,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		orderItems = append(orderItems, orderItem) // append to order items
	}
	orderStatus := "Pending"
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
	// err = store.BeginTransaction(o.orderStore, o.cartsStore)

	err = o.orderStore.InsertOrder(order)
	if err != nil {
		return models.Order{}, fmt.Errorf("Error creating order: %w", err)
	}
	err = o.orderStore.InsertOrderItemBulk(orderItems)
	if err != nil {
		return models.Order{}, fmt.Errorf("Error inserting order items: %w", err)
	}
	err = o.cartsStore.EmptyCart(userID)
	if err != nil {
		return models.Order{}, fmt.Errorf("Error clearing cart: %w", err)
	}
	return order, nil
}
