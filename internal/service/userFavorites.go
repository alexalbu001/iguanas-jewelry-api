package service

import (
	"context"
	"fmt"

	customerrors "github.com/alexalbu001/iguanas-jewelry-api/internal/customErrors"
	"github.com/alexalbu001/iguanas-jewelry-api/internal/models"
	"github.com/google/uuid"
)

type UserFavoritesStore interface {
	GetUserFavorites(ctx context.Context, userID string) ([]models.UserFavorites, error)
	AddUserFavorite(ctx context.Context, userID string, productID string) error
	RemoveUserFavorite(ctx context.Context, userID string, productID string) error
	ClearUserFavorites(ctx context.Context, userID string) error
}

type UserFavoritesService struct {
	UserFavoritesStore UserFavoritesStore
	ProductsStore      ProductsStore
}

func NewUserFavoritesService(userFavoritesStore UserFavoritesStore, productsStore ProductsStore) *UserFavoritesService {
	return &UserFavoritesService{
		UserFavoritesStore: userFavoritesStore,
		ProductsStore:      productsStore,
	}
}

func (u *UserFavoritesService) GetUserFavorites(ctx context.Context, userID string) ([]models.Product, error) {
	err := uuid.Validate(userID)
	if err != nil {
		return nil, &customerrors.ErrUserNotFound
	}
	userFavorites, err := u.UserFavoritesStore.GetUserFavorites(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("Error getting user favorites: %w", err)
	}

	var productIDs []string
	for _, userFavorite := range userFavorites {
		productIDs = append(productIDs, userFavorite.ProductID)
	}
	productMap, err := u.ProductsStore.GetByIDBatch(ctx, productIDs)
	if err != nil {
		return nil, fmt.Errorf("Error getting products by ids: %w", err)
	}
	var products []models.Product
	for _, product := range productMap {
		products = append(products, product)
	}
	return products, nil
}

func (u *UserFavoritesService) AddUserFavorite(ctx context.Context, userID string, productID string) error {
	err := uuid.Validate(userID)
	if err != nil {
		return &customerrors.ErrUserNotFound
	}
	err = uuid.Validate(productID)
	if err != nil {
		return &customerrors.ErrProductNotFound
	}

	_, err = u.ProductsStore.GetByID(ctx, productID)
	if err != nil {
		return &customerrors.ErrProductNotFound
	}

	return u.UserFavoritesStore.AddUserFavorite(ctx, userID, productID)
}

func (u *UserFavoritesService) RemoveUserFavorite(ctx context.Context, userID string, productID string) error {
	err := uuid.Validate(userID)
	if err != nil {
		return &customerrors.ErrUserNotFound
	}
	err = uuid.Validate(productID)
	if err != nil {
		return &customerrors.ErrProductNotFound
	}
	return u.UserFavoritesStore.RemoveUserFavorite(ctx, userID, productID)
}

func (u *UserFavoritesService) ClearUserFavorites(ctx context.Context, userID string) error {
	err := uuid.Validate(userID)
	if err != nil {
		return &customerrors.ErrUserNotFound
	}
	return u.UserFavoritesStore.ClearUserFavorites(ctx, userID)
}

// Can add some more methods to check which product is the most popular
