package store_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/alexalbu001/iguanas-jewelry/internal/store"
	"github.com/go-playground/locales/sd"
	"github.com/jackc/pgx/v5/pgxpool"
)

// func TestGetByID(t *testing.T) {
// 	tx, err := dbpool.Begin
// }

var testDB *pgxpool.Pool
var testStore *ProductsStore

func TestMain(m *testing.M) {
	testDB, err := pgxpool.New(context.Background(), os.Getenv("TEST_DATABASE_URL"))
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
	tx, err := testDB.Begin(context.Background())
	if err != nil {
		t.Fatalf("Error beginning transaction: w%", err)
	}
	defer tx.Rollback(context.Background())
	_, err := tx.Exec(context.Background(), `
	INSERT INTO products (id, name, price, description, category, created_at, updated_at)
	VALUES ('326066b5-5bbc-4286-843e-24306935882e', 'Gold Ring', 5, 'This is a description', 'Rings', '2025-06-11T12:28:29.914144Z', '2025-06-11T12:28:29.914144Z')
	`)
	if err!= nil{
		t.Fatalf("Error executing sql: %w", err)
	}

	result, err := testStore.GetByID("326066b5-5bbc-4286-843e-24306935882e")
	if err != nil {
		t.Fatalf("Error running GetByID: %w", err)
	}
	if result == 

}
