// package store

// import (
// 	"context"
// )

// func BeginTransaction(order *OrdersStore, cart *CartsStore) error {
// 	ctx := context.Background()
// 	transaction, err := order.dbpool.Begin(ctx)
// 	if err != nil {
// 		return err
// 	}

// 	order.tx = transaction
// 	cart.tx = transaction
// 	return nil
// }

// func RollbackTransaction(order *OrdersStore, cart *CartsStore) error {
// 	ctx := context.Background()

// 	transaction := order.tx

// 	order.tx = nil
// 	cart.tx = nil

// 	return transaction.Rollback(ctx)
// }

// func CommitTransaction(order *OrdersStore, cart *CartsStore) error {
// 	ctx := context.Background()
// 	transaction := order.tx

// 	order.tx = nil
// 	cart.tx = nil

// 	return transaction.Commit(ctx)
// }
