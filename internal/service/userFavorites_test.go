package service_test

import (
	"context"
	"fmt"
	"testing"

	customerrors "github.com/alexalbu001/iguanas-jewelry-api/internal/customErrors"
	"github.com/alexalbu001/iguanas-jewelry-api/internal/models"
	"github.com/alexalbu001/iguanas-jewelry-api/internal/service"
	"github.com/alexalbu001/iguanas-jewelry-api/internal/utils"
	"github.com/google/uuid"
)

func SetupFreshFavorites() *service.UserFavoritesService {
	return &service.UserFavoritesService{
		UserFavoritesStore: &utils.MockUserFavoritesStore{
			UserFavorites: utils.CreateTestFavorites(),
		},
		ProductsStore: &utils.MockProductStore{
			Store: utils.CreateJewelryProducts(),
		},
	}
}

func TestGetUserFavorites(t *testing.T) {
	userFavorites := SetupFreshFavorites()
	ctx := context.Background()
	products, err := userFavorites.GetUserFavorites(ctx, utils.KnownUserID)
	if err != nil {
		t.Fatalf("Failed to fetch user favorites: %v", err)
	}

	if len(products) != 2 {
		t.Fatalf("Number of products return should be 2 but is %d instead", len(products))
	}

	// Whatever else to test
}

func TestAddUserFavorite(t *testing.T) {
	userFavorites := SetupFreshFavorites()
	ctx := context.Background()

	err := userFavorites.AddUserFavorite(ctx, utils.KnownUserID, uuid.NewString())
	if err != &customerrors.ErrProductNotFound {
		t.Fatalf("Expected error response %v, got %v", &customerrors.ErrProductNotFound, err)
	}
}

func TestRemoveUserFavorite(t *testing.T) {
	userFavorites := SetupFreshFavorites()
	ctx := context.Background()

	err := userFavorites.RemoveUserFavorite(ctx, utils.KnownUserID, uuid.NewString())
	if err != &customerrors.ErrProductNotFound {
		t.Fatalf("Expected error response %v, got %v", &customerrors.ErrProductNotFound, err)
	}

	// Create a test product with UUID ID for this test
	testProductID := uuid.New().String()
	testProduct := models.Product{
		ID:            testProductID,
		Name:          "Test Product",
		Price:         100.00,
		Description:   "Test product for favorites",
		Category:      "rings",
		StockQuantity: 1,
	}

	// Add the test product to the mock store
	if mockStore, ok := userFavorites.ProductsStore.(*utils.MockProductStore); ok {
		mockStore.Store = append(mockStore.Store, testProduct)
	}

	// First add the favorite
	err = userFavorites.AddUserFavorite(ctx, utils.KnownUserID, testProductID)
	if err != nil {
		t.Fatalf("Failed to add user favorite: %v", err)
	}

	ring, err := userFavorites.ProductsStore.GetByID(ctx, testProductID)
	if err != nil {
		t.Fatalf("Failed to fetch product: %v", err)
	}
	fmt.Printf("Ring id %s, testProductId : %s", ring.ID, testProductID)

	err = userFavorites.RemoveUserFavorite(ctx, utils.KnownUserID, testProductID)
	if err != nil {
		t.Fatalf("Failed to remove user favorite: %v", err)
	}

	products, err := userFavorites.GetUserFavorites(ctx, utils.KnownUserID)
	if err != nil {
		t.Fatalf("Failed to fetch user favorites: %v", err)
	}

	if len(products) != 2 {
		t.Fatalf("Number of products return should be 2 but is %d instead", len(products))
	}
}
