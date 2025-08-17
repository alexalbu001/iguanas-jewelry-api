package main

import (
	"context"
	"fmt"
	"log"
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

	dbpool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("Unable to connect to db", err)
	}
	// Verify the connection
	if err := dbpool.Ping(context.Background()); err != nil {
		log.Fatal("Unable to ping database:", err)
	}

	fmt.Println("Connected to PostgreSQL database!")

	defer dbpool.Close()

	opt, err := redis.ParseURL(os.Getenv("REDIS_URL"))
	if err != nil {
		log.Fatal("Unable to connect to redis", err)
	}
	rdb := redis.NewClient(opt)

	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	log.Println("Stripe SDK configured.")

	ctx := context.Background()
	sdkConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatal("Couldn't load AWS configuration:", err)
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

	routes.SetupRoutes(r, productHandlers, userHandlers, cartHandlers, ordersHandlers, paymentHandlers, authHandlers, authMiddleware, adminMiddleware)

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
