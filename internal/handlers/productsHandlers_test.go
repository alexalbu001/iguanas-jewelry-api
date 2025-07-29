package handlers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alexalbu001/iguanas-jewelry/internal/handlers"
	"github.com/alexalbu001/iguanas-jewelry/internal/models"
	"github.com/alexalbu001/iguanas-jewelry/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type MockProductStore struct {
	store []models.Product
}

func (m *MockProductStore) GetAll() ([]models.Product, error) {
	return m.store, nil
}

func (m *MockProductStore) GetByID(id string) (models.Product, error) {

	for _, product := range m.store {
		if id == product.ID {
			return product, nil
		}
	}
	return models.Product{}, fmt.Errorf("Product not found: %s", id)
}

func (m *MockProductStore) Add(product models.Product) (models.Product, error) {
	product.ID = uuid.NewString()
	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()

	m.store = append(m.store, product)
	return product, nil
}

func (m *MockProductStore) Update(id string, product models.Product) (models.Product, error) {
	for index, value := range m.store {
		if id == value.ID {
			product.ID = id
			product.CreatedAt = value.CreatedAt
			product.UpdatedAt = time.Now()
			m.store[index] = product
			return product, nil
		}
	}
	return models.Product{}, fmt.Errorf("ID: %s not found", id)
}

func (m *MockProductStore) Delete(id string) error {
	for i, value := range m.store {
		if value.ID == id {
			m.store = append(m.store[:i], m.store[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("ID: %s not found", id)
}

func TestGetProducts(t *testing.T) {
	mockStore := &MockProductStore{
		store: []models.Product{
			{ID: uuid.NewString(), Name: "Gold Ring", Price: 99.99, Description: "Nice gold ring", Category: "rings", CreatedAt: time.Now(), UpdatedAt: time.Now()},
			{ID: uuid.NewString(), Name: "Silver Necklace", Price: 10.99, Description: "Nice silver necklace", Category: "earings", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		},
	}

	testProductsService := service.NewProductsService(mockStore)
	testProductHandlers := handlers.NewProductHandlers(testProductsService)
	// Create fake HTTP response recorder
	w := httptest.NewRecorder()
	// Create fake Gin context
	c, _ := gin.CreateTestContext(w)

	testProductHandlers.GetProducts(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected %d got %d instead", http.StatusOK, w.Code)
	}
	var products []models.Product
	err := json.Unmarshal(w.Body.Bytes(), &products)
	if err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if len(products) != 2 {
		t.Errorf("expected 2 arguments, got %d instead", len(products))
	}

	if products[0].Name != "Gold Ring" {
		t.Errorf("Expected first product name 'Gold Ring', got %s", products[0].Name)
	}
}

func TestPostProduct(t *testing.T) {
	product := models.Product{
		Name: "Gold Ring", Price: 99.99, Description: "Nice gold ring", Category: "rings",
	}

	mockStore := &MockProductStore{
		store: []models.Product{
			{ID: uuid.NewString(), Name: "White Ring", Price: 22.99, Description: "Nice white ring", Category: "rings", CreatedAt: time.Now(), UpdatedAt: time.Now()},
			{ID: uuid.NewString(), Name: "Silver Necklace", Price: 10.99, Description: "Nice silver necklace", Category: "earings", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		},
	}
	// Convert product to JSON
	jsonBody, err := json.Marshal(product)
	if err != nil {
		t.Fatalf("failed to marshall products into JSON: %v", err)
	}

	req := httptest.NewRequest("POST", "/products", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	testProductsService := service.NewProductsService(mockStore)
	testProductHandlers := handlers.NewProductHandlers(testProductsService)
	// Create fake HTTP response recorder
	w := httptest.NewRecorder()
	// Create fake Gin context
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	testProductHandlers.PostProducts(c)
	if w.Code != http.StatusCreated {
		t.Errorf("expected %d got %d instead", http.StatusCreated, w.Code)
	}
	newW := httptest.NewRecorder()
	c, _ = gin.CreateTestContext(newW)
	testProductHandlers.GetProducts(c)

	var products []models.Product
	err = json.Unmarshal(newW.Body.Bytes(), &products)
	if err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if len(products) != 3 {
		t.Errorf("expected 3 arguments, got %d instead", len(products))
	}

	if products[2].Name != "Gold Ring" {
		t.Errorf("Expected first product name 'Gold Ring', got %s", products[2].Name)
	}
}

func TestGetProductByID(t *testing.T) {
	mockStore := &MockProductStore{
		store: []models.Product{
			{ID: "known-id", Name: "White Ring", Price: 22.99, Description: "Nice white ring", Category: "rings", CreatedAt: time.Now(), UpdatedAt: time.Now()},
			{ID: uuid.NewString(), Name: "Silver Necklace", Price: 10.99, Description: "Nice silver necklace", Category: "earings", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		},
	}
	req := httptest.NewRequest("GET", "/products/known-id", nil)

	testProductsService := service.NewProductsService(mockStore)
	testProductHandlers := handlers.NewProductHandlers(testProductsService)
	// Create fake HTTP response recorder
	w := httptest.NewRecorder()
	// Create fake Gin context
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	testProductHandlers.GetProductByID(c)

	c.Params = []gin.Param{ //this sets the URL param context in gin for the test
		{Key: "id", Value: "known-id"},
	}
	var product *models.Product
	err := json.Unmarshal(w.Body.Bytes(), &product)
	if err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if product.ID != "known-id" {
		t.Fatalf("expected known-id, got %s", product.ID)
	}
	if w.Code != http.StatusOK {
		t.Errorf("expected %d got %d instead", http.StatusOK, w.Code)
	}
}

func TestUpdateProductByID(t *testing.T) {
	mockStore := &MockProductStore{
		store: []models.Product{
			{ID: "known-id", Name: "White Ring", Price: 22.99, Description: "Nice white ring", Category: "rings", CreatedAt: time.Now(), UpdatedAt: time.Now()},
			{ID: uuid.NewString(), Name: "Silver Necklace", Price: 10.99, Description: "Nice silver necklace", Category: "earings", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		},
	}

	product := models.Product{
		Name: "Gold Ring", Price: 99.99, Description: "Nice gold ring", Category: "rings",
	}
	jsonBody, err := json.Marshal(product)
	if err != nil {
		t.Fatalf("failed to marshall products into JSON: %v", err)
	}

	testProductsService := service.NewProductsService(mockStore)
	testProductHandlers := handlers.NewProductHandlers(testProductsService)
	req := httptest.NewRequest("PUT", "/products/known-id", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = []gin.Param{ //this sets the URL param context in gin for the test
		{Key: "id", Value: "known-id"},
	}
	testProductHandlers.UpdateProductByID(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected %d got %d instead", http.StatusOK, w.Code)
	}

	newReq := httptest.NewRequest("GET", "/products/known-id", nil)
	newW := httptest.NewRecorder()
	c, _ = gin.CreateTestContext(newW)
	c.Request = newReq
	c.Params = []gin.Param{ //this sets the URL param context in gin for the test
		{Key: "id", Value: "known-id"},
	}
	testProductHandlers.GetProductByID(c)

	var newProduct models.Product
	err = json.Unmarshal(newW.Body.Bytes(), &newProduct)
	if err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if newProduct.Name != product.Name {
		t.Fatalf("expected name %s, go %s instead", product.Name, newProduct.Name)
	}
}

func TestDeleteProductByID(t *testing.T) {
	mockStore := &MockProductStore{
		store: []models.Product{
			{ID: "known-id", Name: "White Ring", Price: 22.99, Description: "Nice white ring", Category: "rings", StockQuantity: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()},
			{ID: uuid.NewString(), Name: "Silver Necklace", Price: 10.99, Description: "Nice silver necklace", Category: "earings", StockQuantity: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		},
	}

	testProductsService := service.NewProductsService(mockStore)
	testProductHandlers := handlers.NewProductHandlers(testProductsService)

	req := httptest.NewRequest("DELETE", "/products/:id", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = []gin.Param{
		{Key: "id", Value: "known-id"},
	}
	testProductHandlers.DeleteProductByID(c)

	newReq := httptest.NewRequest("GET", "/products", nil)

	newW := httptest.NewRecorder()
	c, _ = gin.CreateTestContext(newW)
	c.Request = newReq
	testProductHandlers.GetProducts(c)

	var products []models.Product
	err := json.Unmarshal(newW.Body.Bytes(), &products)
	if err != nil {
		t.Fatalf("error parsing JSON, %v", err)
	}
	if len(products) != 1 {
		t.Errorf("expected number of products was 1, got %d instead", len(products))
	}

	if products[0].Name != "Silver Necklace" {
		t.Errorf("expected 'Silver Necklace', got %s instead", products[0].Name)
	}
}
