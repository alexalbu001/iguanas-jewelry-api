package store_test

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/alexalbu001/iguanas-jewelry/internal/models"
	"github.com/alexalbu001/iguanas-jewelry/internal/store"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// func TestGetByID(t *testing.T) {
// 	tx, err := dbpool.Begin
// }

var testDB *pgxpool.Pool
var testStore *store.ProductsStore

func TestMain(m *testing.M) {
	var err error
	testDB, err = pgxpool.New(context.Background(), os.Getenv("TEST_DATABASE_URL"))
	if err != nil {
		log.Fatal("Unable to connect to db", err)
	}
	testStore = store.NewProductStore(testDB)
	// testHandlers := &handlers.ProductHandlers{
	// 	Store: testStore,
	// }

	defer testDB.Close()
	os.Exit(m.Run())
}

func TestGetByID(t *testing.T) {
	// tx, err := testDB.Begin(context.Background())
	// if err != nil {
	// 	t.Fatalf("Error beginning transaction: w%", err)
	// }
	// defer tx.Rollback(context.Background())
	testID := "test-" + uuid.NewString()
	_, err := testDB.Exec(context.Background(), `
	INSERT INTO products (id, name, price, description, category, created_at, updated_at)
	VALUES ($1, 'Gold Ring', 5.00, 'This is a description', 'Rings', '2025-06-11T12:28:29.914144Z', '2025-06-11T12:28:29.914144Z')
	`, testID)
	if err != nil {
		t.Fatalf("Error executing sql: %v", err) // %w is for fmt.Errorf
	}
	defer testDB.Exec(context.Background(),
		"DELETE FROM products WHERE id=$1", testID)

	result, err := testStore.GetByID(testID)
	if err != nil {
		t.Fatalf("Error running GetByID: %v", err)
	}

	if result.Name != "Gold Ring" {
		t.Errorf("Expected name 'Gold Ring', got %s", result.Name)
	}
	if result.Price != 5.00 {
		t.Errorf("Expected price was 5, got %2f", result.Price)
	}

	if result.Description != "This is a description" {
		t.Errorf("Expected description 'This is a description', got %s", result.Description)
	}
}

func createTestProduct(t *testing.T, name string, price float64) string {
	testID := "test" + uuid.NewString()
	_, err := testDB.Exec(context.Background(), `
	INSERT INTO products (id, name, price, description, category, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
	`, testID, name, price, "Test description", "Rings")
	if err != nil {
		t.Fatalf("Error executing sql: %v", err) // %w is for fmt.Errorf
	}
	return testID
}

func TestGetAll(t *testing.T) {
	empty, err := testStore.GetAll()
	if err != nil {
		t.Fatalf("Error running GetAll function %v", err)
	}
	if len(empty) != 0 {
		t.Errorf("Expected empty slice, got %d instead", len(empty))
	}

	Product1 := createTestProduct(t, "Test1", 12.3)
	Product2 := createTestProduct(t, "Test2", 13.00)
	Product3 := createTestProduct(t, "Test3", 14)

	// Clean up
	defer testDB.Exec(context.Background(), "DELETE FROM products WHERE id IN ($1, $2, $3)", Product1, Product2, Product3)

	products, err := testStore.GetAll()
	if err != nil {
		t.Fatalf("Error running GetAll function %v", err)
	}

	if len(products) != 3 {
		t.Errorf("Expected 3 products, got %d instead", len(products))
	}
	if products[0].Name != "Test3" {
		t.Errorf("Expected newest product first, got %s instead", products[0].Name)
	}

	for _, v := range products {
		if v.Price <= 0 {
			t.Errorf("Invalid price, %2f", v.Price)
		}
		if v.Name == "" {
			t.Errorf("Empty product name")
		}
	}
}

func TestAdd(t *testing.T) {

	var testProduct = models.Product{
		Name:        "Test",
		Price:       5,
		Description: "Test Description",
		Category:    "rings",
	}

	product, err := testStore.Add(testProduct)
	if err != nil {
		t.Fatalf("Error running Add: %v", err)
	}
	defer testDB.Exec(context.Background(), "DELETE FROM products WHERE id=$1", product.ID)
	dbProduct, err := testStore.GetByID(product.ID)
	if err != nil {
		t.Fatalf("Error executing GetByID %v", err)
	}
	if product.ID != dbProduct.ID {
		t.Errorf("Expected %s, got %s instead", product.ID, dbProduct.ID)
	}
	if dbProduct.Name != "Test" {
		t.Errorf("Expected product name inside DB to be 'Test', got %s instead", dbProduct.Name)
	}
	if dbProduct.Price != 5 {
		t.Errorf("Expected product price inside DB to be 5, got %2f instead", dbProduct.Price)
	}
}

func TestDelete(t *testing.T) {

	Product1 := createTestProduct(t, "Test1", 12.3)
	Product2 := createTestProduct(t, "Test2", 13.00)
	Product3 := createTestProduct(t, "Test3", 14)
	defer testDB.Exec(context.Background(), "DELETE FROM products WHERE id IN ($1, $2, $3)", Product1, Product2, Product3)

	err := testStore.Delete(Product3)
	if err != nil {
		t.Fatalf("Error executing Delete %v", err)
	}

	products, err := testStore.GetAll()
	if err != nil {
		t.Fatalf("Error running GetAll function %v", err)
	}
	if len(products) != 2 {
		t.Errorf("Expected 2 products, got %d instead", len(products))
	}
	if products[0].Name != "Test2" {
		t.Errorf("Expected first product name inside DB to be 'Test2', got %s instead", products[0].Name)
	}
	_, err = testStore.GetByID(Product3)
	if err == nil {
		t.Errorf("Expected error, got nothing")
	}

	emptyId := ""
	shouldError := testStore.Delete(emptyId)
	if shouldError == nil {
		t.Errorf("Expected error when deleting with empty ID, got nil")
	}
}
