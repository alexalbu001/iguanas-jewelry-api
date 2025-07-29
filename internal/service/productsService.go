package service

import (
	"context"
	"fmt"

	"github.com/alexalbu001/iguanas-jewelry/internal/models"
	"github.com/google/uuid"
)

type ProductsStore interface {
	GetAll(ctx context.Context) ([]models.Product, error)
	GetByID(ctx context.Context, id string) (models.Product, error)
	Add(ctx context.Context, product models.Product) (models.Product, error)
	Update(ctx context.Context, id string, product models.Product) (models.Product, error)
	Delete(ctx context.Context, id string) error
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
		return models.Product{}, fmt.Errorf("Enter a product category")
	}
	if product.Name == "" {
		return models.Product{}, fmt.Errorf("Enter a product name")
	}
	newProduct, err := p.ProductsStore.Add(ctx, product)
	if err != nil {
		return models.Product{}, fmt.Errorf("Error creating product: %w", err)
	}
	return newProduct, nil
}

func (p *ProductsService) GetProductByID(ctx context.Context, productID string) (models.Product, error) {
	if productID == "" {
		return models.Product{}, fmt.Errorf("Product ID can't be empty")
	}
	err := uuid.Validate(productID)
	if err != nil {
		return models.Product{}, fmt.Errorf("Invalid product ID")
	}
	product, err := p.ProductsStore.GetByID(ctx, productID)
	if err != nil {
		return models.Product{}, fmt.Errorf("Error fetching product")
	}
	return product, nil
}

func (p *ProductsService) UpdateProductByID(ctx context.Context, productID string, product models.Product) (models.Product, error) {
	if product.Name == "" {
		return models.Product{}, fmt.Errorf("Product name can't be empty")
	}
	if product.Price <= 0 {
		return models.Product{}, fmt.Errorf("Product price %f should be > 0", product.Price)
	}
	err := uuid.Validate(productID)
	if err != nil {
		return models.Product{}, fmt.Errorf("Invalid product ID")
	}
	updatedProduct, err := p.ProductsStore.Update(ctx, productID, product)
	if err != nil {
		return models.Product{}, fmt.Errorf("Error fetching product")
	}
	return updatedProduct, nil
}

func (p *ProductsService) DeleteProductByID(ctx context.Context, productID string) error {
	if productID == "" {
		fmt.Errorf("Product ID can't be empty")
	}
	err := uuid.Validate(productID)
	if err != nil {
		fmt.Errorf("Invalid product ID")
	}
	err = p.ProductsStore.Delete(ctx, productID)
	if err != nil {
		fmt.Errorf("Error fetching product")
	}
	return nil
}
