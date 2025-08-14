package service_test

import (
	"context"
	"errors"
	"testing"

	customerrors "github.com/alexalbu001/iguanas-jewelry/internal/customErrors"
	"github.com/alexalbu001/iguanas-jewelry/internal/service"
	"github.com/alexalbu001/iguanas-jewelry/internal/utils"
)

func setupFreshCartsService() *service.CartsService {
	return service.NewCartsService(
		&utils.MockCartsStore{
			CartsStore:     utils.CreateTestCarts(),
			CartItemsStore: utils.CreateTestCartItems(), // Fresh data each time
		},
		&utils.MockProductStore{
			Store: utils.CreateJewelryProducts(),
		},
		&utils.MockUsersStore{
			Users: utils.CreateTestUsers(),
		},
		&utils.MockTxManager{},
	)
}

func TestGetUserCart(t *testing.T) {
	mockUsersStore := &utils.MockUsersStore{
		Users: utils.CreateTestUsers(),
	}

	mockCartStore := &utils.MockCartsStore{
		CartsStore:     utils.CreateTestCarts(),
		CartItemsStore: utils.CreateTestCartItems(),
	}

	mockProductStore := &utils.MockProductStore{
		Store: utils.CreateJewelryProducts(),
	}

	serviceMockCart := service.NewCartsService(mockCartStore, mockProductStore, mockUsersStore, &utils.MockTxManager{})

	//valid customer with items in cart

	cartSummary, err := serviceMockCart.GetUserCart(context.Background(), utils.KnownUserID)
	if err != nil {
		t.Fatalf("Error retrieving user cart: %v", err)
	}

	if len(cartSummary.Items) != 3 {
		t.Errorf("Expected 3 cart items, got %d instead", len(cartSummary.Items))
	}

	if cartSummary.Total != 3749.48 {
		t.Errorf("Expected total to be 3749.48 got %f instead", cartSummary.Total)
	}

	//valid customer with empty cart

	cartSummary2, err := serviceMockCart.GetUserCart(context.Background(), utils.AdminUserID)
	if err != nil {
		t.Fatalf("Error retrieving user cart: %v", err)
	}

	if len(cartSummary2.Items) != 0 {
		t.Errorf("Expected 0 cart items, got %d instead", len(cartSummary2.Items))
	}

	if cartSummary2.Total != 0 {
		t.Errorf("Expected total to be 0 got %f instead", cartSummary2.Total)
	}
}

func TestAddToCart(t *testing.T) {
	mockUsersStore := &utils.MockUsersStore{
		Users: utils.CreateTestUsers(),
	}

	mockCartStore := &utils.MockCartsStore{
		CartsStore:     utils.CreateTestCarts(),
		CartItemsStore: utils.CreateTestCartItems(),
	}

	mockProductStore := &utils.MockProductStore{
		Store: utils.CreateJewelryProducts(),
	}

	serviceMockCart := service.NewCartsService(mockCartStore, mockProductStore, mockUsersStore, &utils.MockTxManager{})

	// Adding a product thats already in the cart

	cartSummary, err := serviceMockCart.AddToCart(context.Background(), utils.KnownUserID, utils.GoldRingID, 1)
	if err != nil {
		t.Fatalf("Failed to add product to cart: %v", err)
	}

	if len(cartSummary.Items) != 3 {
		t.Errorf("Expected 3 cart items, got %d instead", len(cartSummary.Items))
	}

	if cartSummary.Total != 4649.47 {
		t.Errorf("Expected total to be 4649.47 got %f instead", cartSummary.Total)
	}

	if cartSummary.Items[2].Quantity != 3 {
		t.Errorf("Expected cart item quantity to be 3, got %d instead", cartSummary.Items[2].Quantity)
	}

	// Adding a product that wasn't in cart already

	cartSummary2, err := serviceMockCart.AddToCart(context.Background(), utils.KnownUserID, utils.SapphireRingID, 1)
	if err != nil {
		t.Fatalf("Failed to add product to cart: %v", err)
	}

	if len(cartSummary2.Items) != 4 {
		t.Errorf("Expected 4 cart items, got %d instead", len(cartSummary2.Items))
	}

	if cartSummary2.Total != 6799.47 {
		t.Errorf("Expected total to be 6799.47 got %f instead", cartSummary2.Total)
	}

	// Adding products to another cart untill stock is overflowed

	cartSummary3, err := serviceMockCart.AddToCart(context.Background(), utils.SecondUserID, utils.GoldRingID, 13)
	if err != nil {
		t.Fatalf("Failed to add product to cart: %v", err)
	}

	if cartSummary3.Items[2].Subtotal != 11699.87 {
		t.Errorf("Expected total to be 11699.87 got %f instead", cartSummary3.Items[2].Subtotal)
	}

	cartSummary4, err := serviceMockCart.AddToCart(context.Background(), utils.SecondUserID, utils.GoldRingID, 3)
	if err == nil {
		t.Errorf("Add to product should fail due to quantity %d > stock 15", cartSummary4.Items[2].Quantity)
	}
}

func TestUpdateCartItemQuantity(t *testing.T) {
	mockUsersStore := &utils.MockUsersStore{
		Users: utils.CreateTestUsers(),
	}

	mockCartStore := &utils.MockCartsStore{
		CartsStore:     utils.CreateTestCarts(),
		CartItemsStore: utils.CreateTestCartItems(),
	}

	mockProductStore := &utils.MockProductStore{
		Store: utils.CreateJewelryProducts(),
	}

	serviceMockCart := service.NewCartsService(mockCartStore, mockProductStore, mockUsersStore, &utils.MockTxManager{})

	cartSummaryGet, err := serviceMockCart.GetUserCart(context.Background(), utils.KnownUserID)
	if err != nil {
		t.Fatalf("Error retrieving user cart: %v", err)
	}
	if cartSummaryGet.Items[0].Quantity != 2 {
		t.Errorf("Expected 2 items, got %d instead", cartSummaryGet.Items[0].Quantity)
	}

	_, err = serviceMockCart.UpdateCartItemQuantity(context.Background(), utils.KnownUserID, utils.KnownCartItemID, 4)
	if err != nil {
		t.Fatalf("Error updating item quantity in user cart: %v", err)
	}

	cartSummaryGet2, err := serviceMockCart.GetUserCart(context.Background(), utils.KnownUserID)
	if err != nil {
		t.Fatalf("Error retrieving user cart: %v", err)
	}
	if cartSummaryGet2.Items[0].Quantity != 4 {
		t.Errorf("Expected 4 items, got %d instead", cartSummaryGet.Items[0].Quantity)
	}

	_, err = serviceMockCart.UpdateCartItemQuantity(context.Background(), utils.KnownUserID, utils.KnownCartItemID, 16)
	if err == nil {
		t.Error("Expected insufficient stock error, got nil")
	}

	// Check if it's the right type of error
	var insufficientStockErr *customerrors.ErrInsufficientStock
	if !errors.As(err, &insufficientStockErr) {
		t.Errorf("Expected InsufficientStockError, got %T", err)
	}

	_, err = serviceMockCart.UpdateCartItemQuantity(context.Background(), utils.KnownUserID, utils.GoldRingID, 16)
	if err == nil {
		t.Error("Expected insufficient wrong item id error, got nil")
	}
}

func TestRemoveFromCart(t *testing.T) {

	serviceMockCart := setupFreshCartsService()

	// cartSummary, _ := serviceMockCart.GetUserCart(context.Background(), utils.KnownUserID)
	// for i, item := range cartSummary.Items {
	// 	fmt.Printf("  [%d] ID: %s, ProductID: %s\n", i, item.ID, item.ProductID)
	// }

	// fmt.Printf("Trying to delete ID: %s\n", utils.KnownCartItemID)

	// User with no cart

	_, err := serviceMockCart.RemoveFromCart(context.Background(), utils.AdminUserID, utils.KnownCartItemID)
	if err == nil {
		t.Errorf("Expected no cart item detected, couldn't remove cart item")
	}

	// User with cart and cart item to delete

	cartSummaryDelete, err := serviceMockCart.RemoveFromCart(context.Background(), utils.KnownUserID, utils.KnownCartItemID)
	if err != nil {
		t.Fatalf("Error removing from cart: %v", err)
	}

	if len(cartSummaryDelete.Items) != 2 {
		t.Errorf("Expected 2 items got %d instead ", len(cartSummaryDelete.Items))
	}
}

func TestClearCart(t *testing.T) {
	serviceMockCart := setupFreshCartsService()

	// Clear cart from knownuser with items
	cartSummary, err := serviceMockCart.ClearCart(context.Background(), utils.KnownUserID)
	if err != nil {
		t.Fatalf("Error clearing cart: %v", err)
	}
	if len(cartSummary.Warning) != 0 {
		t.Errorf("Cart should've been empty")
	}

	_, err = serviceMockCart.ClearCart(context.Background(), utils.AdminUserID)
	if err != nil {
		t.Fatalf("Error clearing cart: %v", err)
	}
}
