package store

import (
	"context"
	"fmt"
	"time"

	customerrors "github.com/alexalbu001/iguanas-jewelry-api/internal/customErrors"
	"github.com/alexalbu001/iguanas-jewelry-api/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// var Products = []models.Product{} // should not live in models because its storage, not pure data definition
// Slices, M
// The ProductStore only knows about storage operations/Defines HOW to work with products (the operations)
type ProductsStore struct {
	dbpool *pgxpool.Pool
}

func NewProductStore(connection *pgxpool.Pool) *ProductsStore { // constructor
	return &ProductsStore{
		dbpool: connection,
	}
}

// This function demonstrates several important patterns when working with PGX:
// We use pool.Query() to execute a SELECT statement that returns multiple rows
// We use defer rows.Close() to ensure resources are cleaned up when the function exits
// We iterate through results with rows.Next() and use rows.Scan() to populate our Task struct
// We check for errors after iteration with rows.Err()
// The order of columns in your Scan() call must match the order of columns in your SELECT statement. PGX doesn't do any mapping by column name.

func (h *ProductsStore) GetAll(ctx context.Context) ([]models.Product, error) {
	sql := `
	SELECT id, name, price, description, category, stock_quantity, created_at, updated_at, deleted_at
	FROM products
	WHERE deleted_at IS NULL
	ORDER BY created_at DESC
	`
	rows, err := h.dbpool.Query(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("Error querying products: %w", err)
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var product models.Product
		err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.Price,
			&product.Description,
			&product.Category,
			&product.StockQuantity,
			&product.CreatedAt,
			&product.UpdatedAt,
			&product.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("Error scanning products row: %w", err)
		}
		products = append(products, product)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("Error iterating product rows: %w", err)
	}
	return products, nil
	// return h.store
}

func (h *ProductsStore) GetByIDBatch(ctx context.Context, productIDs []string) (map[string]models.Product, error) { //more effiecient to lookup maps than to iterate over []models.Product

	sql := `
	SELECT id, name, price, description, category, stock_quantity, created_at, updated_at
	FROM products
	WHERE id=ANY($1)
	`

	rows, err := h.dbpool.Query(ctx, sql, productIDs)
	if err != nil {
		return nil, fmt.Errorf("Error querying products: %w", err)
	}
	defer rows.Close()

	productsMap := make(map[string]models.Product)
	if len(productIDs) == 0 {
		return productsMap, nil
	}

	for rows.Next() {
		var product models.Product
		err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.Price,
			&product.Description,
			&product.Category,
			&product.StockQuantity,
			&product.CreatedAt,
			&product.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("Error scanning products row: %w", err)
		}

		productsMap[product.ID] = product
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("Error iterating product rows: %w", err)
	}
	return productsMap, nil
}

func (h *ProductsStore) GetByID(ctx context.Context, id string) (models.Product, error) {
	sql := `
SELECT id, name, price, description, category, stock_quantity, created_at, updated_at
FROM products
WHERE id=$1
`

	row := h.dbpool.QueryRow(ctx, sql, id)

	var product models.Product
	err := row.Scan(
		&product.ID,
		&product.Name,
		&product.Price,
		&product.Description,
		&product.Category,
		&product.StockQuantity,
		&product.CreatedAt,
		&product.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return models.Product{}, &customerrors.ErrProductNotFound
		}
		return models.Product{}, fmt.Errorf("Error scanning products row: %w", err)
	}
	return product, nil

	// for _, product := range h.store {
	// 	if id == product.ID {
	// 		return product, nil
	// 	}
	// }
	// return models.Product{}, fmt.Errorf("Product not found: %s", id)
}

func (h *ProductsStore) Add(ctx context.Context, product models.Product) (models.Product, error) {
	product.ID = uuid.NewString()
	product.CreatedAt = time.Now() // Set creation time
	product.UpdatedAt = time.Now() // Set update time

	sql := `
	INSERT INTO products (id, name, price, description, category, stock_quantity, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := h.dbpool.Exec(ctx, sql, product.ID, product.Name, product.Price, product.Description, product.Category, product.StockQuantity, product.CreatedAt, product.UpdatedAt)
	if err != nil {
		return models.Product{}, fmt.Errorf("Product could not be created, %w", err)
	}
	return product, nil
	// h.store = append(h.store, product)
	// return product, nil
}

func (h *ProductsStore) AddTx(ctx context.Context, product models.Product, tx pgx.Tx) (models.Product, error) {
	product.ID = uuid.NewString()
	product.CreatedAt = time.Now() // Set creation time
	product.UpdatedAt = time.Now() // Set update time

	sql := `
	INSERT INTO products (id, name, price, description, category, stock_quantity, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := tx.Exec(ctx, sql, product.ID, product.Name, product.Price, product.Description, product.Category, product.StockQuantity, product.CreatedAt, product.UpdatedAt)
	if err != nil {
		return models.Product{}, fmt.Errorf("Product could not be created, %w", err)
	}
	return product, nil
	// h.store = append(h.store, product)
	// return product, nil
}

func (h *ProductsStore) Update(ctx context.Context, id string, product models.Product) (models.Product, error) {
	product.UpdatedAt = time.Now()
	// product.CreatedAt
	sql := `
	UPDATE products
	SET name=$1, price=$2, description=$3, category=$4, stock_quantity=$5, updated_at=$6
	WHERE id=$7
	RETURNING id, created_at`

	row := h.dbpool.QueryRow(ctx, sql, product.Name, product.Price, product.Description, product.Category, product.StockQuantity, product.UpdatedAt, id)

	var newProduct models.Product

	err := row.Scan(&newProduct.ID, &newProduct.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return models.Product{}, &customerrors.ErrProductNotFound
		}
		return models.Product{}, fmt.Errorf("Error scanning row: %w", err)
	}
	newProduct.Name = product.Name
	newProduct.Price = product.Price
	newProduct.Description = product.Description
	newProduct.Category = product.Category
	newProduct.StockQuantity = product.StockQuantity
	newProduct.UpdatedAt = product.UpdatedAt

	return newProduct, nil
	// }
	// for i, value := range h.store {
	// 	if id == value.ID {
	// 		product.ID = id
	// 		product.CreatedAt = value.CreatedAt // Preserve original
	// 		product.UpdatedAt = time.Now()
	// 		h.store[i] = product
	// 		return product, nil
	// 	}
	// return models.Product{}, fmt.Errorf("ID: %s not found", id)
}

func (h *ProductsStore) Delete(ctx context.Context, id string) error {
	sql := `
	UPDATE products
	SET deleted_at = $1
	WHERE id = $2 AND deleted_at IS NULL`

	commandTag, err := h.dbpool.Exec(ctx, sql, time.Now(), id)
	if err != nil {
		return fmt.Errorf("Error soft deleting product: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return &customerrors.ErrProductNotFound
	}
	return nil
	// for i, value := range h.store {
	// 	if id == value.ID {
	// 		h.store = append(h.store[:i], h.store[i+1:]...)
	// 		return nil
	// 	}
	// }
	// return fmt.Errorf("ID: %s not found", id)
}

func (h *ProductsStore) DeleteTx(ctx context.Context, id string, tx pgx.Tx) error {
	sql := `
	UPDATE products
	SET deleted_at = $1
	WHERE id = $2 AND deleted_at IS NULL`

	commandTag, err := tx.Exec(ctx, sql, time.Now(), id)
	if err != nil {
		return fmt.Errorf("Error soft deleting product: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return &customerrors.ErrProductNotFound
	}
	return nil
	// for i, value := range h.store {
	// 	if id == value.ID {
	// 		h.store = append(h.store[:i], h.store[i+1:]...)
	// 		return nil
	// 	}
	// }
	// return fmt.Errorf("ID: %s not found", id)
}

func (h *ProductsStore) UpdateStock(ctx context.Context, productID string, stockChange int) error {
	sql := `
	UPDATE products
	SET stock_quantity=stock_quantity + $2,
		updated_at=$3
	WHERE id=$1
	AND stock_quantity +$2 >=0
	`

	result, err := h.dbpool.Exec(ctx, sql, productID, stockChange, time.Now())
	if err != nil {
		return fmt.Errorf("Error updateing product stock: %w", err)
	}

	if result.RowsAffected() == 0 {
		return &customerrors.ErrInsufficientStockStatic // or product not found
	}

	return nil
}

func (h *ProductsStore) UpdateStockTx(ctx context.Context, productID string, stockChange int, tx pgx.Tx) error {
	sql := `
	UPDATE products
	SET stock_quantity=stock_quantity + $2,
		updated_at=$3
	WHERE id=$1
	AND stock_quantity +$2 >=0
	`

	result, err := tx.Exec(ctx, sql, productID, stockChange, time.Now())
	if err != nil {
		return fmt.Errorf("Error updateing product stock: %w", err)
	}

	if result.RowsAffected() == 0 {
		return &customerrors.ErrInsufficientStockStatic // or product not found
	}

	return nil
}

// InvalidateProductCache is a no-op for the non-cached store
func (h *ProductsStore) InvalidateProductCache(ctx context.Context, productIDs []string) error {
	// No caching in base store, so nothing to invalidate
	return nil
}

// Restore soft-deleted product
func (h *ProductsStore) Restore(ctx context.Context, id string) error {
	sql := `
	UPDATE products
	SET deleted_at = NULL
	WHERE id = $1 AND deleted_at IS NOT NULL`

	commandTag, err := h.dbpool.Exec(ctx, sql, id)
	if err != nil {
		return fmt.Errorf("Error restoring product: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return &customerrors.ErrProductNotFound
	}
	return nil
}

// Get all products including soft-deleted ones (for admin)
func (h *ProductsStore) GetAllIncludingDeleted(ctx context.Context) ([]models.Product, error) {
	sql := `
	SELECT id, name, price, description, category, stock_quantity, created_at, updated_at, deleted_at
	FROM products
	ORDER BY created_at DESC
	`
	rows, err := h.dbpool.Query(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("Error querying products: %w", err)
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var product models.Product
		err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.Price,
			&product.Description,
			&product.Category,
			&product.StockQuantity,
			&product.CreatedAt,
			&product.UpdatedAt,
			&product.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("Error scanning products row: %w", err)
		}
		products = append(products, product)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("Error iterating product rows: %w", err)
	}
	return products, nil
}
