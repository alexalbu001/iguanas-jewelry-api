package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/alexalbu001/iguanas-jewelry-api/docs"
	"github.com/alexalbu001/iguanas-jewelry-api/internal/auth"
	"github.com/alexalbu001/iguanas-jewelry-api/internal/config"
	"github.com/alexalbu001/iguanas-jewelry-api/internal/handlers"
	"github.com/alexalbu001/iguanas-jewelry-api/internal/middleware"
	"github.com/alexalbu001/iguanas-jewelry-api/internal/routes"
	"github.com/alexalbu001/iguanas-jewelry-api/internal/service"
	"github.com/alexalbu001/iguanas-jewelry-api/internal/store"
	"github.com/alexalbu001/iguanas-jewelry-api/internal/telemetry"
	"github.com/alexalbu001/iguanas-jewelry-api/internal/transaction"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/exaring/otelpgx"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/stripe/stripe-go/v82"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func init() {

}

// @title           Iguanas Jewelry API
// @version         1.0
// @description     A jewelry e-commerce API
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	ctx := context.Background()
	r := gin.Default()
	cfg, err := config.Load()
	if err != nil {
		fmt.Errorf("Failed to load ENV vars")
		os.Exit(1)
	}

	logger := setupLogger(cfg)

	if err := cfg.Validate(); err != nil {
		logger.Error("Configuration validation failed", "error", err)
		fmt.Errorf("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	telemtry, err := telemetry.InitTelemetry(ctx, "iguanas-jewelry", cfg.Version, cfg.Env)
	if err != nil {
		logger.Error("Failed to init telemetry:", "error", err)
		fmt.Errorf("Server forced to shutdown", "error", err)
		os.Exit(1)
	}
	defer telemtry.Shutdown(ctx)

	pgxConfig, err := pgxpool.ParseConfig(cfg.Database.DatabaseURL)
	if err != nil {
		fmt.Errorf("Server forced to shutdown", "error", err)
		logger.Error("Unable to parse database URL", "error", err)
		os.Exit(1)
	}

	pgxConfig.ConnConfig.Tracer = otelpgx.NewTracer()

	dbpool, err := pgxpool.NewWithConfig(context.Background(), pgxConfig)
	if err != nil {
		fmt.Errorf("Server forced to shutdown", "error", err)
		logger.Error("Unable to connect to db", "error", err)
		os.Exit(1)
	}

	if err := dbpool.Ping(ctx); err != nil {
		fmt.Errorf("Server forced to shutdown", "error", err)
		logger.Error("Unable to ping database:", "error", err)
		os.Exit(1)
	}

	logger.Info("Connected to PostgreSQL database!")
	defer dbpool.Close()

	opt, err := redis.ParseURL(cfg.Redis.RedisURL)
	if err != nil {
		fmt.Errorf("Server forced to shutdown", "error", err)
		logger.Error("Unable to connect to redis", "error", err)
		os.Exit(1)
	}
	rdb := redis.NewClient(opt)

	stripe.Key = cfg.Stripe.StripeSK
	logger.Info("Stripe SDK configured.")

	sdkConfig, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		fmt.Errorf("Server forced to shutdown", "error", err)
		logger.Error("Couldn't load AWS configuration:", "error", err)
		os.Exit(1)
	}

	var conf = &oauth2.Config{
		ClientID:     cfg.Google.ClientID,
		ClientSecret: cfg.Google.ClientSecret,
		RedirectURL:  cfg.Google.RedirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
	adminEmail := cfg.AdminEmail
	queueURL := cfg.SQS.QueueURL
	stripeWebhookSecret := cfg.Stripe.StripeWebhookSecret

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

	// Cached
	cachedProductsStore := store.NewCachedProductsStore(productStore, rdb)

	productsService := service.NewProductsService(cachedProductsStore)
	userService := service.NewUserService(userStore)
	cartsService := service.NewCartsService(cartsStore, productStore, userStore, tx)
	ordersService := service.NewOrderService(ordersStore, productStore, cartsStore, tx)
	paymentService := service.NewPaymentService(paymentStore, ordersStore)
	jwtService := auth.NewJWTService(cfg.JWTSecret)
	//create handlers with store

	productHandlers := handlers.NewProductHandlers(productsService)
	userHandlers := handlers.NewUserHandler(userService)
	authHandlers := auth.NewAuthHandlers(userStore, sessionsStore, conf, adminEmail, jwtService)
	cartHandlers := handlers.NewCartsHandler(cartsService, productsService)
	ordersHandlers := handlers.NewOrdersHandlers(ordersService, paymentService, sqsClient, queueURL)
	paymentHandlers := handlers.NewPaymentHandler(paymentService, ordersService, stripeWebhookSecret)

	authMiddleware := middleware.NewAuthMiddleware(sessionsStore)
	adminMiddleware := middleware.NewAdminMiddleware(sessionsStore, userStore)
	loggingMiddleware := middleware.NewLoggingMiddleware(logger)
	rateLimitMiddleware := middleware.NewRateLimiter(rdb, "")

	routes.SetupRoutes(r, cfg, productHandlers, userHandlers, cartHandlers, ordersHandlers, paymentHandlers, authHandlers, authMiddleware, adminMiddleware, loggingMiddleware, rateLimitMiddleware)

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.GET("/health", healthCheck(dbpool, rdb))

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello World",
		})
	})

	// Add this after your routes setup:
	if cfg.Env != "production" {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.AppPort),
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Channel to listen for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if cfg.Env == "production" {
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Error("Server failed to start", "error", err)
				os.Exit(1)
			}
		} else {
			if err := srv.ListenAndServeTLS("localhost+2.pem", "localhost+2-key.pem"); err != nil && err != http.ErrServerClosed {
				fmt.Errorf("Server forced to shutdown", "error", err)
				logger.Error("Server failed to start", "error", err)
				os.Exit(1)
			}
		}
	}()

	logger.Info("Server started", "port", cfg.AppPort, "env", cfg.Env)

	<-quit
	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {

		logger.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	logger.Info("Server exited")

}

func setupLogger(cfg *config.Config) *slog.Logger {
	level := slog.LevelInfo
	switch cfg.Logging.LogLevel {
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
	if cfg.Logging.LogFormat == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}
	logger := slog.New(handler).With(
		"service", "jewelry-api", // or from env var
		"env", cfg.Env,
		"version", cfg.Version,
	)
	return logger
}

func healthCheck(dbpool *pgxpool.Pool, redis *redis.Client) gin.HandlerFunc { // returns gin.HandlerFun because of factory pattern has access to db dependencies
	return func(c *gin.Context) {
		if err := dbpool.Ping(context.Background()); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "unhealthy",
				"error":  "database"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
		})
	}
}
