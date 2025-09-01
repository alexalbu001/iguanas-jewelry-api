package service

import (
	"context"
	"fmt"
	"log"

	customerrors "github.com/alexalbu001/iguanas-jewelry-api/internal/customErrors"
	"github.com/alexalbu001/iguanas-jewelry-api/internal/models"
	"github.com/alexalbu001/iguanas-jewelry-api/internal/transaction"
	"github.com/alexalbu001/iguanas-jewelry-api/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type CartsStore interface {
	GetOrCreateCartByUserID(ctx context.Context, userID string) (models.Cart, error)
	EmptyCart(ctx context.Context, userID string) error
	EmptyCartTx(ctx context.Context, userID string, tx pgx.Tx) error
	GetCartItemByID(ctx context.Context, id string) (models.CartItems, error)
	AddItemToCart(ctx context.Context, cartID, productID string, quantity int) (models.CartItems, error)
	GetCartItems(ctx context.Context, cartID string) ([]models.CartItems, error)
	UpdateCartItemQuantity(ctx context.Context, itemID string, newQuantity int) error
	DeleteCartItem(ctx context.Context, cartItemID string) error
	GetCartItemByProductID(ctx context.Context, productID, cartID string) (models.CartItems, error)
}

type CartsService struct {
	CartsStore    CartsStore
	ProductsStore ProductsStore
	UsersStore    UsersStore
	TxManager     transaction.TxManager
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

func NewCartsService(cartsStore CartsStore, productsStore ProductsStore, usersStore UsersStore, TxManager transaction.TxManager) *CartsService {
	return &CartsService{
		CartsStore:    cartsStore,
		ProductsStore: productsStore,
		UsersStore:    usersStore,
		TxManager:     TxManager,
	}
}

func (c *CartsService) GetUserCart(ctx context.Context, userID string) (CartSummary, error) {
	err := uuid.Validate(userID)
	if err != nil {
		return CartSummary{}, &customerrors.ErrUserNotFound
	}
	c.UsersStore.GetUserByID(ctx, userID)
	if err != nil {
		return CartSummary{}, fmt.Errorf("Error fetching user : %w", err)
	}

	cart, err := c.CartsStore.GetOrCreateCartByUserID(ctx, userID)
	if err != nil {
		return CartSummary{}, fmt.Errorf("Error retrieving or creating cart: %w", err)
	}

	if cart.ID == "" {
		return CartSummary{}, &customerrors.ErrCartEmpty
	}
	cartItems, err := c.CartsStore.GetCartItems(ctx, cart.ID)
	if err != nil {
		return CartSummary{}, fmt.Errorf("Failed to retrieve cart items: %w", err)
	}

	if len(cartItems) == 0 {
		return CartSummary{
			CartID:  cart.ID,
			Items:   []CartItemSummary{},
			Total:   0.0,
			Warning: []string{},
		}, nil
	}
	var cartSummary CartSummary
	cartSummary.Items = []CartItemSummary{}
	var warnings []string

	productIDs, err := utils.ExtractProductIDs(cartItems)
	if err != nil {
		return CartSummary{}, fmt.Errorf("Failed to extract product IDs: %w", err)
	}
	productMap, err := c.ProductsStore.GetByIDBatch(ctx, productIDs)
	if err != nil {
		return CartSummary{}, fmt.Errorf("Failed to get products by ids: %w", err)
	}

	for _, cartItem := range cartItems {
		product, exists := productMap[cartItem.ProductID]
		if !exists {
			deleteErr := c.CartsStore.DeleteCartItem(ctx, cartItem.ID)
			if deleteErr != nil {
				// Log it but don't fail the whole request
				log.Printf("Failed to delete invalid cart item %s: %v", cartItem.ID, deleteErr)
			}
			warnings = append(warnings, fmt.Sprintf("Removed unavailable item from cart"))
			continue
		}
		cartItemSummary := CartItemSummary{
			ID:          cartItem.ID,
			ProductID:   product.ID,
			ProductName: product.Name,
			Price:       product.Price,
			Quantity:    cartItem.Quantity,
			Subtotal:    product.Price * float64(cartItem.Quantity),
		}

		cartSummary.Items = append(cartSummary.Items, cartItemSummary)
	}
	for _, v := range cartSummary.Items {
		cartSummary.Total += v.Subtotal
	}
	cartSummary.Warning = warnings
	cartSummary.CartID = cart.ID

	return cartSummary, nil
}

func (c *CartsService) AddToCart(ctx context.Context, userID, productID string, quantity int) (CartSummary, error) {

	cart, err := c.CartsStore.GetOrCreateCartByUserID(ctx, userID)
	if err != nil {
		return CartSummary{}, fmt.Errorf("Error getting or creating cart : %w", err) // System error getting cart
	}

	product, err := c.ProductsStore.GetByID(ctx, productID)
	if err != nil {
		return CartSummary{}, fmt.Errorf("Error retrieving product with id : %s, : %w", productID, err)
	}

	currentCartItemQuantity := c.GetCartItemQuantity(ctx, cart.ID, productID)

	if currentCartItemQuantity > product.StockQuantity {
		return CartSummary{}, customerrors.NewErrInsufficientStock(productID, quantity, product.StockQuantity, currentCartItemQuantity)
	}

	totalAfterAddition := currentCartItemQuantity + quantity
	if totalAfterAddition > product.StockQuantity {
		return CartSummary{}, customerrors.NewErrInsufficientStock(
			productID,
			quantity,
			product.StockQuantity,
			currentCartItemQuantity,
		)
	}

	_, err = c.CartsStore.AddItemToCart(ctx, cart.ID, productID, quantity) // dont reduce stock if added to cart
	if err != nil {
		return CartSummary{}, fmt.Errorf("Error adding product: %s to cart: %w", productID, err)
	}

	updatedCartSummary, err := c.GetUserCart(ctx, userID)
	if err != nil {
		return CartSummary{}, fmt.Errorf("Error retrieving cart : %w", err)
	}

	return updatedCartSummary, nil
}

func (c *CartsService) UpdateCartItemQuantity(ctx context.Context, userID, itemID string, quantity int) (CartSummary, error) {
	cart, err := c.CartsStore.GetOrCreateCartByUserID(ctx, userID)
	if err != nil {
		return CartSummary{}, fmt.Errorf("Failed to retrieve cart : %s , %w", userID, err)
	}

	cartItem, err := c.CartsStore.GetCartItemByID(ctx, itemID)
	if err != nil {
		return CartSummary{}, fmt.Errorf("Failed to retrieve cart item: %s , %w", itemID, err)
	}

	if cart.ID != cartItem.CartID {
		return CartSummary{}, &customerrors.ErrCartItemNotFound
	}

	product, err := c.ProductsStore.GetByID(ctx, cartItem.ProductID)
	if err != nil {
		return CartSummary{}, fmt.Errorf("Failed to retrieve product: %s , %w", cartItem.ProductID, err)
	}
	// var warnings []string
	if quantity > product.StockQuantity {
		// cartSummary, err := c.GetUserCart(ctx, userID)
		// if err != nil {
		// 	return CartSummary{}, fmt.Errorf("Failed to retrieve user cart: %s , %w", userID, err)
		// }
		// warnings = append(warnings, fmt.Sprintf("Only %d items available, requested %d", product.StockQuantity, quantity))
		return CartSummary{}, customerrors.NewErrInsufficientStock(product.ID, quantity, product.StockQuantity, cartItem.Quantity)
	}

	if quantity == 0 {
		err = c.CartsStore.DeleteCartItem(ctx, itemID)
		if err != nil {
			return CartSummary{}, fmt.Errorf("Failed to delete cart item: %s , %w", itemID, err)
		}

		cartSummary, err := c.GetUserCart(ctx, userID)
		if err != nil {
			return CartSummary{}, fmt.Errorf("Failed to retrieve user cart: %s , %w", userID, err)
		}
		return cartSummary, nil
	}

	err = c.CartsStore.UpdateCartItemQuantity(ctx, itemID, quantity)
	if err != nil {
		return CartSummary{}, fmt.Errorf("Failed to update item quantity: %w", err)
	}

	currentCartSummary, err := c.GetUserCart(ctx, userID)
	if err != nil {
		return CartSummary{}, fmt.Errorf("Failed to retrieve user cart: %s , %w", userID, err)
	}

	return currentCartSummary, nil
}

func (c *CartsService) RemoveFromCart(ctx context.Context, userID, itemID string) (CartSummary, error) {
	cart, err := c.CartsStore.GetOrCreateCartByUserID(ctx, userID)
	if err != nil {
		return CartSummary{}, fmt.Errorf("Failed to retrieve cart : %s , %w", userID, err)
	}

	cartItem, err := c.CartsStore.GetCartItemByID(ctx, itemID)
	if err != nil {
		return CartSummary{}, fmt.Errorf("Failed to retrieve cart item: %s , %w", itemID, err)
	}

	if cart.ID != cartItem.CartID {
		return CartSummary{}, &customerrors.ErrCartItemNotFound
	}

	err = c.CartsStore.DeleteCartItem(ctx, cartItem.ID)
	if err != nil {
		return CartSummary{}, fmt.Errorf("Failed to delete cart item: %s , %w", itemID, err)
	}

	cartSummary, err := c.GetUserCart(ctx, userID)
	if err != nil {
		return CartSummary{}, fmt.Errorf("Failed to retrieve user cart: %s , %w", userID, err)
	}

	return cartSummary, nil
}

func (c *CartsService) ClearCart(ctx context.Context, userID string) (CartSummary, error) {
	err := c.CartsStore.EmptyCart(ctx, userID)
	if err != nil {
		return CartSummary{}, fmt.Errorf("Failed to clear cart: %w", err)
	}

	cartSummary, err := c.GetUserCart(ctx, userID)
	if err != nil {
		return CartSummary{}, fmt.Errorf("Failed to retrieve user cart: %s , %w", userID, err)
	}
	return cartSummary, nil
}

func (c *CartsService) GetCartItemQuantity(ctx context.Context, cartID, productID string) int {
	cartItem, err := c.CartsStore.GetCartItemByProductID(ctx, productID, cartID)
	if err != nil {
		return 0
	}
	return cartItem.Quantity
}
