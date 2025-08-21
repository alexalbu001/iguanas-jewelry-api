package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/alexalbu001/iguanas-jewelry/internal/auth"
	"github.com/alexalbu001/iguanas-jewelry/internal/handlers"
	"github.com/alexalbu001/iguanas-jewelry/internal/middleware"
	"github.com/alexalbu001/iguanas-jewelry/internal/routes"
	"github.com/alexalbu001/iguanas-jewelry/internal/service"
	"github.com/alexalbu001/iguanas-jewelry/internal/store"
	"github.com/alexalbu001/iguanas-jewelry/internal/transaction"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/stripe/stripe-go/v82"
)

func init() {

}

func main() {
	r := gin.Default()

	logger := setupLogger()

	dbpool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		logger.Error("Unable to connect to db", "error", err)
		os.Exit(1)
	}
	// Verify the connection
	if err := dbpool.Ping(context.Background()); err != nil {
		logger.Error("Unable to ping database:", "error", err)
		os.Exit(1)
	}

	logger.Info("Connected to PostgreSQL database!")
	defer dbpool.Close()

	opt, err := redis.ParseURL(os.Getenv("REDIS_URL"))
	if err != nil {
		logger.Error("Unable to connect to redis", "error", err)
		os.Exit(1)
	}
	rdb := redis.NewClient(opt)

	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	logger.Info("Stripe SDK configured.")

	ctx := context.Background()
	sdkConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		logger.Error("Couldn't load AWS configuration:", "error", err)
		os.Exit(1)
	}

	// Create SQS client
	sqsClient := sqs.NewFromConfig(sdkConfig)

	//create repository layer
	productStore := store.NewProductStore(dbpool)

	userStore := store.NewUsersStore(dbpool)

	cartsStore := store.NewCartsStore(dbpool)

	sessionsStore := auth.NewSessionStore(rdb)

	ordersStore := store.NewOrdersStore(dbpool)

	paymentStore := store.NewPaymentStore(dbpool)

	//create service layer
	tx := transaction.NewTxManager(dbpool)

	productsService := service.NewProductsService(productStore)
	userService := service.NewUserService(userStore)
	cartsService := service.NewCartsService(cartsStore, productStore, userStore, tx)
	ordersService := service.NewOrderService(ordersStore, productStore, cartsStore, tx)
	paymentService := service.NewPaymentService(paymentStore, ordersStore)
	//create handlers with store

	productHandlers := handlers.NewProductHandlers(productsService)
	userHandlers := handlers.NewUserHandler(userService)
	authHandlers := auth.NewAuthHandlers(userStore, sessionsStore)
	cartHandlers := handlers.NewCartsHandler(cartsService, productsService)
	ordersHandlers := handlers.NewOrdersHandlers(ordersService, paymentService, sqsClient)
	paymentHandlers := handlers.NewPaymentHandler(paymentService, ordersService)

	authMiddleware := middleware.NewAuthMiddleware(sessionsStore)
	adminMiddleware := middleware.NewAdminMiddleware(sessionsStore, userStore)
	loggingMiddleware := middleware.NewLoggingMiddleware(logger)

	routes.SetupRoutes(r, productHandlers, userHandlers, cartHandlers, ordersHandlers, paymentHandlers, authHandlers, authMiddleware, adminMiddleware, loggingMiddleware)

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello World",
		})
	})

	r.Run(os.Getenv("PORT"))

}

func setupLogger() *slog.Logger {
	level := slog.LevelInfo
	switch os.Getenv("LOG_LEVEL") {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	case "info":
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler
	if os.Getenv("LOG_FORMAT") == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}
	logger := slog.New(handler).With(
		"service", "jewelry-api", // or from env var
		"env", os.Getenv("ENV"),
		"version", os.Getenv("VERSION"),
	)
	return logger
}
