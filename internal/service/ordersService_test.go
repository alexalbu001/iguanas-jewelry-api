package service_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/alexalbu001/iguanas-jewelry-api/internal/service"
	"github.com/alexalbu001/iguanas-jewelry-api/internal/utils"
)

// MockEmailService is a test double that implements EmailService
type MockEmailService struct{}

func (m *MockEmailService) SendOrderConfirmation(ctx context.Context, orderSummary service.OrderSummary) error {
	return nil // No-op for tests
}

func (m *MockEmailService) SendWelcome(ctx context.Context, userName, userEmail string) error {
	return nil // No-op for tests
}

func (m *MockEmailService) SendAdminOrderNotification(ctx context.Context, orderSummary service.OrderSummary, adminEmail string) error {
	return nil // No-op for tests
}

func setupFreshOrderService() *service.OrdersService {
	return service.NewOrderService(
		&utils.MockOrderStore{
			OrderStore:     utils.CreateTestOrders(),
			OrderItemStore: utils.CreateTestOrderItems(),
		},
		&utils.MockProductStore{
			Store: utils.CreateJewelryProducts(),
		},
		&utils.MockCartsStore{
			CartsStore:     utils.CreateTestCarts(),
			CartItemsStore: utils.CreateTestCartItems(), // Fresh data each time
		},
		&MockEmailService{},
		"test-admin@example.com",
		&utils.MockTxManager{},
	)
}

var KnownShippingInfo = service.ShippingInfo{
	Name:         "Michael Chen",
	Email:        "michael.chen@example.com",
	Phone:        "+1-555-0456",
	AddressLine1: "456 Diamond Street",
	AddressLine2: "",
	City:         "Los Angeles",
	State:        "CA",
	PostalCode:   "90210",
	Country:      "USA",
}

var BsShippingInfo = service.ShippingInfo{
	Name:         "",
	Email:        "michael.chen@.com",
	Phone:        "+1-555-",
	AddressLine1: "456 mond Street",
	AddressLine2: "",
	City:         "Angeles",
	State:        "CA",
	PostalCode:   "90210",
	Country:      "SA",
}

func TestCreateOrderFromCart(t *testing.T) {
	orderService := setupFreshOrderService()

	orderSummary, err := orderService.CreateOrderFromCart(context.Background(), utils.KnownUserID, KnownShippingInfo)
	if err != nil {
		t.Fatalf("Failed to create order from cart: %v", err)
	}

	fmt.Printf("Available order items IDs:\n")
	for i, orderItem := range orderSummary.OrderItems {
		fmt.Printf("[%d] ID: OrderItemID: %s, product name: %s \n", i, orderItem.ID, orderItem.ProductName)
	}

	if len(orderSummary.OrderItems) != 3 {
		t.Errorf("Expected 3 order items, got %d instead", len(orderSummary.OrderItems))
	}

	if orderSummary.Total != 3749.48 {
		t.Errorf("Expected total: 3749.48, got %f instead", orderSummary.Total)
	}

	if orderSummary.ShippingName != "Michael Chen" {
		t.Errorf("Expected name: 'Michael Chen', got %s instead", orderSummary.ShippingName)
	}

	orderSummary2, err := orderService.CreateOrderFromCart(context.Background(), utils.AdminUserID, KnownShippingInfo)
	if err == nil {
		t.Errorf("This should error because user has no items in cart")
	}

	if len(orderSummary2.OrderItems) != 0 {
		t.Errorf("Expected 0 order items, got %d instead", len(orderSummary2.OrderItems))
	}

	// Should also test BS shipping address but dont have validation ATM
}

func TestGetOrdersHistory(t *testing.T) {
	orderService := setupFreshOrderService()

	orderSummaries, err := orderService.GetOrdersHistory(context.Background(), utils.KnownUserID)
	if err != nil {
		t.Fatalf("Failed to create order from cart: %v", err)
	}
	if len(orderSummaries) != 2 {
		t.Errorf("Expected 2 orders, got %d instead", len(orderSummaries))
	}

	if orderSummaries[0].Total != 3749.48 {
		t.Errorf("Expected total: 3749.48, got %f instead", orderSummaries[0].Total)
	}
	for i, order := range orderSummaries {
		for j, orderItem := range order.OrderItems {
			fmt.Printf("[%d] OrderID : %s, [%d] Order ItemID %s\n", i, order.ID, j, orderItem.ProductName)
		}
	}
	if orderSummaries[0].OrderItems[0].Quantity != 2 {
		t.Errorf("Expected quantity: 2, got %d instead", orderSummaries[0].OrderItems[0].Quantity)
	}
}

func TestCancelOrder(t *testing.T) {
	orderservice := setupFreshOrderService()

	err := orderservice.CancelOrder(context.Background(), utils.KnownUserID, utils.KnownOrderID)
	if err != nil {
		t.Fatalf("Failed to cancel order %v", err)
	}

	err = orderservice.CancelOrder(context.Background(), utils.KnownUserID, utils.SecondOrderID)
	if err == nil {
		t.Errorf("Cancel operation should fail as order status is 'paid")
	}

	err = orderservice.CancelOrder(context.Background(), utils.KnownUserID, utils.AdminOrderID)
	if err == nil {
		t.Errorf("Cancel operation should fail as order is not owned by this user")
	}

}

// Rest of order methods
