package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/alexalbu001/iguanas-jewelry-api/internal/models"
	"github.com/alexalbu001/iguanas-jewelry-api/internal/service"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockProductImagesStore is a mock implementation of ProductImagesStore
type MockProductImagesStore struct {
	mock.Mock
}

func (m *MockProductImagesStore) GetByProductID(ctx context.Context, productID string) ([]models.ProductImage, error) {
	args := m.Called(ctx, productID)
	return args.Get(0).([]models.ProductImage), args.Error(1)
}

func (m *MockProductImagesStore) InsertProductImage(ctx context.Context, productImage models.ProductImage) (models.ProductImage, error) {
	args := m.Called(ctx, productImage)
	return args.Get(0).(models.ProductImage), args.Error(1)
}

func (m *MockProductImagesStore) InsertProductImageBulk(ctx context.Context, productImages []models.ProductImage) error {
	args := m.Called(ctx, productImages)
	return args.Error(0)
}

func (m *MockProductImagesStore) GetPrimaryImageForProduct(ctx context.Context, productID string) (models.ProductImage, error) {
	args := m.Called(ctx, productID)
	return args.Get(0).(models.ProductImage), args.Error(1)
}

func (m *MockProductImagesStore) GetPrimaryImageForProductBulk(ctx context.Context, productIDs []string) (map[string]models.ProductImage, error) {
	args := m.Called(ctx, productIDs)
	return args.Get(0).(map[string]models.ProductImage), args.Error(1)
}

func (m *MockProductImagesStore) UpdateProductImage(ctx context.Context, productImage models.ProductImage) error {
	args := m.Called(ctx, productImage)
	return args.Error(0)
}

func (m *MockProductImagesStore) DeleteProductImage(ctx context.Context, imageID, productID string) error {
	args := m.Called(ctx, imageID, productID)
	return args.Error(0)
}

func (m *MockProductImagesStore) GetProductImagesBulk(ctx context.Context, productIDs []string) (map[string][]models.ProductImage, error) {
	args := m.Called(ctx, productIDs)
	return args.Get(0).(map[string][]models.ProductImage), args.Error(1)
}

func (m *MockProductImagesStore) SetPrimaryImage(ctx context.Context, productID, imageID string) error {
	args := m.Called(ctx, productID, imageID)
	return args.Error(0)
}

func (m *MockProductImagesStore) SetPrimaryImageTx(ctx context.Context, productID, imageID string, tx pgx.Tx) error {
	args := m.Called(ctx, productID, imageID, tx)
	return args.Error(0)
}

func (m *MockProductImagesStore) UnsetPrimaryImage(ctx context.Context, productID string) error {
	args := m.Called(ctx, productID)
	return args.Error(0)
}

func (m *MockProductImagesStore) UnsetPrimaryImageTx(ctx context.Context, productID string, tx pgx.Tx) error {
	args := m.Called(ctx, productID, tx)
	return args.Error(0)
}

func (m *MockProductImagesStore) UpdateImageOrder(ctx context.Context, imageOrder []string) error {
	args := m.Called(ctx, imageOrder)
	return args.Error(0)
}

// MockImageStorage is a mock implementation of storage.ImageStorage
type MockImageStorage struct {
	mock.Mock
}

func (m *MockImageStorage) GenerateUploadURL(ctx context.Context, key string, contentType string) (string, error) {
	args := m.Called(ctx, key, contentType)
	return args.String(0), args.Error(1)
}

func (m *MockImageStorage) GetImageURL(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockImageStorage) ProcessUploadComplete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockImageStorage) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockImageStorage) Store(ctx context.Context, key string, data []byte) error {
	args := m.Called(ctx, key, data)
	return args.Error(0)
}

func (m *MockImageStorage) Get(ctx context.Context, key string) ([]byte, error) {
	args := m.Called(ctx, key)
	return args.Get(0).([]byte), args.Error(1)
}

// MockTxManager is a mock implementation of transaction.TxManager
type MockTxManager struct {
	mock.Mock
}

func (m *MockTxManager) WithTransaction(ctx context.Context, fn func(tx pgx.Tx) error) error {
	args := m.Called(ctx, fn)
	return args.Error(0)
}

func TestProductImagesService_GetProductImages(t *testing.T) {
	tests := []struct {
		name          string
		productID     string
		mockImages    []models.ProductImage
		mockError     error
		expectedError bool
	}{
		{
			name:      "successful image retrieval",
			productID: uuid.New().String(),
			mockImages: []models.ProductImage{
				{
					ID:        uuid.New().String(),
					ProductID: uuid.New().String(),
					ImageKey:  "test-image-1.jpg",
					IsMain:    true,
				},
				{
					ID:        uuid.New().String(),
					ProductID: uuid.New().String(),
					ImageKey:  "test-image-2.jpg",
					IsMain:    false,
				},
			},
			mockError:     nil,
			expectedError: false,
		},
		{
			name:          "invalid product ID",
			productID:     "invalid-uuid",
			mockImages:    nil,
			mockError:     nil,
			expectedError: true,
		},
		{
			name:          "database error",
			productID:     uuid.New().String(),
			mockImages:    nil,
			mockError:     errors.New("database error"),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := new(MockProductImagesStore)
			mockUsersStore := new(MockUsersStore)
			mockImageStorage := new(MockImageStorage)
			mockTxManager := new(MockTxManager)

			if !tt.expectedError || tt.mockError != nil {
				mockStore.On("GetByProductID", mock.Anything, tt.productID).Return(tt.mockImages, tt.mockError)

				// Mock image storage calls for each image
				for i := range tt.mockImages {
					mockImageStorage.On("GetImageURL", mock.Anything, tt.mockImages[i].ImageKey).Return("https://example.com/"+tt.mockImages[i].ImageKey, nil)
				}
			}

			service := service.NewProductImagesService(mockStore, mockUsersStore, mockImageStorage, mockTxManager)

			result, err := service.GetProductImages(context.Background(), tt.productID)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(tt.mockImages), len(result))
				// Check that ImageURL is set
				for i, img := range result {
					assert.NotEmpty(t, img.ImageURL)
					assert.Equal(t, tt.mockImages[i].ImageKey, img.ImageKey)
				}
			}

			mockStore.AssertExpectations(t)
			mockImageStorage.AssertExpectations(t)
		})
	}
}

func TestProductImagesService_InsertProductImage(t *testing.T) {
	validProductID := uuid.New().String()
	validUserID := uuid.New().String()

	tests := []struct {
		name          string
		productImage  models.ProductImage
		userID        string
		productID     string
		mockImages    []models.ProductImage
		mockError     error
		expectedError bool
	}{
		{
			name: "successful image insertion - first image (becomes main)",
			productImage: models.ProductImage{
				ProductID: validProductID,
				ImageKey:  "test-image.jpg",
				IsMain:    false, // Will be set to true for first image
			},
			userID:        validUserID,
			productID:     validProductID,
			mockImages:    []models.ProductImage{}, // No existing images
			mockError:     nil,
			expectedError: false,
		},
		{
			name: "successful image insertion - additional image",
			productImage: models.ProductImage{
				ProductID: validProductID,
				ImageKey:  "test-image-2.jpg",
				IsMain:    false,
			},
			userID:    validUserID,
			productID: validProductID,
			mockImages: []models.ProductImage{
				{ID: uuid.New().String(), ProductID: validProductID, ImageKey: "existing.jpg", IsMain: true},
			},
			mockError:     nil,
			expectedError: false,
		},
		{
			name: "empty product ID",
			productImage: models.ProductImage{
				ProductID: "",
				ImageKey:  "test-image.jpg",
			},
			userID:        validUserID,
			productID:     "",
			mockImages:    nil,
			mockError:     nil,
			expectedError: true,
		},
		{
			name: "invalid product ID",
			productImage: models.ProductImage{
				ProductID: "invalid-uuid",
				ImageKey:  "test-image.jpg",
			},
			userID:        validUserID,
			productID:     "invalid-uuid",
			mockImages:    nil,
			mockError:     nil,
			expectedError: true,
		},
		{
			name: "missing image key",
			productImage: models.ProductImage{
				ProductID: validProductID,
				ImageKey:  "",
			},
			userID:        validUserID,
			productID:     validProductID,
			mockImages:    nil,
			mockError:     nil,
			expectedError: true,
		},
		{
			name: "product ID mismatch",
			productImage: models.ProductImage{
				ProductID: "different-product-id",
				ImageKey:  "test-image.jpg",
			},
			userID:        validUserID,
			productID:     validProductID,
			mockImages:    nil,
			mockError:     nil,
			expectedError: true,
		},
		{
			name: "user not admin",
			productImage: models.ProductImage{
				ProductID: validProductID,
				ImageKey:  "test-image.jpg",
			},
			userID:        uuid.New().String(), // Use valid UUID
			productID:     validProductID,
			mockImages:    nil,
			mockError:     nil,
			expectedError: true,
		},
		{
			name: "image limit reached",
			productImage: models.ProductImage{
				ProductID: validProductID,
				ImageKey:  "test-image.jpg",
			},
			userID:    validUserID,
			productID: validProductID,
			mockImages: []models.ProductImage{
				{ID: uuid.New().String(), ProductID: validProductID, ImageKey: "img1.jpg"},
				{ID: uuid.New().String(), ProductID: validProductID, ImageKey: "img2.jpg"},
				{ID: uuid.New().String(), ProductID: validProductID, ImageKey: "img3.jpg"},
				{ID: uuid.New().String(), ProductID: validProductID, ImageKey: "img4.jpg"},
				{ID: uuid.New().String(), ProductID: validProductID, ImageKey: "img5.jpg"},
			},
			mockError:     nil,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := new(MockProductImagesStore)
			mockUsersStore := new(MockUsersStore)
			mockImageStorage := new(MockImageStorage)
			mockTxManager := new(MockTxManager)

			// Set up mocks based on test case
			if tt.expectedError {
				if tt.productImage.ProductID == "" {
					// Empty product ID case
				} else if tt.productImage.ProductID == "invalid-uuid" {
					// Invalid UUID case
				} else if tt.productImage.ImageKey == "" {
					// Missing image key case
				} else if tt.productImage.ProductID != tt.productID {
					// Product ID mismatch case
				} else if tt.name == "user not admin" {
					// Non-admin user case
					mockUsersStore.On("GetUserByID", mock.Anything, tt.userID).Return(models.User{Role: "customer"}, nil)
				} else if len(tt.mockImages) >= 5 {
					// Image limit reached case
					mockUsersStore.On("GetUserByID", mock.Anything, tt.userID).Return(models.User{Role: "admin"}, nil)
					mockStore.On("GetByProductID", mock.Anything, tt.productID).Return(tt.mockImages, nil)
				}
			} else {
				// Success case
				mockUsersStore.On("GetUserByID", mock.Anything, tt.userID).Return(models.User{Role: "admin"}, nil)
				mockStore.On("GetByProductID", mock.Anything, tt.productID).Return(tt.mockImages, nil)

				// Mock the insert call
				expectedImage := tt.productImage
				if len(tt.mockImages) == 0 {
					expectedImage.IsMain = true // First image becomes main
				}
				expectedImage.ID = uuid.New().String() // Store will assign ID

				mockStore.On("InsertProductImage", mock.Anything, mock.MatchedBy(func(img models.ProductImage) bool {
					return img.ProductID == expectedImage.ProductID && img.ImageKey == expectedImage.ImageKey
				})).Return(expectedImage, nil)

				mockImageStorage.On("GetImageURL", mock.Anything, expectedImage.ImageKey).Return("https://example.com/"+expectedImage.ImageKey, nil)
			}

			service := service.NewProductImagesService(mockStore, mockUsersStore, mockImageStorage, mockTxManager)

			result, err := service.InsertProductImage(context.Background(), tt.productImage, tt.userID, tt.productID)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Equal(t, models.ProductImage{}, result)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, result.ID)
				assert.Equal(t, tt.productImage.ProductID, result.ProductID)
				assert.Equal(t, tt.productImage.ImageKey, result.ImageKey)
				assert.NotEmpty(t, result.ImageURL)
			}

			mockStore.AssertExpectations(t)
			mockUsersStore.AssertExpectations(t)
			mockImageStorage.AssertExpectations(t)
		})
	}
}

func TestProductImagesService_CheckIfUserIsAdmin(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		mockUser      models.User
		mockError     error
		expectedError bool
	}{
		{
			name:          "admin user",
			userID:        uuid.New().String(),
			mockUser:      models.User{ID: uuid.New().String(), Role: "admin"},
			mockError:     nil,
			expectedError: false,
		},
		{
			name:          "customer user",
			userID:        uuid.New().String(),
			mockUser:      models.User{ID: uuid.New().String(), Role: "customer"},
			mockError:     nil,
			expectedError: true,
		},
		{
			name:          "empty user ID",
			userID:        "",
			mockUser:      models.User{},
			mockError:     nil,
			expectedError: true,
		},
		{
			name:          "invalid user ID",
			userID:        "invalid-uuid",
			mockUser:      models.User{},
			mockError:     nil,
			expectedError: true,
		},
		{
			name:          "user not found",
			userID:        uuid.New().String(),
			mockUser:      models.User{},
			mockError:     errors.New("user not found"),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := new(MockProductImagesStore)
			mockUsersStore := new(MockUsersStore)
			mockImageStorage := new(MockImageStorage)
			mockTxManager := new(MockTxManager)

			if tt.userID != "" && tt.userID != "invalid-uuid" {
				mockUsersStore.On("GetUserByID", mock.Anything, tt.userID).Return(tt.mockUser, tt.mockError)
			}

			service := service.NewProductImagesService(mockStore, mockUsersStore, mockImageStorage, mockTxManager)

			err := service.CheckIfUserIsAdmin(context.Background(), tt.userID)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockUsersStore.AssertExpectations(t)
		})
	}
}

func TestProductImagesService_GenerateUploadURL(t *testing.T) {
	tests := []struct {
		name          string
		productID     string
		filename      string
		contentType   string
		mockURL       string
		mockError     error
		expectedError bool
	}{
		{
			name:          "successful URL generation",
			productID:     uuid.New().String(),
			filename:      "test-image.jpg",
			contentType:   "image/jpeg",
			mockURL:       "https://example.com/upload-url",
			mockError:     nil,
			expectedError: false,
		},
		{
			name:          "storage error",
			productID:     uuid.New().String(),
			filename:      "test-image.jpg",
			contentType:   "image/jpeg",
			mockURL:       "",
			mockError:     errors.New("storage error"),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := new(MockProductImagesStore)
			mockUsersStore := new(MockUsersStore)
			mockImageStorage := new(MockImageStorage)
			mockTxManager := new(MockTxManager)

			expectedKey := "products/" + tt.productID + "/" + tt.filename
			mockImageStorage.On("GenerateUploadURL", mock.Anything, expectedKey, tt.contentType).Return(tt.mockURL, tt.mockError)

			service := service.NewProductImagesService(mockStore, mockUsersStore, mockImageStorage, mockTxManager)

			url, key, err := service.GenerateUploadURL(context.Background(), tt.productID, tt.filename, tt.contentType)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Empty(t, url)
				assert.Empty(t, key)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.mockURL, url)
				assert.Equal(t, expectedKey, key)
			}

			mockImageStorage.AssertExpectations(t)
		})
	}
}
