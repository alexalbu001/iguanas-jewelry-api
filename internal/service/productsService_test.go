package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/alexalbu001/iguanas-jewelry/internal/models"
	"github.com/alexalbu001/iguanas-jewelry/internal/service"
	"github.com/alexalbu001/iguanas-jewelry/internal/utils"
	"github.com/google/uuid"
)

func TestGetProducts(t *testing.T) {
	mockProductStore := &utils.MockProductStore{
		Store: []models.Product{
			{ID: uuid.NewString(), Name: "Gold Ring", Price: 99.99, Description: "Nice gold ring", Category: "rings", StockQuantity: 2, CreatedAt: time.Now(), UpdatedAt: time.Now()},
			{ID: uuid.NewString(), Name: "Silver Necklace", Price: 10.99, Description: "Nice silver necklace", Category: "necklaces", StockQuantity: 2, CreatedAt: time.Now(), UpdatedAt: time.Now()},
			{ID: uuid.NewString(), Name: "Gold Necklace", Price: 5.99, Description: "Nice silver necklace", Category: "necklaces", StockQuantity: 0, CreatedAt: time.Now(), UpdatedAt: time.Now()},
			{ID: uuid.NewString(), Name: "Silver earings", Price: 10.99, Description: "Nice silver earings", Category: "earings", StockQuantity: 4, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		},
	}

	testProductService := service.NewProductsService(mockProductStore)
	products, err := testProductService.GetProducts(context.Background())
	if err != nil {
		t.Fatalf("Failed to get products: %v", err)
	}

	if len(products) != 4 {
		t.Errorf("expected 4 arguments, got %d instead", len(products))
	}

	if products[0].Name != "Gold Ring" {
		t.Errorf("Expected first product name 'Gold Ring', got %s", products[0].Name)
	}

	if products[3].Name != "Silver earings" {
		t.Errorf("Expected first product name 'Silver earings', got %s", products[0].Name)
	}
}

func TestPostProducts(t *testing.T) {
	mockProduct := models.Product{
		Name: "Gold Ring", Price: 99.99, Description: "Nice gold ring", Category: "rings",
	}

	mockProductStore := &utils.MockProductStore{
		Store: []models.Product{
			{ID: uuid.NewString(), Name: "White Ring", Price: 22.99, Description: "Nice white ring", Category: "rings", StockQuantity: 2, CreatedAt: time.Now(), UpdatedAt: time.Now()},
			{ID: uuid.NewString(), Name: "Silver Necklace", Price: 10.99, Description: "Nice silver necklace", Category: "earings", StockQuantity: 2, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		},
	}

	testProductService := service.NewProductsService(mockProductStore)

	_, err := testProductService.PostProduct(context.Background(), mockProduct)
	if err != nil {
		t.Fatalf("Failed to post product: %v", err)
	}

	products, err := testProductService.GetProducts(context.Background())
	if err != nil {
		t.Fatalf("Failed to get products: %v", err)
	}

	if len(products) != 3 {
		t.Errorf("expected 3 arguments, got %d instead", len(products))
	}

	if products[2].Name != "Gold Ring" {
		t.Errorf("Expected first product name 'Gold Ring', got %s", products[2].Name)
	}
}

func TestGetProductByID(t *testing.T) {
	mockProductID := uuid.New().String()
	mockProductStore := &utils.MockProductStore{
		Store: []models.Product{
			{ID: mockProductID, Name: "White Ring", Price: 22.99, Description: "Nice white ring", Category: "rings", StockQuantity: 2, CreatedAt: time.Now(), UpdatedAt: time.Now()},
			{ID: uuid.NewString(), Name: "Silver Necklace", Price: 10.99, Description: "Nice silver necklace", Category: "earings", StockQuantity: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		},
	}

	testProductService := service.NewProductsService(mockProductStore)

	product, err := testProductService.ProductsStore.GetByID(context.Background(), mockProductID)
	if err != nil {
		t.Fatalf("Failed to get product: %v", err)
	}

	if product.ID != mockProductID {
		t.Fatalf("expected %s, got %s", mockProductID, product.ID)
	}
}

func TestUpdateProductByID(t *testing.T) {
	mockProductID := uuid.New().String()
	mockProductStore := &utils.MockProductStore{
		Store: []models.Product{
			{ID: mockProductID, Name: "White Ring", Price: 22.99, Description: "Nice white ring", Category: "rings", StockQuantity: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()},
			{ID: uuid.NewString(), Name: "Silver Necklace", Price: 10.99, Description: "Nice silver necklace", Category: "earings", StockQuantity: 2, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		},
	}
	mockNewProduct := models.Product{
		Name: "Gold Ring", Price: 99.99, Description: "Nice gold ring", Category: "rings",
	}

	testProductService := service.NewProductsService(mockProductStore)
	updatedProduct, err := testProductService.UpdateProductByID(context.Background(), mockProductID, mockNewProduct)
	if err != nil {
		t.Fatalf("Failed to update product: %v", err)
	}

	fetchedProduct, err := testProductService.GetProductByID(context.Background(), mockProductID)
	if err != nil {
		t.Fatalf("Failed to get product: %v", err)
	}
	if updatedProduct.Name != fetchedProduct.Name {
		t.Fatalf("expected name %s, go %s instead", updatedProduct.Name, fetchedProduct.Name)
	}
}

func TestDeleteProductByID(t *testing.T) {
	mockProductID := uuid.New().String()
	mockProductStore := &utils.MockProductStore{
		Store: []models.Product{
			{ID: mockProductID, Name: "White Ring", Price: 22.99, Description: "Nice white ring", Category: "rings", StockQuantity: 2, CreatedAt: time.Now(), UpdatedAt: time.Now()},
			{ID: uuid.NewString(), Name: "Silver Necklace", Price: 10.99, Description: "Nice silver necklace", Category: "earings", StockQuantity: 2, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		},
	}
	testProductService := service.NewProductsService(mockProductStore)
	err := testProductService.DeleteProductByID(context.Background(), mockProductID)
	if err != nil {
		t.Fatalf("Failed to delete product: %v", err)
	}

	products, err := testProductService.GetProducts(context.Background())
	if err != nil {
		t.Fatalf("Failed to get products: %v", err)
	}
	if len(products) != 1 {
		t.Errorf("expected number of products was 1, got %d instead", len(products))
	}
	if products[0].Name != "Silver Necklace" {
		t.Errorf("expected 'Silver Necklace', got %s instead", products[0].Name)
	}
}

func TestUpdateStock(t *testing.T) {
	mockProductID := uuid.New().String()
	mockProductStore := &utils.MockProductStore{
		Store: []models.Product{
			{ID: mockProductID, Name: "White Ring", Price: 22.99, Description: "Nice white ring", Category: "rings", StockQuantity: 2, CreatedAt: time.Now(), UpdatedAt: time.Now()},
			{ID: uuid.NewString(), Name: "Silver Necklace", Price: 10.99, Description: "Nice silver necklace", Category: "earings", StockQuantity: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		},
	}
	testProductService := service.NewProductsService(mockProductStore)
	err := testProductService.UpdateStock(context.Background(), mockProductID, -2)
	if err != nil {
		t.Fatalf("Failed to delete product: %v", err)
	}

	product, err := testProductService.GetProductByID(context.Background(), mockProductID)
	if err != nil {
		t.Fatalf("Failed to get products: %v", err)
	}
	if product.StockQuantity != 0 {
		t.Errorf("expected 0 got %d instead", product.StockQuantity)
	}
}
