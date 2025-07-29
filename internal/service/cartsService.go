package service

import (
	"fmt"
	"log"

	"github.com/alexalbu001/iguanas-jewelry/internal/models"
	"github.com/google/uuid"
)

type CartsStore interface {
	GetOrCreateCartByUserID(id string) (models.Cart, error)
	EmptyCart(userID string) error
	GetCartItemByID(id string) (models.CartItems, error)
	AddItemToCart(cartID, productID string, quantity int) (models.CartItems, error)
	GetCartItems(cartID string) ([]models.CartItems, error)
	UpdateCartItemQuantity(itemID string, newQuantity int) error
	DeleteCartItem(cartItemID string) error
}

type CartsService struct {
	CartsStore    CartsStore
	ProductsStore ProductsStore
}

type CartOperationResult struct {
	Success     bool
	CartSummary CartSummary // Always populated (current cart state)
	Error       string      // Only populated on failure
}

type CartSummary struct {
	CartID  string
	Items   []CartItemSummary
	Total   float64
	Warning []string
}

type CartItemSummary struct {
	ID          string
	ProductID   string
	ProductName string
	Price       float64
	Quantity    int
	Subtotal    float64
}

func NewCartsService(cartsStore CartsStore, productsStore ProductsStore) *CartsService {
	return &CartsService{
		CartsStore:    cartsStore,
		ProductsStore: productsStore,
	}
}

func (c *CartsService) GetUserCart(userID string) (CartSummary, error) {
	err := uuid.Validate(userID)
	if err != nil {
		return CartSummary{}, fmt.Errorf("Invalid user ID")
	}

	cart, err := c.CartsStore.GetOrCreateCartByUserID(userID)
	if err != nil {
		return CartSummary{}, fmt.Errorf("Error retrieving or creating cart: %w", err)
	}
	if cart.ID == "" {
		return CartSummary{}, fmt.Errorf("Cart ID is empty")
	}
	cartItems, err := c.CartsStore.GetCartItems(cart.ID)
	if err != nil {
		return CartSummary{}, fmt.Errorf("Failed to retrieve cart items: %w", err)
	}

	var cartSummary CartSummary
	cartSummary.Items = []CartItemSummary{}
	var warnings []string

	for _, cartItem := range cartItems {
		cartProducts, err := c.ProductsStore.GetByID(cartItem.ProductID)
		if err != nil {
			deleteErr := c.CartsStore.DeleteCartItem(cartItem.ID)
			if deleteErr != nil {
				// Log it but don't fail the whole request
				log.Printf("Failed to delete invalid cart item %s: %v", cartItem.ID, deleteErr)
			}
			warnings = append(warnings, fmt.Sprintf("Removed unavailable item from cart"))
			continue
		}
		cartItemSummary := CartItemSummary{
			ID:          cartItem.ID,
			ProductID:   cartProducts.ID,
			ProductName: cartProducts.Name,
			Price:       cartProducts.Price,
			Quantity:    cartItem.Quantity,
			Subtotal:    cartProducts.Price * float64(cartItem.Quantity),
		}

		cartSummary.Items = append(cartSummary.Items, cartItemSummary)
	}
	for _, v := range cartSummary.Items {
		cartSummary.Total += v.Subtotal
	}
	cartSummary.Warning = warnings
	return cartSummary, nil
}

func (c *CartsService) AddToCart(userID, productID string, quantity int) (CartOperationResult, error) {

	product, err := c.ProductsStore.GetByID(productID)
	if err != nil {
		return CartOperationResult{}, fmt.Errorf("Error retrieving product with id : %s, : %w", productID, err)
	}
	if quantity > product.StockQuantity {
		currentCart, err := c.GetUserCart(userID)
		if err != nil {
			return CartOperationResult{}, err // System error getting cart
		}
		return CartOperationResult{
			Success:     false,
			CartSummary: currentCart,
			Error:       fmt.Sprintf("Only %d items available, requested %d", product.StockQuantity, quantity),
		}, nil

	}
	cart, err := c.CartsStore.GetOrCreateCartByUserID(userID)
	if err != nil {
		return CartOperationResult{}, err // System error getting cart
	}

	_, err = c.CartsStore.AddItemToCart(cart.ID, productID, quantity)
	if err != nil {
		return CartOperationResult{}, fmt.Errorf("Error adding product: %s to cart: %w", productID, err)
	}

	updatedCartSummary, err := c.GetUserCart(userID)
	if err != nil {
		return CartOperationResult{}, err
	}

	return CartOperationResult{
		Success:     true,
		CartSummary: updatedCartSummary,
		Error:       "",
	}, nil
}

func (c *CartsService) UpdateCartItemQuantity(userID, itemID string, quantity int) (CartOperationResult, error) {
	cartItem, err := c.CartsStore.GetCartItemByID(itemID)
	if err != nil {
		return CartOperationResult{}, fmt.Errorf("Failed to retrieve cart item: %s , %w", itemID, err)
	}

	product, err := c.ProductsStore.GetByID(cartItem.ProductID)
	if err != nil {
		return CartOperationResult{}, fmt.Errorf("Failed to retrieve product: %s , %w", cartItem.ProductID, err)
	}

	if quantity > product.StockQuantity {
		cartSummary, err := c.GetUserCart(userID)
		if err != nil {
			return CartOperationResult{}, fmt.Errorf("Failed to retrieve user cart: %s , %w", userID, err)
		}
		return CartOperationResult{
			Success:     false,
			CartSummary: cartSummary,
			Error:       fmt.Sprintf("Only %d items available, requested %d", product.StockQuantity, quantity),
		}, nil
	}

	if quantity == 0 {
		err = c.CartsStore.DeleteCartItem(itemID)
		if err != nil {
			return CartOperationResult{}, fmt.Errorf("Failed to delete cart item: %s , %w", itemID, err)
		}

		cartSummary, err := c.GetUserCart(userID)
		if err != nil {
			return CartOperationResult{}, fmt.Errorf("Failed to retrieve user cart: %s , %w", userID, err)
		}
		return CartOperationResult{
			Success:     true,
			CartSummary: cartSummary,
			Error:       "",
		}, nil
	}

	err = c.CartsStore.UpdateCartItemQuantity(itemID, quantity)
	if err != nil {
		return CartOperationResult{}, fmt.Errorf("Failed to update item quantity: %w", err)
	}

	currentCartSummary, err := c.GetUserCart(userID)
	if err != nil {
		return CartOperationResult{}, err
	}

	return CartOperationResult{
		Success:     true,
		CartSummary: currentCartSummary,
		Error:       "",
	}, nil
}

func (c *CartsService) RemoveFromCart(userID, itemID string) (CartOperationResult, error) {
	cartItem, err := c.CartsStore.GetCartItemByID(itemID)
	if err != nil {
		return CartOperationResult{}, fmt.Errorf("Failed to retrieve cart item: %s , %w", itemID, err)
	}

	err = c.CartsStore.DeleteCartItem(cartItem.ID)
	if err != nil {
		return CartOperationResult{}, fmt.Errorf("Failed to delete cart item: %s , %w", itemID, err)
	}

	cartSummary, err := c.GetUserCart(userID)
	if err != nil {
		return CartOperationResult{}, fmt.Errorf("Failed to retrieve user cart: %s , %w", userID, err)
	}

	return CartOperationResult{
		Success:     true,
		CartSummary: cartSummary,
		Error:       "",
	}, nil
}

func (c *CartsService) ClearCart(userID string) (CartOperationResult, error) {
	err := c.CartsStore.EmptyCart(userID)
	if err != nil {
		return CartOperationResult{}, fmt.Errorf("Failed to clear cart: %w", err)
	}

	cartSummary, err := c.GetUserCart(userID)
	if err != nil {
		return CartOperationResult{}, fmt.Errorf("Failed to retrieve user cart: %s , %w", userID, err)
	}
	return CartOperationResult{
		Success:     true,
		CartSummary: cartSummary,
		Error:       "",
	}, nil
}
