package store

import (
	"context"
	"fmt"
	"time"

	"github.com/alexalbu001/iguanas-jewelry/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrdersStore struct {
	dbpool *pgxpool.Pool
	tx     pgx.Tx
	// carts    *CartsStore
	// products *ProductsStore
}

func NewOrdersStore(connection *pgxpool.Pool, carts *CartsStore, products *ProductsStore) *OrdersStore {
	return &OrdersStore{
		dbpool: connection,
		// carts:    carts,
		// products: products,
	}
}

func (o *OrdersStore) InsertOrder(ctx context.Context, order models.Order) error {
	sql := `
	INSERT INTO orders (id, user_id, total_amount, status, shipping_name, shipping_email, shipping_phone, shipping_address_line_1, shipping_address_line_2, shipping_city, shipping_state, shipping_postal_code, shipping_country, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`

	_, err := o.dbpool.Exec(ctx, sql, order.ID, order.UserID, order.TotalAmount, order.Status, order.ShippingName, order.ShippingEmail, order.ShippingPhone, order.ShippingAddressLine1, order.ShippingAddressLine2, order.ShippingCity, order.ShippingState, order.ShippingPostalCode, order.ShippingCountry, order.CreatedAt, order.UpdatedAt)
	if err != nil {
		return fmt.Errorf("Order could not be created, %w", err)
	}
	return nil
}

func (o *OrdersStore) InsertOrderTx(ctx context.Context, order models.Order, tx pgx.Tx) error {
	sql := `
	INSERT INTO orders (id, user_id, total_amount, status, shipping_name, shipping_email, shipping_phone, shipping_address_line_1, shipping_address_line_2, shipping_city, shipping_state, shipping_postal_code, shipping_country, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`

	_, err := tx.Exec(ctx, sql, order.ID, order.UserID, order.TotalAmount, order.Status, order.ShippingName, order.ShippingEmail, order.ShippingPhone, order.ShippingAddressLine1, order.ShippingAddressLine2, order.ShippingCity, order.ShippingState, order.ShippingPostalCode, order.ShippingCountry, order.CreatedAt, order.UpdatedAt)
	if err != nil {
		return fmt.Errorf("Order could not be created, %w", err)
	}
	return nil
}

func (o *OrdersStore) InsertOrderItem(ctx context.Context, orderItem models.OrderItem) error {
	sql := `
INSERT INTO order_items (id, order_id, product_id, quantity, price, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := o.dbpool.Exec(ctx, sql, orderItem.ID, orderItem.OrderID, orderItem.ProductID, orderItem.Quantity, orderItem.Price, orderItem.CreatedAt, orderItem.UpdatedAt)
	if err != nil {
		return fmt.Errorf("Order item could not be created, %w", err)
	}
	return nil
}

func (o *OrdersStore) InsertOrderItemTx(ctx context.Context, orderItem models.OrderItem, tx pgx.Tx) error {
	sql := `
INSERT INTO order_items (id, order_id, product_id, quantity, price, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := tx.Exec(ctx, sql, orderItem.ID, orderItem.OrderID, orderItem.ProductID, orderItem.Quantity, orderItem.Price, orderItem.CreatedAt, orderItem.UpdatedAt)
	if err != nil {
		return fmt.Errorf("Order item could not be created, %w", err)
	}
	return nil
}

func (o *OrdersStore) InsertOrderItemBulk(ctx context.Context, orderItems []models.OrderItem) error {
	for _, orderItem := range orderItems {
		err := o.InsertOrderItem(ctx, orderItem)
		if err != nil {
			return fmt.Errorf("Order item in bulk could not be created, %w", err)
		}
	}
	return nil
}

func (o *OrdersStore) InsertOrderItemBulkTx(ctx context.Context, orderItems []models.OrderItem, tx pgx.Tx) error {
	for _, orderItem := range orderItems {
		err := o.InsertOrderItemTx(ctx, orderItem, tx)
		if err != nil {
			return fmt.Errorf("Order item in bulk could not be created, %w", err)
		}
	}
	return nil
}

func (o *OrdersStore) GetOrderByID(ctx context.Context, orderID string) (models.Order, error) {
	sql := `
SELECT id, user_id, total_amount, status, shipping_name, shipping_email, shipping_phone, shipping_address_line_1, shipping_address_line_2, shipping_city, shipping_state, shipping_postal_code, shipping_country, created_at, updated_at
FROM orders
WHERE id=$1
`

	row := o.dbpool.QueryRow(ctx, sql, orderID)
	var order models.Order
	err := row.Scan(
		&order.ID,
		&order.UserID,
		&order.TotalAmount,
		&order.Status,
		&order.ShippingName,
		&order.ShippingEmail,
		&order.ShippingPhone,
		&order.ShippingAddressLine1,
		&order.ShippingAddressLine2,
		&order.ShippingCity,
		&order.ShippingState,
		&order.ShippingPostalCode,
		&order.ShippingCountry,
		&order.CreatedAt,
		&order.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return models.Order{}, fmt.Errorf("Product not found with id: %s", orderID)
		}
		return models.Order{}, fmt.Errorf("Error scanning orders: %w", err)
	}
	return order, nil
}

func (o *OrdersStore) GetOrderItems(ctx context.Context, orderID string) ([]models.OrderItem, error) {
	sql := `
	SELECT id, order_id, product_id, quantity, price, created_at, updated_at
	FROM order_items
	WHERE order_id=$1
	`

	rows, err := o.dbpool.Query(ctx, sql, orderID)
	if err != nil {
		return nil, fmt.Errorf("Error querying order items: %w", err)
	}
	defer rows.Close()

	var orderItems []models.OrderItem
	for rows.Next() {
		var orderItem models.OrderItem
		err := rows.Scan(
			&orderItem.ID,
			&orderItem.OrderID,
			&orderItem.ProductID,
			&orderItem.Quantity,
			&orderItem.Price,
			&orderItem.CreatedAt,
			&orderItem.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("Error scanning order items row: %w", err)
		}
		orderItems = append(orderItems, orderItem)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("Error iterating order items rows: %w", err)
	}
	return orderItems, nil
}

func (o *OrdersStore) GetUsersOrders(ctx context.Context, userID string) ([]models.Order, error) {
	sql := `
SELECT id, user_id, total_amount, status, shipping_name, shipping_email, shipping_phone, shipping_address_line_1, shipping_address_line_2, shipping_city, shipping_state, shipping_postal_code, shipping_country, created_at, updated_at
FROM orders
WHERE user_id=$1 
`
	rows, err := o.dbpool.Query(ctx, sql, userID)
	defer rows.Close()
	if err != nil {
		return nil, fmt.Errorf("Error querying orders: %w", err)
	}
	var orders []models.Order
	for rows.Next() {
		var order models.Order
		err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.TotalAmount,
			&order.Status,
			&order.ShippingName,
			&order.ShippingEmail,
			&order.ShippingPhone,
			&order.ShippingAddressLine1,
			&order.ShippingAddressLine2,
			&order.ShippingCity,
			&order.ShippingState,
			&order.ShippingPostalCode,
			&order.ShippingCountry,
			&order.CreatedAt,
			&order.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("Error scanning orders row: %w", err)
		}
		orders = append(orders, order)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("Error iterating order items rows: %w", err)
	}
	return orders, nil
}

func (o *OrdersStore) UpdateOrderStatus(ctx context.Context, status, orderID string) error {
	sql := `
	UPDATE orders
	SET status=$1, updated_at=$2
	WHERE id=$3
	`

	commandTag, err := o.dbpool.Exec(ctx, sql, status, time.Now(), orderID)
	if err != nil {
		return fmt.Errorf("Error updating order status: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("Order not found with id: %s", orderID)
	}
	return nil
}
