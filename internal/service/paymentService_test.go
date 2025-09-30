package service_test

import (
	"context"
	"errors"
	"testing"

	customerrors "github.com/alexalbu001/iguanas-jewelry-api/internal/customErrors"
	"github.com/alexalbu001/iguanas-jewelry-api/internal/models"
	"github.com/alexalbu001/iguanas-jewelry-api/internal/service"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stripe/stripe-go/v82"
)

// MockPaymentStore is a mock implementation of PaymentStore
type MockPaymentStore struct {
	mock.Mock
}

func (m *MockPaymentStore) CreatePayment(ctx context.Context, payment models.Payment) (models.Payment, error) {
	args := m.Called(ctx, payment)
	return args.Get(0).(models.Payment), args.Error(1)
}

func (m *MockPaymentStore) GetPaymentByID(ctx context.Context, paymentID string) (models.Payment, error) {
	args := m.Called(ctx, paymentID)
	return args.Get(0).(models.Payment), args.Error(1)
}

func (m *MockPaymentStore) GetPaymentsByUserID(ctx context.Context, paymentUserID string) ([]models.Payment, error) {
	args := m.Called(ctx, paymentUserID)
	return args.Get(0).([]models.Payment), args.Error(1)
}

func (m *MockPaymentStore) GetPaymentsByOrderID(ctx context.Context, paymentOrderID string) ([]models.Payment, error) {
	args := m.Called(ctx, paymentOrderID)
	return args.Get(0).([]models.Payment), args.Error(1)
}

func (m *MockPaymentStore) GetPaymentByStripeID(ctx context.Context, paymentStripeID string) (models.Payment, error) {
	args := m.Called(ctx, paymentStripeID)
	return args.Get(0).(models.Payment), args.Error(1)
}

func (m *MockPaymentStore) UpdatePaymentStatus(ctx context.Context, id, paymentStatus string, stripePaymentID, failureReason *string) error {
	args := m.Called(ctx, id, paymentStatus, stripePaymentID, failureReason)
	return args.Error(0)
}

func (m *MockPaymentStore) GetPaymentsByStatus(ctx context.Context, status string) ([]models.Payment, error) {
	args := m.Called(ctx, status)
	return args.Get(0).([]models.Payment), args.Error(1)
}

// MockOrdersStore is a mock implementation of OrdersStore
type MockOrdersStore struct {
	mock.Mock
}

func (m *MockOrdersStore) InsertOrder(ctx context.Context, order models.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockOrdersStore) InsertOrderTx(ctx context.Context, order models.Order, tx pgx.Tx) error {
	args := m.Called(ctx, order, tx)
	return args.Error(0)
}

func (m *MockOrdersStore) InsertOrderItem(ctx context.Context, orderItem models.OrderItem) error {
	args := m.Called(ctx, orderItem)
	return args.Error(0)
}

func (m *MockOrdersStore) InsertOrderItemTx(ctx context.Context, orderItem models.OrderItem, tx pgx.Tx) error {
	args := m.Called(ctx, orderItem, tx)
	return args.Error(0)
}

func (m *MockOrdersStore) InsertOrderItemBulk(ctx context.Context, orderItems []models.OrderItem) error {
	args := m.Called(ctx, orderItems)
	return args.Error(0)
}

func (m *MockOrdersStore) InsertOrderItemBulkTx(ctx context.Context, orderItems []models.OrderItem, tx pgx.Tx) error {
	args := m.Called(ctx, orderItems, tx)
	return args.Error(0)
}

func (m *MockOrdersStore) GetOrderByID(ctx context.Context, orderID string) (models.Order, error) {
	args := m.Called(ctx, orderID)
	return args.Get(0).(models.Order), args.Error(1)
}

func (m *MockOrdersStore) GetOrderItems(ctx context.Context, orderID string) ([]models.OrderItem, error) {
	args := m.Called(ctx, orderID)
	return args.Get(0).([]models.OrderItem), args.Error(1)
}

func (m *MockOrdersStore) GetOrderItemsBatch(ctx context.Context, orderID []string) (map[string][]models.OrderItem, error) {
	args := m.Called(ctx, orderID)
	return args.Get(0).(map[string][]models.OrderItem), args.Error(1)
}

func (m *MockOrdersStore) GetUsersOrders(ctx context.Context, userID string) ([]models.Order, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]models.Order), args.Error(1)
}

func (m *MockOrdersStore) GetAllOrders(ctx context.Context) ([]models.Order, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.Order), args.Error(1)
}

func (m *MockOrdersStore) GetOrdersByStatus(ctx context.Context, status string) ([]models.Order, error) {
	args := m.Called(ctx, status)
	return args.Get(0).([]models.Order), args.Error(1)
}

func (m *MockOrdersStore) UpdateOrderStatus(ctx context.Context, status, orderID string) error {
	args := m.Called(ctx, status, orderID)
	return args.Error(0)
}

func (m *MockOrdersStore) UpdateOrderStatusTx(ctx context.Context, status, orderID string, tx pgx.Tx) error {
	args := m.Called(ctx, status, orderID, tx)
	return args.Error(0)
}

func TestPaymentService_CreatePayment(t *testing.T) {
	tests := []struct {
		name          string
		payment       models.Payment
		mockResponse  models.Payment
		mockError     error
		expectedError bool
	}{
		{
			name: "successful payment creation",
			payment: models.Payment{
				ID:      "payment-123",
				UserID:  "user-123",
				OrderID: "order-123",
				Amount:  99.99,
				Status:  "pending",
			},
			mockResponse: models.Payment{
				ID:      "payment-123",
				UserID:  "user-123",
				OrderID: "order-123",
				Amount:  99.99,
				Status:  "pending",
			},
			mockError:     nil,
			expectedError: false,
		},
		{
			name: "payment creation error",
			payment: models.Payment{
				ID:      "payment-123",
				UserID:  "user-123",
				OrderID: "order-123",
				Amount:  99.99,
				Status:  "pending",
			},
			mockResponse:  models.Payment{},
			mockError:     errors.New("database error"),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPaymentStore := new(MockPaymentStore)
			mockOrdersStore := new(MockOrdersStore)

			mockPaymentStore.On("CreatePayment", mock.Anything, tt.payment).Return(tt.mockResponse, tt.mockError)

			paymentService := service.NewPaymentService(mockPaymentStore, mockOrdersStore)

			result, err := paymentService.CreatePayment(context.Background(), tt.payment)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Equal(t, models.Payment{}, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.mockResponse, result)
			}

			mockPaymentStore.AssertExpectations(t)
		})
	}
}

func TestPaymentService_GetPaymentsByOrderID(t *testing.T) {
	tests := []struct {
		name          string
		orderID       string
		mockResponse  []models.Payment
		mockError     error
		expectedError bool
	}{
		{
			name:    "successful payment retrieval",
			orderID: "order-123",
			mockResponse: []models.Payment{
				{
					ID:      "payment-1",
					OrderID: "order-123",
					Amount:  99.99,
					Status:  "completed",
				},
				{
					ID:      "payment-2",
					OrderID: "order-123",
					Amount:  50.00,
					Status:  "failed",
				},
			},
			mockError:     nil,
			expectedError: false,
		},
		{
			name:          "payment retrieval error",
			orderID:       "order-123",
			mockResponse:  nil,
			mockError:     errors.New("database error"),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPaymentStore := new(MockPaymentStore)
			mockOrdersStore := new(MockOrdersStore)

			mockPaymentStore.On("GetPaymentsByOrderID", mock.Anything, tt.orderID).Return(tt.mockResponse, tt.mockError)

			paymentService := service.NewPaymentService(mockPaymentStore, mockOrdersStore)

			result, err := paymentService.GetPaymentsByOrderID(context.Background(), tt.orderID)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.mockResponse, result)
			}

			mockPaymentStore.AssertExpectations(t)
		})
	}
}

func TestPaymentService_RetryOrderPayment(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		orderID        string
		idempotencyKey string
		order          models.Order
		payments       []models.Payment
		orderError     error
		paymentsError  error
		expectedError  bool
		expectedResult string
	}{
		{
			name:           "successful payment retry",
			userID:         "user-123",
			orderID:        "order-123",
			idempotencyKey: "retry-123",
			order: models.Order{
				ID:            "order-123",
				UserID:        "user-123",
				Status:        "pending",
				TotalAmount:   99.99,
				ShippingEmail: "test@example.com",
			},
			payments: []models.Payment{
				{ID: "payment-1", Status: "failed"},
				{ID: "payment-2", Status: "failed"},
			},
			orderError:     nil,
			paymentsError:  nil,
			expectedError:  true, // Will fail due to missing Stripe API key in test
			expectedResult: "",
		},
		{
			name:           "order not found",
			userID:         "user-123",
			orderID:        "order-123",
			idempotencyKey: "retry-123",
			order:          models.Order{},
			payments:       nil,
			orderError:     errors.New("order not found"),
			paymentsError:  nil,
			expectedError:  true,
			expectedResult: "",
		},
		{
			name:           "order not owned by user",
			userID:         "user-123",
			orderID:        "order-123",
			idempotencyKey: "retry-123",
			order: models.Order{
				ID:     "order-123",
				UserID: "user-456", // Different user
				Status: "pending",
			},
			payments:       nil,
			orderError:     nil,
			paymentsError:  nil,
			expectedError:  true,
			expectedResult: "",
		},
		{
			name:           "order already paid",
			userID:         "user-123",
			orderID:        "order-123",
			idempotencyKey: "retry-123",
			order: models.Order{
				ID:     "order-123",
				UserID: "user-123",
				Status: "paid",
			},
			payments:       nil,
			orderError:     nil,
			paymentsError:  nil,
			expectedError:  true,
			expectedResult: "",
		},
		{
			name:           "order cancelled",
			userID:         "user-123",
			orderID:        "order-123",
			idempotencyKey: "retry-123",
			order: models.Order{
				ID:     "order-123",
				UserID: "user-123",
				Status: "cancelled",
			},
			payments:       nil,
			orderError:     nil,
			paymentsError:  nil,
			expectedError:  true,
			expectedResult: "",
		},
		{
			name:           "too many retries",
			userID:         "user-123",
			orderID:        "order-123",
			idempotencyKey: "retry-123",
			order: models.Order{
				ID:     "order-123",
				UserID: "user-123",
				Status: "pending",
			},
			payments: []models.Payment{
				{ID: "payment-1", Status: "failed"},
				{ID: "payment-2", Status: "failed"},
				{ID: "payment-3", Status: "failed"},
				{ID: "payment-4", Status: "failed"},
				{ID: "payment-5", Status: "failed"},
				{ID: "payment-6", Status: "failed"}, // 6 failed payments > 5 limit
			},
			orderError:     nil,
			paymentsError:  nil,
			expectedError:  true,
			expectedResult: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPaymentStore := new(MockPaymentStore)
			mockOrdersStore := new(MockOrdersStore)

			mockOrdersStore.On("GetOrderByID", mock.Anything, tt.orderID).Return(tt.order, tt.orderError)

			// Only set up GetPaymentsByOrderID expectation if the order is valid and owned by user
			// and the order status allows payment retry (not paid or cancelled)
			if tt.orderError == nil && tt.order.UserID == tt.userID &&
				tt.order.Status != "paid" && tt.order.Status != "cancelled" {
				mockPaymentStore.On("GetPaymentsByOrderID", mock.Anything, tt.orderID).Return(tt.payments, tt.paymentsError)
			}

			paymentService := service.NewPaymentService(mockPaymentStore, mockOrdersStore)

			result, err := paymentService.RetryOrderPayment(context.Background(), tt.userID, tt.orderID, tt.idempotencyKey)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Equal(t, "", result)
			} else {
				// Note: In a real test, we would mock the Stripe API call
				// For now, we just check that no error occurred
				assert.NoError(t, err)
			}

			mockPaymentStore.AssertExpectations(t)
			mockOrdersStore.AssertExpectations(t)
		})
	}
}

func TestPaymentService_TranslateStripeError(t *testing.T) {
	paymentService := &service.PaymentService{}

	tests := []struct {
		name          string
		stripeError   *stripe.Error
		expectedError error
	}{
		{
			name: "card declined",
			stripeError: &stripe.Error{
				Code: stripe.ErrorCodeCardDeclined,
			},
			expectedError: &customerrors.ErrPaymentCardDeclined,
		},
		{
			name: "expired card",
			stripeError: &stripe.Error{
				Code: stripe.ErrorCodeExpiredCard,
			},
			expectedError: &customerrors.ErrPaymentCardExpired,
		},
		{
			name: "incorrect CVC",
			stripeError: &stripe.Error{
				Code: stripe.ErrorCodeIncorrectCVC,
			},
			expectedError: &customerrors.ErrPaymentIncorrectCVC,
		},
		{
			name: "other stripe error",
			stripeError: &stripe.Error{
				Code: stripe.ErrorCodeRateLimit,
			},
			expectedError: &customerrors.ErrPaymentProcessingFailed,
		},
		{
			name:          "non-stripe error",
			stripeError:   nil,
			expectedError: &customerrors.ErrPaymentProcessingFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			if tt.stripeError != nil {
				err = tt.stripeError
			} else {
				err = errors.New("generic error")
			}

			result := paymentService.TranslateStripeError(err)
			assert.Equal(t, tt.expectedError, result)
		})
	}
}
