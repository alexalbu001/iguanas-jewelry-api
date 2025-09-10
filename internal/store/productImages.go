package store

import (
	"context"
	"fmt"
	"strings"
	"time"

	customerrors "github.com/alexalbu001/iguanas-jewelry-api/internal/customErrors"
	"github.com/alexalbu001/iguanas-jewelry-api/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductImagesStore struct {
	dbpool *pgxpool.Pool
}

func NewProductImagesStore(dbpool *pgxpool.Pool) *ProductImagesStore {
	return &ProductImagesStore{dbpool: dbpool}
}

func (p *ProductImagesStore) GetByProductID(ctx context.Context, productID string) ([]models.ProductImage, error) {
	sql := `
	SELECT id, product_id, image_key, content_type, is_main, display_order, created_at, updated_at
	FROM product_images
	WHERE product_id=$1
	`
	rows, err := p.dbpool.Query(ctx, sql, productID)
	if err != nil {
		return nil, fmt.Errorf("Error querying product images: %w", err)
	}
	defer rows.Close()
	var productImages []models.ProductImage
	for rows.Next() {
		var productImage models.ProductImage
		err := rows.Scan(
			&productImage.ID,
			&productImage.ProductID,
			&productImage.ImageKey,
			&productImage.ContentType,
			&productImage.IsMain,
			&productImage.DisplayOrder,
			&productImage.CreatedAt,
			&productImage.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("Error scanning product images row: %w", err)
		}
		productImages = append(productImages, productImage)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("Error iterating product images rows: %w", err)
	}
	return productImages, nil
}

func (p *ProductImagesStore) InsertProductImage(ctx context.Context, productImage models.ProductImage) (models.ProductImage, error) {
	productImage.ID = uuid.NewString()
	productImage.CreatedAt = time.Now()
	productImage.UpdatedAt = time.Now()

	sql := `
	INSERT INTO product_images (id, product_id, image_key, content_type, is_main, display_order, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := p.dbpool.Exec(ctx, sql, productImage.ID, productImage.ProductID, productImage.ImageKey, productImage.ContentType, productImage.IsMain, productImage.DisplayOrder, productImage.CreatedAt, productImage.UpdatedAt)
	if err != nil {
		return models.ProductImage{}, fmt.Errorf("Product image could not be created: %w", err)
	}

	return productImage, nil
}

func (p *ProductImagesStore) InsertProductImageBulk(ctx context.Context, productImages []models.ProductImage) error {
	for _, productImage := range productImages {
		_, err := p.InsertProductImage(ctx, productImage)
		if err != nil {
			return fmt.Errorf("Product image in bulk could not be created, %w", err)
		}
	}
	return nil
}

func (p *ProductImagesStore) GetPrimaryImageForProduct(ctx context.Context, productID string) (models.ProductImage, error) {
	sql := `
	SELECT id, product_id, image_key, content_type, is_main, display_order, created_at, updated_at
	FROM product_images
	WHERE product_id=$1 AND is_main=true
	`
	row := p.dbpool.QueryRow(ctx, sql, productID)
	var productImage models.ProductImage
	err := row.Scan(
		&productImage.ID,
		&productImage.ProductID,
		&productImage.ImageKey,
		&productImage.ContentType,
		&productImage.IsMain,
		&productImage.DisplayOrder,
		&productImage.CreatedAt,
		&productImage.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return models.ProductImage{}, &customerrors.ErrProductImageNotFound
		}
		return models.ProductImage{}, fmt.Errorf("Error scanning product images row: %w", err)
	}
	return productImage, nil
}

func (p *ProductImagesStore) GetPrimaryImageForProductBulk(ctx context.Context, productIDs []string) (map[string]models.ProductImage, error) {
	sql := `
	SELECT id, product_id, image_key, content_type, is_main, display_order, created_at, updated_at
	FROM product_images
	WHERE product_id=ANY($1) AND is_main=true
	`

	rows, err := p.dbpool.Query(ctx, sql, productIDs)
	if err != nil {
		return nil, fmt.Errorf("Error querying product images: %w", err)
	}
	defer rows.Close()
	productImagesMap := make(map[string]models.ProductImage)
	for rows.Next() {
		var productImage models.ProductImage
		err := rows.Scan(
			&productImage.ID,
			&productImage.ProductID,
			&productImage.ImageKey,
			&productImage.ContentType,
			&productImage.IsMain,
			&productImage.DisplayOrder,
			&productImage.CreatedAt,
			&productImage.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("Error scanning product images row: %w", err)
		}
		productImagesMap[productImage.ProductID] = productImage
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("Error iterating product images rows: %w", err)
	}
	return productImagesMap, nil
}

func (p *ProductImagesStore) UpdateProductImage(ctx context.Context, productImage models.ProductImage) error {
	productImage.UpdatedAt = time.Now()
	sql := `
	UPDATE product_images
	SET image_key=$1, is_main=$2, display_order=$3, updated_at=$4
	WHERE id=$5
	`
	_, err := p.dbpool.Exec(ctx, sql, productImage.ImageKey, productImage.IsMain, productImage.DisplayOrder, productImage.UpdatedAt, productImage.ID)
	if err != nil {
		return fmt.Errorf("Error updating product image: %w", err)
	}
	return nil
}

func (p *ProductImagesStore) DeleteProductImage(ctx context.Context, imageID, productID string) error {
	sql := `
	DELETE FROM product_images
	WHERE id=$1 AND product_id=$2
	`
	_, err := p.dbpool.Exec(ctx, sql, imageID, productID)
	if err != nil {
		return fmt.Errorf("Error deleting product image: %w", err)
	}
	return nil
}

func (p *ProductImagesStore) GetProductImagesBulk(ctx context.Context, productIDs []string) (map[string][]models.ProductImage, error) {
	sql := `
	SELECT id, product_id, image_key, content_type, is_main, display_order, created_at, updated_at
	FROM product_images
	WHERE product_id=ANY($1)
	`
	rows, err := p.dbpool.Query(ctx, sql, productIDs)
	if err != nil {
		return nil, fmt.Errorf("Error querying product images: %w", err)
	}
	defer rows.Close()
	productImagesMap := make(map[string][]models.ProductImage)
	for rows.Next() {
		var productImage models.ProductImage
		err := rows.Scan(
			&productImage.ID,
			&productImage.ProductID,
			&productImage.ImageKey,
			&productImage.ContentType,
			&productImage.IsMain,
			&productImage.DisplayOrder,
			&productImage.CreatedAt,
			&productImage.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("Error scanning product images row: %w", err)
		}
		productImagesMap[productImage.ProductID] = append(productImagesMap[productImage.ProductID], productImage)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("Error iterating product images rows: %w", err)
	}
	return productImagesMap, nil
}

func (p *ProductImagesStore) SetPrimaryImage(ctx context.Context, productID, imageID string) error {
	sql := `
	UPDATE product_images
	SET is_main=$1, updated_at=$2
	WHERE product_id=$3 AND id=$4
	`

	commandTag, err := p.dbpool.Exec(ctx, sql, true, time.Now(), productID, imageID)
	if err != nil {
		return fmt.Errorf("Error updating product image: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return &customerrors.ErrProductImageNotFound
	}
	return nil
}

func (p *ProductImagesStore) SetPrimaryImageTx(ctx context.Context, productID, imageID string, tx pgx.Tx) error {
	sql := `
	UPDATE product_images
	SET is_main=$1, updated_at=$2
	WHERE product_id=$3 AND id=$4
	`

	commandTag, err := tx.Exec(ctx, sql, true, time.Now(), productID, imageID)
	if err != nil {
		return fmt.Errorf("Error updating product image: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return &customerrors.ErrProductImageNotFound
	}
	return nil
}

func (p *ProductImagesStore) UnsetPrimaryImage(ctx context.Context, productID string) error {
	sql := `
	UPDATE product_images
	SET is_main=$1, updated_at=$2
	WHERE product_id=$3 AND is_main=true
	`

	commandTag, err := p.dbpool.Exec(ctx, sql, false, time.Now(), productID)
	if err != nil {
		return fmt.Errorf("Error updating product image: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return &customerrors.ErrProductImageNotFound
	}
	return nil
}

func (p *ProductImagesStore) UnsetPrimaryImageTx(ctx context.Context, productID string, tx pgx.Tx) error {
	sql := `
	UPDATE product_images
	SET is_main=$1, updated_at=$2
	WHERE product_id=$3 AND is_main=true
	`

	commandTag, err := tx.Exec(ctx, sql, false, time.Now(), productID)
	if err != nil {
		return fmt.Errorf("Error updating product image: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return &customerrors.ErrProductImageNotFound
	}
	return nil
}

func (s *ProductImagesStore) UpdateImageOrder(ctx context.Context, imageOrder []string) error {
	// Build the CASE statement dynamically
	query := "UPDATE product_images SET display_order = CASE "

	var args []interface{}
	var placeholders []string

	for i, imageID := range imageOrder {
		newOrder := i + 1 // 1-based ordering
		query += fmt.Sprintf("WHEN id = $%d THEN %d ", len(args)+1, newOrder)
		args = append(args, imageID)
		placeholders = append(placeholders, fmt.Sprintf("$%d", len(args)))
	}

	query += "END WHERE id IN (" + strings.Join(placeholders, ",") + ")"

	_, err := s.dbpool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("Failed to reorder images: %w", err)
	}
	return nil
}
