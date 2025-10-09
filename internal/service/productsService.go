package service

import (
	"context"
	"fmt"

	customerrors "github.com/alexalbu001/iguanas-jewelry-api/internal/customErrors"
	"github.com/alexalbu001/iguanas-jewelry-api/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type ProductsStore interface {
	GetAll(ctx context.Context) ([]models.Product, error)
	GetAllIncludingDeleted(ctx context.Context) ([]models.Product, error)
	GetByIDBatch(ctx context.Context, productIDs []string) (map[string]models.Product, error)
	GetByID(ctx context.Context, id string) (models.Product, error)
	Add(ctx context.Context, product models.Product) (models.Product, error)
	AddTx(ctx context.Context, product models.Product, tx pgx.Tx) (models.Product, error)
	Update(ctx context.Context, id string, product models.Product) (models.Product, error)
	Delete(ctx context.Context, id string) error
	DeleteTx(ctx context.Context, id string, tx pgx.Tx) error
	Restore(ctx context.Context, id string) error
	UpdateStock(ctx context.Context, productID string, stockChange int) error
	UpdateStockTx(ctx context.Context, productID string, stockChange int, tx pgx.Tx) error
	InvalidateProductCache(ctx context.Context, productIDs []string) error
}

type ProductsService struct {
	ProductsStore ProductsStore
}

func NewProductsService(productsStore ProductsStore) *ProductsService {
	return &ProductsService{
		ProductsStore: productsStore,
	}
}

func (p *ProductsService) GetProducts(ctx context.Context) ([]models.Product, error) {
	products, err := p.ProductsStore.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("Error fetching all products: %w", err)
	}
	return products, nil
}

func (p *ProductsService) PostProduct(ctx context.Context, product models.Product) (models.Product, error) {
	if product.Price < 0 {
		return models.Product{}, fmt.Errorf("Price %f should be > 0", product.Price)
	}
	if product.Category == "" {
		return models.Product{}, &customerrors.ErrMissingCategory
	}
	if product.Name == "" {
		return models.Product{}, &customerrors.ErrMissingName
	}
	newProduct, err := p.ProductsStore.Add(ctx, product)
	if err != nil {
		return models.Product{}, fmt.Errorf("Error creating product: %w", err)
	}
	return newProduct, nil
}

func (p *ProductsService) GetProductByID(ctx context.Context, productID string) (models.Product, error) {
	if productID == "" {
		return models.Product{}, &customerrors.ErrEmptyProductID
	}
	err := uuid.Validate(productID)
	if err != nil {
		return models.Product{}, &customerrors.ErrInvalidProductID
	}
	product, err := p.ProductsStore.GetByID(ctx, productID)
	if err != nil {
		return models.Product{}, fmt.Errorf("Error fetching product: %w", err)
	}
	return product, nil
}

func (p *ProductsService) UpdateProductByID(ctx context.Context, productID string, product models.Product) (models.Product, error) {
	if product.Name == "" {
		return models.Product{}, &customerrors.ErrMissingName
	}
	if product.Price <= 0 {
		return models.Product{}, &customerrors.ErrInvalidPrice
	}
	_, err := p.ProductsStore.GetByID(ctx, productID)
	if err != nil {
		return models.Product{}, fmt.Errorf("Invalid product ID: %w", err)
	}
	updatedProduct, err := p.ProductsStore.Update(ctx, productID, product)
	if err != nil {
		return models.Product{}, fmt.Errorf("Error fetching product: %w", err)
	}
	return updatedProduct, nil
}

func (p *ProductsService) DeleteProductByID(ctx context.Context, productID string) error {
	if productID == "" {
		return &customerrors.ErrEmptyProductID
	}
	err := uuid.Validate(productID)
	if err != nil {
		return &customerrors.ErrInvalidProductID
	}
	err = p.ProductsStore.Delete(ctx, productID)
	if err != nil {
		return fmt.Errorf("Error fetching product: %w", err)
	}
	return nil
}

func (p *ProductsService) UpdateStock(ctx context.Context, productID string, quantity int) error {
	if productID == "" {
		return &customerrors.ErrEmptyProductID
	}
	err := uuid.Validate(productID)
	if err != nil {
		return &customerrors.ErrInvalidProductID
	}

	err = p.ProductsStore.UpdateStock(ctx, productID, quantity)
	if err != nil {
		return fmt.Errorf("Failed to update product stock: %w", err)
	}
	return nil
}

// Get all products including soft-deleted ones (for admin)
func (p *ProductsService) GetAllProductsIncludingDeleted(ctx context.Context) ([]models.Product, error) {
	products, err := p.ProductsStore.GetAllIncludingDeleted(ctx)
	if err != nil {
		return nil, fmt.Errorf("Error fetching all products including deleted: %w", err)
	}
	return products, nil
}

// Restore soft-deleted product
func (p *ProductsService) RestoreProduct(ctx context.Context, productID string) error {
	if productID == "" {
		return &customerrors.ErrEmptyProductID
	}
	err := uuid.Validate(productID)
	if err != nil {
		return &customerrors.ErrInvalidProductID
	}

	err = p.ProductsStore.Restore(ctx, productID)
	if err != nil {
		return fmt.Errorf("Error restoring product: %w", err)
	}
	return nil
}

// InvalidateAllProductCache invalidates the entire product cache
func (p *ProductsService) InvalidateAllProductCache(ctx context.Context) error {
	// Passing empty slice will invalidate the "products:all" key in the cache implementation
	err := p.ProductsStore.InvalidateProductCache(ctx, []string{})
	if err != nil {
		return fmt.Errorf("Error invalidating product cache: %w", err)
	}
	return nil
}
