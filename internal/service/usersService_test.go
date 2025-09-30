package service_test

import (
	"context"
	"errors"
	"testing"

	customerrors "github.com/alexalbu001/iguanas-jewelry-api/internal/customErrors"
	"github.com/alexalbu001/iguanas-jewelry-api/internal/models"
	"github.com/alexalbu001/iguanas-jewelry-api/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUsersStore is a mock implementation of UsersStore
type MockUsersStore struct {
	mock.Mock
}

func (m *MockUsersStore) GetUsers(ctx context.Context) ([]models.User, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *MockUsersStore) GetUserByGoogleID(ctx context.Context, googleID string) (models.User, error) {
	args := m.Called(ctx, googleID)
	return args.Get(0).(models.User), args.Error(1)
}

func (m *MockUsersStore) GetUserByID(ctx context.Context, id string) (models.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(models.User), args.Error(1)
}

func (m *MockUsersStore) AddUser(ctx context.Context, user models.User) (models.User, error) {
	args := m.Called(ctx, user)
	return args.Get(0).(models.User), args.Error(1)
}

func (m *MockUsersStore) DeleteUser(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUsersStore) UpdateUser(ctx context.Context, id string, user models.User) (models.User, error) {
	args := m.Called(ctx, id, user)
	return args.Get(0).(models.User), args.Error(1)
}

func (m *MockUsersStore) UpdateUserRole(ctx context.Context, id string, role string) error {
	args := m.Called(ctx, id, role)
	return args.Error(0)
}

func TestUserService_GetUsers(t *testing.T) {
	tests := []struct {
		name          string
		mockUsers     []models.User
		mockError     error
		expectedError bool
	}{
		{
			name: "successful users retrieval",
			mockUsers: []models.User{
				{
					ID:    uuid.New().String(),
					Name:  "John Doe",
					Email: "john@example.com",
					Role:  "customer",
				},
				{
					ID:    uuid.New().String(),
					Name:  "Jane Smith",
					Email: "jane@example.com",
					Role:  "admin",
				},
			},
			mockError:     nil,
			expectedError: false,
		},
		{
			name:          "database error",
			mockUsers:     nil,
			mockError:     errors.New("database error"),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := new(MockUsersStore)
			mockStore.On("GetUsers", mock.Anything).Return(tt.mockUsers, tt.mockError)

			userService := service.NewUserService(mockStore)

			result, err := userService.GetUsers(context.Background())

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.mockUsers, result)
			}

			mockStore.AssertExpectations(t)
		})
	}
}

func TestUserService_GetUserByID(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		mockUser      models.User
		mockError     error
		expectedError bool
	}{
		{
			name:   "successful user retrieval",
			userID: uuid.New().String(),
			mockUser: models.User{
				ID:    uuid.New().String(),
				Name:  "John Doe",
				Email: "john@example.com",
				Role:  "customer",
			},
			mockError:     nil,
			expectedError: false,
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
			mockStore := new(MockUsersStore)
			mockStore.On("GetUserByID", mock.Anything, tt.userID).Return(tt.mockUser, tt.mockError)

			userService := service.NewUserService(mockStore)

			result, err := userService.GetUserByID(context.Background(), tt.userID)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Equal(t, models.User{}, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.mockUser, result)
			}

			mockStore.AssertExpectations(t)
		})
	}
}

func TestUserService_UpdateUserByID(t *testing.T) {
	validUserID := uuid.New().String()

	tests := []struct {
		name          string
		userID        string
		user          models.User
		mockUser      models.User
		mockError     error
		expectedError bool
	}{
		{
			name:   "successful user update",
			userID: validUserID,
			user: models.User{
				Name:  "John Updated",
				Email: "john.updated@example.com",
			},
			mockUser: models.User{
				ID:    validUserID,
				Name:  "John Updated",
				Email: "john.updated@example.com",
				Role:  "customer",
			},
			mockError:     nil,
			expectedError: false,
		},
		{
			name:   "invalid user ID",
			userID: "invalid-uuid",
			user: models.User{
				Name:  "John Updated",
				Email: "john.updated@example.com",
			},
			mockUser:      models.User{},
			mockError:     nil,
			expectedError: true,
		},
		{
			name:   "missing name",
			userID: validUserID,
			user: models.User{
				Name:  "",
				Email: "john.updated@example.com",
			},
			mockUser:      models.User{},
			mockError:     nil,
			expectedError: true,
		},
		{
			name:   "database error",
			userID: validUserID,
			user: models.User{
				Name:  "John Updated",
				Email: "john.updated@example.com",
			},
			mockUser:      models.User{},
			mockError:     errors.New("database error"),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := new(MockUsersStore)

			if !tt.expectedError || tt.mockError != nil {
				mockStore.On("UpdateUser", mock.Anything, tt.userID, tt.user).Return(tt.mockUser, tt.mockError)
			}

			userService := service.NewUserService(mockStore)

			result, err := userService.UpdateUserByID(context.Background(), tt.userID, tt.user)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Equal(t, models.User{}, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.mockUser, result)
			}

			mockStore.AssertExpectations(t)
		})
	}
}

func TestUserService_UpdateUserRole(t *testing.T) {
	validUserID := uuid.New().String()

	tests := []struct {
		name          string
		userID        string
		role          string
		mockError     error
		expectedError bool
	}{
		{
			name:          "successful role update to admin",
			userID:        validUserID,
			role:          "admin",
			mockError:     nil,
			expectedError: false,
		},
		{
			name:          "successful role update to customer",
			userID:        validUserID,
			role:          "customer",
			mockError:     nil,
			expectedError: false,
		},
		{
			name:          "invalid user ID",
			userID:        "invalid-uuid",
			role:          "admin",
			mockError:     nil,
			expectedError: true,
		},
		{
			name:          "invalid role",
			userID:        validUserID,
			role:          "invalid-role",
			mockError:     nil,
			expectedError: true,
		},
		{
			name:          "database error",
			userID:        validUserID,
			role:          "admin",
			mockError:     errors.New("database error"),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := new(MockUsersStore)

			if !tt.expectedError || tt.mockError != nil {
				mockStore.On("UpdateUserRole", mock.Anything, tt.userID, tt.role).Return(tt.mockError)
			}

			userService := service.NewUserService(mockStore)

			err := userService.UpdateUserRole(context.Background(), tt.userID, tt.role)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockStore.AssertExpectations(t)
		})
	}
}

func TestUserService_DeleteUserByID(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		mockError     error
		expectedError bool
	}{
		{
			name:          "successful user deletion",
			userID:        uuid.New().String(),
			mockError:     nil,
			expectedError: false,
		},
		{
			name:          "invalid user ID",
			userID:        "invalid-uuid",
			mockError:     nil,
			expectedError: true,
		},
		{
			name:          "database error",
			userID:        uuid.New().String(),
			mockError:     errors.New("database error"),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := new(MockUsersStore)

			if !tt.expectedError || tt.mockError != nil {
				mockStore.On("DeleteUser", mock.Anything, tt.userID).Return(tt.mockError)
			}

			userService := service.NewUserService(mockStore)

			err := userService.DeleteUserByID(context.Background(), tt.userID)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockStore.AssertExpectations(t)
		})
	}
}

func TestUserService_AddUser(t *testing.T) {
	tests := []struct {
		name          string
		user          models.User
		mockUser      models.User
		mockError     error
		expectedError bool
	}{
		{
			name: "successful user creation",
			user: models.User{
				GoogleID: "google-123",
				Name:     "John Doe",
				Email:    "john@example.com",
				Role:     "customer",
			},
			mockUser: models.User{
				ID:       uuid.New().String(),
				GoogleID: "google-123",
				Name:     "John Doe",
				Email:    "john@example.com",
				Role:     "customer",
			},
			mockError:     nil,
			expectedError: false,
		},
		{
			name: "missing name",
			user: models.User{
				GoogleID: "google-123",
				Name:     "",
				Email:    "john@example.com",
				Role:     "customer",
			},
			mockUser:      models.User{},
			mockError:     nil,
			expectedError: true,
		},
		{
			name: "missing Google ID",
			user: models.User{
				GoogleID: "",
				Name:     "John Doe",
				Email:    "john@example.com",
				Role:     "customer",
			},
			mockUser:      models.User{},
			mockError:     nil,
			expectedError: true,
		},
		{
			name: "missing email",
			user: models.User{
				GoogleID: "google-123",
				Name:     "John Doe",
				Email:    "",
				Role:     "customer",
			},
			mockUser:      models.User{},
			mockError:     nil,
			expectedError: true,
		},
		{
			name: "database error",
			user: models.User{
				GoogleID: "google-123",
				Name:     "John Doe",
				Email:    "john@example.com",
				Role:     "customer",
			},
			mockUser:      models.User{},
			mockError:     errors.New("database error"),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := new(MockUsersStore)

			if !tt.expectedError || tt.mockError != nil {
				mockStore.On("AddUser", mock.Anything, tt.user).Return(tt.mockUser, tt.mockError)
			}

			userService := service.NewUserService(mockStore)

			result, err := userService.AddUser(context.Background(), tt.user)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Equal(t, models.User{}, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.mockUser, result)
			}

			mockStore.AssertExpectations(t)
		})
	}
}

func TestUserService_ErrorTypes(t *testing.T) {
	mockStore := new(MockUsersStore)
	userService := service.NewUserService(mockStore)

	t.Run("UpdateUserByID with invalid UUID returns ErrUserNotFound", func(t *testing.T) {
		_, err := userService.UpdateUserByID(context.Background(), "invalid-uuid", models.User{Name: "Test"})
		assert.Error(t, err)
		assert.Equal(t, &customerrors.ErrUserNotFound, err)
	})

	t.Run("UpdateUserByID with missing name returns ErrMissingName", func(t *testing.T) {
		validID := uuid.New().String()
		_, err := userService.UpdateUserByID(context.Background(), validID, models.User{Name: ""})
		assert.Error(t, err)
		assert.Equal(t, &customerrors.ErrMissingName, err)
	})

	t.Run("UpdateUserRole with invalid UUID returns ErrUserNotFound", func(t *testing.T) {
		err := userService.UpdateUserRole(context.Background(), "invalid-uuid", "admin")
		assert.Error(t, err)
		assert.Equal(t, &customerrors.ErrUserNotFound, err)
	})

	t.Run("UpdateUserRole with invalid role returns ErrInvalidInput", func(t *testing.T) {
		validID := uuid.New().String()
		err := userService.UpdateUserRole(context.Background(), validID, "invalid-role")
		assert.Error(t, err)
		assert.Equal(t, &customerrors.ErrInvalidInput, err)
	})

	t.Run("DeleteUserByID with invalid UUID returns ErrUserNotFound", func(t *testing.T) {
		err := userService.DeleteUserByID(context.Background(), "invalid-uuid")
		assert.Error(t, err)
		assert.Equal(t, &customerrors.ErrUserNotFound, err)
	})

	t.Run("AddUser with missing name returns ErrMissingName", func(t *testing.T) {
		_, err := userService.AddUser(context.Background(), models.User{Name: ""})
		assert.Error(t, err)
		assert.Equal(t, &customerrors.ErrMissingName, err)
	})
}
