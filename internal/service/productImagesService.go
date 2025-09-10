package service

import (
	"context"
	"fmt"

	customerrors "github.com/alexalbu001/iguanas-jewelry-api/internal/customErrors"
	"github.com/alexalbu001/iguanas-jewelry-api/internal/models"
	"github.com/alexalbu001/iguanas-jewelry-api/internal/storage"
	"github.com/alexalbu001/iguanas-jewelry-api/internal/transaction"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type ProductImagesStore interface {
	GetByProductID(ctx context.Context, productID string) ([]models.ProductImage, error)
	InsertProductImage(ctx context.Context, productImage models.ProductImage) (models.ProductImage, error)
	InsertProductImageBulk(ctx context.Context, productImages []models.ProductImage) error
	GetPrimaryImageForProduct(ctx context.Context, productID string) (models.ProductImage, error)
	GetPrimaryImageForProductBulk(ctx context.Context, productIDs []string) (map[string]models.ProductImage, error)
	UpdateProductImage(ctx context.Context, productImage models.ProductImage) error
	DeleteProductImage(ctx context.Context, imageID, productID string) error
	GetProductImagesBulk(ctx context.Context, productIDs []string) (map[string][]models.ProductImage, error)
	SetPrimaryImage(ctx context.Context, productID, imageID string) error
	SetPrimaryImageTx(ctx context.Context, productID, imageID string, tx pgx.Tx) error
	UnsetPrimaryImage(ctx context.Context, productID string) error
	UnsetPrimaryImageTx(ctx context.Context, productID string, tx pgx.Tx) error
	UpdateImageOrder(ctx context.Context, imageOrder []string) error
}

type ProductImagesService struct {
	ProductImagesStore ProductImagesStore
	UsersStore         UsersStore
	TxManager          transaction.TxManager
	ImageStorage       storage.ImageStorage
}

func NewProductImagesService(productImagesStore ProductImagesStore, usersStore UsersStore, imageStorage storage.ImageStorage, txManager transaction.TxManager) *ProductImagesService {
	return &ProductImagesService{
		ProductImagesStore: productImagesStore,
		UsersStore:         usersStore,
		ImageStorage:       imageStorage,
		TxManager:          txManager,
	}
}

func (p *ProductImagesService) GetProductImages(ctx context.Context, productID string) ([]models.ProductImage, error) {
	err := uuid.Validate(productID)
	if err != nil {
		return nil, &customerrors.ErrInvalidProductID
	}
	productImages, err := p.ProductImagesStore.GetByProductID(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("Error fetching product images: %w", err)
	}

	for i := range productImages {
		productImages[i].ImageURL = p.ImageStorage.GetImageURL(ctx, productImages[i].ImageKey)
	}
	return productImages, nil
}

func (p *ProductImagesService) GetProductImagesBulk(ctx context.Context, productIDs []string) (map[string][]models.ProductImage, error) {
	for _, productID := range productIDs {
		err := uuid.Validate(productID)
		if err != nil {
			return nil, &customerrors.ErrInvalidProductID
		}
	}
	productImagesMap, err := p.ProductImagesStore.GetProductImagesBulk(ctx, productIDs)
	if err != nil {
		return nil, fmt.Errorf("Error fetching product images: %w", err)
	}
	for _, productImages := range productImagesMap {
		for i := range productImages {
			productImages[i].ImageURL = p.ImageStorage.GetImageURL(ctx, productImages[i].ImageKey)
		}
	}
	return productImagesMap, nil
}

func (p *ProductImagesService) GetPrimaryImageForProduct(ctx context.Context, productID string) (models.ProductImage, error) {
	err := uuid.Validate(productID)
	if err != nil {
		return models.ProductImage{}, &customerrors.ErrInvalidProductID
	}
	productImage, err := p.ProductImagesStore.GetPrimaryImageForProduct(ctx, productID)
	if err != nil {
		return models.ProductImage{}, fmt.Errorf("Error fetching primary product image: %w", err)
	}
	productImage.ImageURL = p.ImageStorage.GetImageURL(ctx, productImage.ImageKey)
	return productImage, nil
}

func (p *ProductImagesService) InsertProductImage(ctx context.Context, productImage models.ProductImage, userID, productID string) (models.ProductImage, error) {
	if productID == "" {
		return models.ProductImage{}, &customerrors.ErrEmptyProductID
	}

	if productID != productImage.ProductID {
		return models.ProductImage{}, &customerrors.ErrInvalidProductID
	}

	err := uuid.Validate(productImage.ProductID)
	if err != nil {
		return models.ProductImage{}, &customerrors.ErrInvalidProductID
	}
	if productImage.ImageKey == "" {
		return models.ProductImage{}, &customerrors.ErrMissingImageURL
	}
	err = p.CheckIfUserIsAdmin(ctx, userID)
	if err != nil {
		return models.ProductImage{}, &customerrors.ErrUserUnauthorized
	}

	productImages, err := p.ProductImagesStore.GetByProductID(ctx, productImage.ProductID)
	if err != nil {
		return models.ProductImage{}, fmt.Errorf("Error fetching product images: %w", err)
	}
	if len(productImages) >= 5 {
		return models.ProductImage{}, &customerrors.ErrProductImagesLimitReached
	}

	if len(productImages) > 0 {

		productImage, err = p.ProductImagesStore.InsertProductImage(ctx, productImage)
		if err != nil {
			return models.ProductImage{}, fmt.Errorf("Error inserting product image: %w", err)
		}
		productImage.ImageURL = p.ImageStorage.GetImageURL(ctx, productImage.ImageKey)
		return productImage, nil
	} else {
		productImage.IsMain = true
		productImage, err = p.ProductImagesStore.InsertProductImage(ctx, productImage)
		if err != nil {
			return models.ProductImage{}, fmt.Errorf("Error inserting product image: %w", err)
		}
		productImage.ImageURL = p.ImageStorage.GetImageURL(ctx, productImage.ImageKey)
		return productImage, nil
	}
}

func (p *ProductImagesService) InsertProductImageBulk(ctx context.Context, productImages []models.ProductImage, userID string) error {
	for _, productImage := range productImages {
		if productImage.ProductID == "" {
			return &customerrors.ErrEmptyProductID
		}
		if productImage.ImageKey == "" {
			return &customerrors.ErrMissingImageURL
		}
		err := uuid.Validate(productImage.ProductID)
		if err != nil {
			return &customerrors.ErrInvalidProductID
		}
	}

	err := p.CheckIfUserIsAdmin(ctx, userID)
	if err != nil {
		return &customerrors.ErrUserUnauthorized
	}
	err = p.ProductImagesStore.InsertProductImageBulk(ctx, productImages)
	if err != nil {
		return fmt.Errorf("Error inserting product images: %w", err)
	}
	return nil
}

func (p *ProductImagesService) UpdateProductImage(ctx context.Context, productImage models.ProductImage, userID string) error {
	if productImage.ProductID == "" {
		return &customerrors.ErrEmptyProductID
	}
	err := uuid.Validate(productImage.ProductID)
	if err != nil {
		return &customerrors.ErrInvalidProductID
	}
	if productImage.ImageKey == "" {
		return &customerrors.ErrMissingImageURL
	}
	err = p.CheckIfUserIsAdmin(ctx, userID)
	if err != nil {
		return &customerrors.ErrUserUnauthorized
	}

	err = p.ProductImagesStore.UpdateProductImage(ctx, productImage)
	if err != nil {
		return fmt.Errorf("Error updating product image: %w", err)
	}
	return nil
}

func (p *ProductImagesService) DeleteProductImage(ctx context.Context, imageID, productID, userID string) error {
	if imageID == "" {
		return &customerrors.ErrEmptyImageURL
	}
	err := uuid.Validate(imageID)
	if err != nil {
		return &customerrors.ErrInvalidUUID
	}

	if productID == "" {
		return &customerrors.ErrEmptyProductID
	}
	err = uuid.Validate(productID)
	if err != nil {
		return &customerrors.ErrInvalidProductID
	}

	err = p.CheckIfUserIsAdmin(ctx, userID)
	if err != nil {
		return &customerrors.ErrUserUnauthorized
	}
	err = p.ProductImagesStore.DeleteProductImage(ctx, imageID, productID)
	if err != nil {
		return fmt.Errorf("Error deleting product image: %w", err)
	}
	return nil
}

func (p *ProductImagesService) GetPrimaryImageForProductsBulk(ctx context.Context, productIDs []string) (map[string]models.ProductImage, error) {
	for _, productID := range productIDs {
		err := uuid.Validate(productID)
		if err != nil {
			return nil, &customerrors.ErrInvalidProductID
		}
	}
	productImagesMap, err := p.ProductImagesStore.GetPrimaryImageForProductBulk(ctx, productIDs)
	if err != nil {
		return nil, fmt.Errorf("Error fetching primary product images: %w", err)
	}
	for productID, productImage := range productImagesMap {
		productImage.ImageURL = p.ImageStorage.GetImageURL(ctx, productImage.ImageKey)
		productImagesMap[productID] = productImage
	}
	return productImagesMap, nil
}

func (p *ProductImagesService) CheckIfUserIsAdmin(ctx context.Context, userID string) error {
	if userID == "" {
		return &customerrors.ErrUserNotFound
	}
	err := uuid.Validate(userID)
	if err != nil {
		return &customerrors.ErrUserNotFound
	}
	user, err := p.UsersStore.GetUserByID(ctx, userID)
	if err != nil {
		return &customerrors.ErrUserNotFound
	}
	if user.Role != "admin" {
		return &customerrors.ErrUserUnauthorized
	}
	return nil
}

func (p *ProductImagesService) SetPrimaryImage(ctx context.Context, productID, imageID, userID string) error {
	err := uuid.Validate(productID)
	if err != nil {
		return &customerrors.ErrInvalidProductID
	}

	err = uuid.Validate(imageID)
	if err != nil {
		return &customerrors.ErrInvalidUUID
	}

	err = p.CheckIfUserIsAdmin(ctx, userID)
	if err != nil {
		return &customerrors.ErrUserUnauthorized
	}
	err = p.TxManager.WithTransaction(ctx, func(tx pgx.Tx) error {
		err = p.ProductImagesStore.UnsetPrimaryImageTx(ctx, productID, tx)
		if err != nil {
			return err
		}

		err = p.ProductImagesStore.SetPrimaryImageTx(ctx, productID, imageID, tx)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("Transaction failed: %w", err)
	}
	return nil
}

func (p *ProductImagesService) ReorderImages(ctx context.Context, productID string, imageOrder []string, userID string) error {
	if productID == "" {
		return &customerrors.ErrEmptyProductID
	}
	err := uuid.Validate(productID)
	if err != nil {
		return &customerrors.ErrInvalidProductID
	}

	if len(imageOrder) == 0 {
		return &customerrors.ErrEmptyImageOrder
	}

	for _, imageID := range imageOrder {
		if err := uuid.Validate(imageID); err != nil {
			return &customerrors.ErrInvalidUUID
		}
	}

	err = p.CheckIfUserIsAdmin(ctx, userID)
	if err != nil {
		return &customerrors.ErrUserUnauthorized
	}

	err = p.ProductImagesStore.UpdateImageOrder(ctx, imageOrder)
	if err != nil {
		return err
	}

	return nil
}

func (p *ProductImagesService) GenerateUploadURL(ctx context.Context, productID, filename, contentType string) (string, string, error) {
	// Generate image key
	imageKey := fmt.Sprintf("products/%s/%s", productID, filename)

	// Generate upload URL
	uploadURL, err := p.ImageStorage.GenerateUploadURL(ctx, imageKey, contentType)
	if err != nil {
		return "", "", err
	}

	return uploadURL, imageKey, nil
}

func (p *ProductImagesService) ConfirmImageUpload(
	ctx context.Context,
	productID, imageKey string,
	isMain bool,
	userID string,
) (models.ProductImage, error) {
	// Validate inputs
	if err := uuid.Validate(productID); err != nil {
		return models.ProductImage{}, &customerrors.ErrInvalidProductID
	}

	if imageKey == "" {
		return models.ProductImage{}, &customerrors.ErrMissingImageURL
	}

	// Check if user is admin
	if err := p.CheckIfUserIsAdmin(ctx, userID); err != nil {
		return models.ProductImage{}, err
	}

	// Create ProductImage record
	productImage := models.ProductImage{
		ID:        uuid.New().String(),
		ProductID: productID,
		ImageKey:  imageKey,
		IsMain:    isMain,
		// DisplayOrder will be set by the store layer
	}

	// Save to database
	savedImage, err := p.ProductImagesStore.InsertProductImage(ctx, productImage)
	if err != nil {
		return models.ProductImage{}, fmt.Errorf("Error saving image metadata: %w", err)
	}

	// Generate the ImageURL for the response
	savedImage.ImageURL = p.ImageStorage.GetImageURL(ctx, savedImage.ImageKey)

	return savedImage, nil
}
