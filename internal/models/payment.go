package models

import "time"

type Payment struct {
	ID              string    `json:"id"`
	UserID          string    `json:"user_id"`
	OrderID         string    `json:"order_id"`
	StripePaymentID *string   `json:"stripe_payment_id,omitempty"`
	Amount          float64   `json:"amount"`
	Status          string    `json:"status"`
	FailureReason   *string   `json:"failure_reason,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
