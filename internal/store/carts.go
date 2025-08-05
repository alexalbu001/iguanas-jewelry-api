package store

import (
	"context"
	"fmt"
	"time"

	customerrors "github.com/alexalbu001/iguanas-jewelry/internal/customErrors"
	"github.com/alexalbu001/iguanas-jewelry/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CartsStore struct {
	dbpool *pgxpool.Pool
}

func NewCartsStore(connection *pgxpool.Pool) *CartsStore {
	return &CartsStore{
		dbpool: connection,
	}
}

func (c *CartsStore) GetOrCreateCartByUserID(ctx context.Context, id string) (models.Cart, error) {
	sql := `
 SELECT id, user_id, created_at, updated_at
 FROM carts
 WHERE user_id=$1
 `

	row := c.dbpool.QueryRow(ctx, sql, id)

	var cart models.Cart

	err := row.Scan(
		&cart.ID,
		&cart.UserID,
		&cart.CreatedAt,
		&cart.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			sql = `
INSERT INTO carts (id, user_id, created_at, updated_at)
VALUES ($1, $2, $3, $4)
`
			cartID := uuid.NewString()
			now := time.Now()
			_, err = c.dbpool.Exec(ctx, sql, cartID, id, now, now)
			if err != nil {
				return models.Cart{}, fmt.Errorf("Cart could not be created, %w", err)
			}
			createdCart := models.Cart{
				ID:        cartID,
				UserID:    id,
				CreatedAt: now,
				UpdatedAt: now,
			}
			return createdCart, nil // how to return the created cart easily?
		}
		return models.Cart{}, fmt.Errorf("Error scanning carts row: %w", err)
	}
	return cart, nil
}

func (c *CartsStore) EmptyCart(ctx context.Context, userID string) error {
	sql := `
	DELETE FROM cart_items
	WHERE cart_id IN ( SELECT id FROM carts WHERE user_id=$1)
	`

	_, err := c.dbpool.Exec(ctx, sql, userID)
	if err != nil {
		return fmt.Errorf("Error deleting cart from user %s: %w", userID, err)
	}

	return nil
}

func (c *CartsStore) EmptyCartTx(ctx context.Context, userID string, tx pgx.Tx) error {
	sql := `
	DELETE FROM cart_items
	WHERE cart_id IN ( SELECT id FROM carts WHERE user_id=$1)
	`

	_, err := tx.Exec(ctx, sql, userID)
	if err != nil {
		return fmt.Errorf("Error deleting cart from user %s: %w", userID, err)
	}

	return nil
}

func (c *CartsStore) GetCartItemByID(ctx context.Context, id string) (models.CartItems, error) {
	sql := `
	SELECT id, product_id, cart_id, quantity, created_at, updated_at
	FROM cart_items
	WHERE id=$1
	`

	row := c.dbpool.QueryRow(ctx, sql, id)

	var cartItem models.CartItems
	err := row.Scan(
		&cartItem.ID,
		&cartItem.ProductID,
		&cartItem.CartID,
		&cartItem.Quantity,
		&cartItem.CreatedAt,
		&cartItem.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return models.CartItems{}, &customerrors.ErrCartItemNotFound
		}
		return models.CartItems{}, fmt.Errorf("Error scanning products row: %w", err)
	}
	return cartItem, nil
}

func (c *CartsStore) AddItemToCart(ctx context.Context, cartID, productID string, quantity int) (models.CartItems, error) {
	sql := `SELECT id, cart_id, product_id, quantity, created_at, updated_at FROM cart_items WHERE cart_id=$1 AND product_id=$2`
	row := c.dbpool.QueryRow(ctx, sql, cartID, productID)
	var foundCartItem models.CartItems
	err := row.Scan(
		&foundCartItem.ID,
		&foundCartItem.ProductID,
		&foundCartItem.CartID,
		&foundCartItem.Quantity,
		&foundCartItem.CreatedAt,
		&foundCartItem.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {

			createdAt := time.Now()
			updatedAt := time.Now()
			ID := uuid.NewString()
			newCartItem := models.CartItems{
				ID:        ID,
				ProductID: productID,
				CartID:    cartID,
				Quantity:  quantity,
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
			}
			sql := `
		INSERT INTO cart_items (id, cart_id, product_id, quantity, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`
			_, err = c.dbpool.Exec(ctx, sql, ID, cartID, productID, quantity, createdAt, updatedAt)
			if err != nil {
				return models.CartItems{}, fmt.Errorf("Cart item could not be created, %w", err)
			}
			return newCartItem, nil
		}
		return models.CartItems{}, fmt.Errorf("Error finding cart items: %w", err)
	}
	sql = `
	UPDATE cart_items
	SET quantity=$1, updated_at=$2
	WHERE id=$3
	RETURNING id, quantity, updated_at`
	foundCartItem.UpdatedAt = time.Now()
	foundCartItem.Quantity = foundCartItem.Quantity + quantity
	_, err = c.dbpool.Exec(ctx, sql, foundCartItem.Quantity, foundCartItem.UpdatedAt, foundCartItem.ID)
	if err != nil {
		return models.CartItems{}, fmt.Errorf("Error updating cart items: %w", err)
	}
	return foundCartItem, nil
}

func (c *CartsStore) GetCartItems(ctx context.Context, cartID string) ([]models.CartItems, error) {
	sql := `
	SELECT id, product_id, cart_id, quantity, created_at, updated_at
	FROM cart_items
	WHERE cart_id=$1
	ORDER BY created_at DESC
	`

	rows, err := c.dbpool.Query(ctx, sql, cartID)
	if err != nil {
		return nil, fmt.Errorf("Error querying cart items: %w", err)
	}
	defer rows.Close()

	var cartItems []models.CartItems
	for rows.Next() {
		var cartItem models.CartItems
		err := rows.Scan(
			&cartItem.ID,
			&cartItem.ProductID,
			&cartItem.CartID,
			&cartItem.Quantity,
			&cartItem.CreatedAt,
			&cartItem.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("Error scanning products row: %w", err)
		}
		cartItems = append(cartItems, cartItem)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("Error iterating product rows: %w", err)
	}
	return cartItems, nil
}

func (c *CartsStore) UpdateCartItemQuantity(ctx context.Context, itemID string, newQuantity int) error {
	cartItem, err := c.GetCartItemByID(ctx, itemID)
	if err != nil {
		return fmt.Errorf("Error finding item with itemID %s, %w", itemID, err)
	}
	sql := `
	SELECT id, name, price, description, category, stock_quantity, created_at, updated_at
	FROM products
	WHERE product_id=$1
	RETURNING stock_quantity`

	row := c.dbpool.QueryRow(ctx, sql, cartItem.ProductID)
	var product models.Product
	err = row.Scan(
		&product.ID,
		&product.Name,
		&product.Price,
		&product.Description,
		&product.Category,
		&product.StockQuantity,
		&product.CreatedAt,
		&product.UpdatedAt,
	)
	if newQuantity <= product.StockQuantity {
		sql = `
	UPDATE cart_items
	SET quantity=$1
	WHERE id=$2
	`
		_, err = c.dbpool.Exec(ctx, sql, newQuantity, itemID)
		if err != nil {
			return fmt.Errorf("Error updating item %s quantity: %w", itemID, err)
		}
		return nil
	}
	return fmt.Errorf("Available stock for this product is %d, the requested quantity is %d", product.StockQuantity, newQuantity)
}

func (c *CartsStore) DeleteCartItem(ctx context.Context, cartItemID string) error {
	sql := `
	DELETE FROM cart_items
	WHERE id=$1
	`

	commandTag, err := c.dbpool.Exec(ctx, sql, cartItemID)
	if err != nil {
		return fmt.Errorf("Error deleting cart item: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("Product not found with id: %s", cartItemID)
	}
	return nil
}

func (c *CartsStore) GetCartItemByProductID(ctx context.Context, productID, cartID string) (models.CartItems, error) {
	sql := `SELECT id, cart_id, product_id, quantity, created_at, updated_at FROM cart_items WHERE cart_id=$1 AND product_id=$2`
	row := c.dbpool.QueryRow(ctx, sql, cartID, productID)

	var cartItem models.CartItems
	err := row.Scan(
		&cartItem.ID,
		&cartItem.ProductID,
		&cartItem.CartID,
		&cartItem.Quantity,
		&cartItem.CreatedAt,
		&cartItem.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return models.CartItems{}, &customerrors.ErrCartItemNotFound
		}
		return models.CartItems{}, fmt.Errorf("Error scanning products row: %w", err)
	}
	return cartItem, nil
}
