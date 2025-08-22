package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/alexalbu001/iguanas-jewelry/internal/config"
	"github.com/alexalbu001/iguanas-jewelry/internal/service"
	"github.com/alexalbu001/iguanas-jewelry/internal/store"
	"github.com/alexalbu001/iguanas-jewelry/internal/transaction"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ExpirationMessage struct {
	OrderID   string    `json:"order_id"`
	CreatedAt time.Time `json:"created_at`
}

var ordersService *service.OrdersService

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Errorf("Failed to load ENV vars")
		os.Exit(1)
	}
	dbpool, err := pgxpool.New(context.Background(), cfg.Database.DatabaseURL)
	if err != nil {
		log.Fatal("Unable to connect to db", err)
	}
	// Verify the connection
	if err := dbpool.Ping(context.Background()); err != nil {
		log.Fatal("Unable to ping database:", err)
	}

	defer dbpool.Close()

	productStore := store.NewProductStore(dbpool)

	cartsStore := store.NewCartsStore(dbpool)

	ordersStore := store.NewOrdersStore(dbpool)

	tx := transaction.NewTxManager(dbpool)

	ordersService = service.NewOrderService(ordersStore, productStore, cartsStore, tx)

	lambda.Start(handleExpiration)

}

func handleExpiration(ctx context.Context, sqsEvent events.SQSEvent) ([]events.SQSBatchItemFailure, error) {
	log.Printf("🚀 Lambda started! Processing %d messages", len(sqsEvent.Records))

	var batchFailures []events.SQSBatchItemFailure
	for _, record := range sqsEvent.Records {
		var msg ExpirationMessage
		err := json.Unmarshal([]byte(record.Body), &msg)
		if err != nil {
			batchFailures = append(batchFailures, events.SQSBatchItemFailure{
				ItemIdentifier: record.MessageId,
			})
			continue
		}

		_, err = ordersService.GetOrderByIDAdmin(ctx, msg.OrderID)
		if err != nil {
			batchFailures = append(batchFailures, events.SQSBatchItemFailure{
				ItemIdentifier: record.MessageId,
			})
			continue
		}
		err = ordersService.UpdateOrderStatus(ctx, "cancelled", msg.OrderID)
		if err != nil {
			batchFailures = append(batchFailures, events.SQSBatchItemFailure{
				ItemIdentifier: record.MessageId,
			})
			continue
		}
		log.Printf("✅ Successfully processed order: %s", msg.OrderID)
	}

	return batchFailures, nil
}
