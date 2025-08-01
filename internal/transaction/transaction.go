package transaction

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Transaction struct {
	dbpool *pgxpool.Pool
}

type TxManager interface {
	WithTransaction(ctx context.Context, fn func(tx pgx.Tx) error) error
}

func NewTxManager(dbpool *pgxpool.Pool) *Transaction {
	return &Transaction{
		dbpool: dbpool,
	}
}

func (t *Transaction) WithTransaction(ctx context.Context, fn func(tx pgx.Tx) error) error {
	tx, err := t.dbpool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("Failed to begin transaction: %w", err)
	}

	defer tx.Rollback(ctx)

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
