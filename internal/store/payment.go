package store

import (
	"context"
	"fmt"
	"time"

	customerrors "github.com/alexalbu001/iguanas-jewelry/internal/customErrors"
	"github.com/alexalbu001/iguanas-jewelry/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PaymentStore struct {
	dbpool *pgxpool.Pool
}

func NewPaymentStore(dbpool *pgxpool.Pool) *PaymentStore {
	return &PaymentStore{
		dbpool: dbpool,
	}
}

func (p *PaymentStore) CreatePayment(ctx context.Context, payment models.Payment) (models.Payment, error) {
	sql := `
	INSERT INTO payment (id, user_id, order_id, stripe_payment_id, amount, status, failure_reason, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	paymentID := uuid.NewString()
	now := time.Now()
	_, err := p.dbpool.Exec(ctx, sql, paymentID, payment.UserID, payment.OrderID, payment.StripePaymentID, payment.Amount, payment.Status, payment.FailureReason, now, now)
	if err != nil {
		return models.Payment{}, fmt.Errorf("Failed to create payment: %w", err)
	}

	createdPayment := models.Payment{
		ID:              paymentID,
		UserID:          payment.UserID,
		OrderID:         payment.OrderID,
		StripePaymentID: payment.StripePaymentID,
		Amount:          payment.Amount,
		Status:          payment.Status,
		FailureReason:   payment.FailureReason,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	return createdPayment, nil
}

func (p *PaymentStore) GetPaymentByID(ctx context.Context, paymentID string) (models.Payment, error) {
	sql := `
	SELECT id, user_id, order_id, stripe_payment_id, amount, status, failure_reason, created_at, updated_at
	FROM payment
	WHERE id=$1
	`

	row := p.dbpool.QueryRow(ctx, sql, paymentID)
	var returnedPayment models.Payment
	err := row.Scan(
		&returnedPayment.ID,
		&returnedPayment.UserID,
		&returnedPayment.OrderID,
		&returnedPayment.StripePaymentID,
		&returnedPayment.Amount,
		&returnedPayment.Status,
		&returnedPayment.FailureReason,
		&returnedPayment.CreatedAt,
		&returnedPayment.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return models.Payment{}, &customerrors.ErrPaymentNotFound
		}
		return models.Payment{}, fmt.Errorf("Failed to scan payment: %w", err)
	}

	return returnedPayment, nil
}

func (p *PaymentStore) GetPaymentsByUserID(ctx context.Context, paymentUserID string) ([]models.Payment, error) {
	sql := `
	SELECT id, user_id, order_id, stripe_payment_id, amount, status, failure_reason, created_at, updated_at
	FROM payment
	WHERE user_id=$1
	ORDER by user_id, created_at
	`

	rows, err := p.dbpool.Query(ctx, sql, paymentUserID)
	if err != nil {
		return nil, fmt.Errorf("Error retrieving orders: %w", err)
	}
	defer rows.Close()
	var returnedPayments []models.Payment
	for rows.Next() {
		var returnedPayment models.Payment
		err := rows.Scan(
			&returnedPayment.ID,
			&returnedPayment.UserID,
			&returnedPayment.OrderID,
			&returnedPayment.StripePaymentID,
			&returnedPayment.Amount,
			&returnedPayment.Status,
			&returnedPayment.FailureReason,
			&returnedPayment.CreatedAt,
			&returnedPayment.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("Error scanning payment row: %w", err)
		}
		returnedPayments = append(returnedPayments, returnedPayment)
	}

	return returnedPayments, nil
}

func (p *PaymentStore) GetPaymentsByOrderID(ctx context.Context, paymentOrderID string) ([]models.Payment, error) {
	sql := `
	SELECT id, user_id, order_id, stripe_payment_id, amount, status, failure_reason, created_at, updated_at
	FROM payment
	WHERE order_id=$1
	ORDER by user_id, created_at
	`

	rows, err := p.dbpool.Query(ctx, sql, paymentOrderID)
	if err != nil {
		return nil, fmt.Errorf("Error retrieving orders: %w", err)
	}
	defer rows.Close()
	var returnedPayments []models.Payment
	for rows.Next() {
		var returnedPayment models.Payment
		err := rows.Scan(
			&returnedPayment.ID,
			&returnedPayment.UserID,
			&returnedPayment.OrderID,
			&returnedPayment.StripePaymentID,
			&returnedPayment.Amount,
			&returnedPayment.Status,
			&returnedPayment.FailureReason,
			&returnedPayment.CreatedAt,
			&returnedPayment.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("Error scanning payment row: %w", err)
		}
		returnedPayments = append(returnedPayments, returnedPayment)
	}

	return returnedPayments, nil
}

func (p *PaymentStore) GetPaymentByStripeID(ctx context.Context, paymentStripeID string) (models.Payment, error) {
	sql := `
	SELECT id, user_id, order_id, stripe_payment_id, amount, status, failure_reason, created_at, updated_at
	FROM payment
	WHERE stripe_payment_id=$1
	`

	row := p.dbpool.QueryRow(ctx, sql, paymentStripeID)
	var returnedPayment models.Payment
	err := row.Scan(
		&returnedPayment.ID,
		&returnedPayment.UserID,
		&returnedPayment.OrderID,
		&returnedPayment.StripePaymentID,
		&returnedPayment.Amount,
		&returnedPayment.Status,
		&returnedPayment.FailureReason,
		&returnedPayment.CreatedAt,
		&returnedPayment.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return models.Payment{}, &customerrors.ErrPaymentNotFound
		}
		return models.Payment{}, fmt.Errorf("Failed to scan payment: %w", err)
	}

	return returnedPayment, nil
}

func (p *PaymentStore) UpdatePaymentStatus(ctx context.Context, id, paymentStatus string, stripePaymentID, failureReason *string) error {
	sql := `
	UPDATE payment 
	SET 
		status = $2,
		stripe_payment_id = CASE WHEN $3 IS NOT NULL THEN $3 ELSE stripe_payment_id END,
		failure_reason = CASE WHEN $4 IS NOT NULL THEN $4 ELSE failure_reason END,
		updated_at = $5
	WHERE id = $1
	`

	commandTag, err := p.dbpool.Exec(ctx, sql, id, paymentStatus, stripePaymentID, failureReason, time.Now())
	if err != nil {
		return fmt.Errorf("Failed to update payment status : %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return &customerrors.ErrPaymentNotFound
	}

	return nil
}

func (p *PaymentStore) GetPaymentsByStatus(ctx context.Context, status string) ([]models.Payment, error) {
	sql := `
	SELECT id, user_id, order_id, stripe_payment_id, amount, status, failure_reason, created_at, updated_at
	FROM payment
	WHERE status=$1
	ORDER by user_id, created_at
	`

	rows, err := p.dbpool.Query(ctx, sql, status)
	if err != nil {
		return nil, fmt.Errorf("Error retrieving orders: %w", err)
	}
	defer rows.Close()
	var returnedPayments []models.Payment
	for rows.Next() {
		var returnedPayment models.Payment
		err := rows.Scan(
			&returnedPayment.ID,
			&returnedPayment.UserID,
			&returnedPayment.OrderID,
			&returnedPayment.StripePaymentID,
			&returnedPayment.Amount,
			&returnedPayment.Status,
			&returnedPayment.FailureReason,
			&returnedPayment.CreatedAt,
			&returnedPayment.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("Error scanning payment row: %w", err)
		}
		returnedPayments = append(returnedPayments, returnedPayment)
	}

	return returnedPayments, nil
}
