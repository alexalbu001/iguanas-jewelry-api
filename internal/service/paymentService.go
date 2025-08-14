package service

import (
	"context"
	"fmt"
	"time"

	"math/rand/v2"

	customerrors "github.com/alexalbu001/iguanas-jewelry/internal/customErrors"
	"github.com/alexalbu001/iguanas-jewelry/internal/models"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/paymentintent"
)

type PaymentStore interface {
	CreatePayment(ctx context.Context, payment models.Payment) (models.Payment, error)
	GetPaymentByID(ctx context.Context, paymentID string) (models.Payment, error)
	GetPaymentsByUserID(ctx context.Context, paymentUserID string) ([]models.Payment, error)
	GetPaymentsByOrderID(ctx context.Context, paymentOrderID string) ([]models.Payment, error)
	GetPaymentByStripeID(ctx context.Context, paymentStripeID string) (models.Payment, error)
	UpdatePaymentStatus(ctx context.Context, id, paymentStatus string, stripePaymentID, failureReason *string) error
	GetPaymentsByStatus(ctx context.Context, status string) ([]models.Payment, error)
}

type PaymentService struct {
	PaymentStore PaymentStore
	OrdersStore  OrdersStore
}

func NewPaymentService(paymentStore PaymentStore, ordersStore OrdersStore) *PaymentService {
	return &PaymentService{
		PaymentStore: paymentStore,
		OrdersStore:  ordersStore,
	}
}

func (p *PaymentService) CreatePaymentIntent(ctx context.Context, orderID, idempotencyKey string) (string, error) {
	// --- 1. SETUP ---
	// We declare the variables we'll need here.
	var pi *stripe.PaymentIntent
	var err error
	maxRetries := 3

	// Get order details once, before the loop starts.
	order, err := p.OrdersStore.GetOrderByID(ctx, orderID)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve order: %w", err)
	}

	params := &stripe.PaymentIntentParams{
		Amount:       stripe.Int64(int64(order.TotalAmount)),
		Currency:     stripe.String(string(stripe.CurrencyGBP)),
		ReceiptEmail: stripe.String(string(order.ShippingEmail)),
		Customer:     stripe.String(string(order.UserID)),
	}
	params.SetIdempotencyKey(idempotencyKey)

	// --- 2. THE RETRY LOOP ---
	for attempt := 1; attempt <= maxRetries; attempt++ {
		pi, err = paymentintent.New(params)

		// If there's no error, we succeeded! Exit the loop.
		if err == nil {
			break
		}

		// Check if the error is a permanent one from Stripe.
		if stripeErr, ok := err.(*stripe.Error); ok {
			// If it's NOT a temporary connection error, it's permanent.
			// We should give up immediately.
			if stripeErr.Type != stripe.ErrorTypeAPI {
				// Translate this permanent error and exit the entire function now.
				return "", p.translateStripeError(err)
			}
		}

		// If we're here, the error was temporary (connection error or generic).
		// Wait before the next try, unless it's the very last attempt.
		if attempt < maxRetries {
			jitter := time.Duration(rand.N(100)) * time.Millisecond
			time.Sleep(time.Second*time.Duration(attempt) + jitter)
		}
	}

	// --- 3. FINAL CHECK ---
	// After the loop, if we still have an error, it means all our retries failed.
	// We translate this final error and return it.
	if err != nil {
		return "", p.translateStripeError(err)
	}

	// If we're here, `pi` must be valid and `err` must be nil. Success!
	return pi.ClientSecret, nil
}

func (p *PaymentService) translateStripeError(err error) error {
	if stripeErr, ok := err.(*stripe.Error); ok {
		switch stripeErr.Code {
		case stripe.ErrorCodeCardDeclined:
			return &customerrors.ErrPaymentCardDeclined
		case stripe.ErrorCodeExpiredCard:
			return &customerrors.ErrPaymentCardExpired
		case stripe.ErrorCodeIncorrectCVC:
			return &customerrors.ErrPaymentIncorrectCVC
		default:
			// For any other specific Stripe code, or a connection error after all retries.
			return &customerrors.ErrPaymentProcessingFailed
		}
	}
	// For a generic non-Stripe error (e.g. net/http) after all retries.
	return &customerrors.ErrPaymentProcessingFailed
}
