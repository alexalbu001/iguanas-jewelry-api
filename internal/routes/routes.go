package routes

import (
	"github.com/alexalbu001/iguanas-jewelry-api/internal/auth"
	"github.com/alexalbu001/iguanas-jewelry-api/internal/config"
	"github.com/alexalbu001/iguanas-jewelry-api/internal/handlers"
	"github.com/alexalbu001/iguanas-jewelry-api/internal/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/secure"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func SetupRoutes(r *gin.Engine, cfg *config.Config, productHandlers *handlers.ProductHandlers, userHandlers *handlers.UserHandlers, cartsHandlers *handlers.CartsHandlers, ordersHandlers *handlers.OrdersHandlers, paymentHandlers *handlers.PaymentHandler, productImagesHandlers *handlers.ProductImagesHandlers, userFavoritesHandlers *handlers.UserFavoritesHandlers, imageStorageHandler *handlers.ImageStorageHandlers, authHandlers *auth.AuthHandlers, authMiddleware *middleware.AuthMiddleware, adminMiddleware *middleware.AdminMiddleware, loggingMiddleware *middleware.LoggingMiddleware, rateLimitMiddleware gin.HandlerFunc) {
	r.Use(cors.New(cors.Config{ // CORS
		AllowOrigins:     cfg.CORS.AllowOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Content-Type", "Authorization", "X-CSRF-Token"},
		AllowCredentials: true, //Cookies
	}))
	r.Use(middleware.LimitRequestSize(1 << 20)) // 1MB requests size
	if cfg.Env == "production" {                // SECURITY HEADERS
		r.Use(secure.New(secure.Config{
			STSSeconds:            31536000,
			STSIncludeSubdomains:  true,
			FrameDeny:             true,
			ContentTypeNosniff:    true,
			ReferrerPolicy:        "strict-origin-when-cross-origin",
			ContentSecurityPolicy: "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self' data: https:; connect-src 'self';",
		}))
	} else {
		r.Use(secure.New(secure.Config{
			FrameDeny:             true,
			ContentTypeNosniff:    true,
			ReferrerPolicy:        "strict-origin-when-cross-origin",
			ContentSecurityPolicy: "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self' data: https:; connect-src 'self';",
		}))
	}
	r.Use(middleware.CorrelationID())
	r.Use(loggingMiddleware.RequestLogging()) //Use middleware abroad whole gin engine
	r.Use(otelgin.Middleware("iguanas-jewelry"))
	r.Use(middleware.ErrorHandler()) // Add error handler middleware
	r.Use(middleware.ValidateCSRF())

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	r.GET("/images/*filepath", imageStorageHandler.ServeImage) // Local development.
	// Auth routes (no /api/v1 prefix for OAuth)
	auth := r.Group("/auth")
	auth.Use(rateLimitMiddleware)
	{
		auth.GET("/google", authHandlers.GoogleLogin)
		auth.GET("/google/callback", authHandlers.GoogleCallback)
		auth.POST("/logout", authHandlers.Logout)
	}
	r.POST("/webhooks/stripe", paymentHandlers.HandleWebhook)
	// API v1 routes
	api := r.Group("/api/v1")

	// Public routes (no authentication required)
	public := api.Group("")
	{
		public.GET("/products", productHandlers.GetProducts) //also handles images
		public.GET("/products/:id", productHandlers.GetProductByID)
		// public.GET("/products/:id/images")
		// public.GET("/products/images")
	}

	// Protected routes (authentication required)
	protected := api.Group("")
	protected.Use(authMiddleware.RequireAuth())
	{
		user := protected.Group("/user")
		{
			user.GET("/profile", userHandlers.GetMyProfile)
			// user.PUT("/profile", userHandlers.UpdateUserProfile) update profile?
			user.DELETE("/account", userHandlers.DeleteMyAccount)
		}

		cart := protected.Group("/cart")
		{
			cart.GET("", cartsHandlers.GetUserCart)
			cart.POST("", cartsHandlers.AddToCart)
			cart.PUT("/items/:id", cartsHandlers.UpdateCartItem)
			cart.DELETE("/items/:id", cartsHandlers.RemoveFromCart)
			cart.DELETE("", cartsHandlers.ClearCart)
		}

		orders := protected.Group("/orders")
		orders.Use(rateLimitMiddleware)
		{
			orders.GET("", ordersHandlers.ViewOrderHistory)
			orders.GET("/:id", ordersHandlers.GetOrderInfo)
			orders.POST("/checkout", ordersHandlers.CreateOrder) //checkout
			orders.PUT("/:id/cancel", ordersHandlers.CancelOrder)
		}

		payment := protected.Group("/payment")
		{
			payment.POST("/intents/:order_id", paymentHandlers.RetryOrderPayment)
		}

		favorites := protected.Group("/favorites")
		{
			favorites.GET("", userFavoritesHandlers.GetUserFavorites)
			favorites.DELETE("", userFavoritesHandlers.ClearUserFavorites)
		}

		products := protected.Group("/products")
		{
			products.PUT("/:id/favorite", userFavoritesHandlers.AddUserFavorite)
			products.DELETE("/:id/favorite", userFavoritesHandlers.RemoveUserFavorite)
		}
	}

	// Admin routes (authentication + admin role required)
	admin := api.Group("/admin")
	admin.Use(authMiddleware.RequireAuth())
	admin.Use(adminMiddleware.RequireAdmin())
	{
		// Product management
		products := admin.Group("/products")
		{
			products.POST("", productHandlers.PostProducts)
			products.PUT("/:id", productHandlers.UpdateProductByID)
			products.DELETE("/:id", productHandlers.DeleteProductByID)
			products.GET("/:id/images", productImagesHandlers.GetProductImages)
			products.POST("/:id/images", productImagesHandlers.AddProductImage)
			products.DELETE("/:id/images/:imageID", productImagesHandlers.RemoveProductImage)
			products.PUT("/:id/images/:imageID", productImagesHandlers.SetPrimaryImage)
			products.PUT("/:id/images/reorder", productImagesHandlers.ReorderImages)
			products.POST("/:id/images/generate-upload-url", productImagesHandlers.GenerateUploadURL)
			products.GET("/:id/images/generate-upload-url", imageStorageHandler.GetUploadURL)
			products.POST("/:id/images/confirm", productImagesHandlers.ConfirmImageUpload)
		}

		// Image upload (local development only)
		if cfg.Env == "dev" {
			admin.PUT("/uploads/*filepath", imageStorageHandler.HandleLocalUpload)
		}

		// User management
		users := admin.Group("/users")
		{
			users.GET("", userHandlers.GetUsers)
			users.GET("/:id", userHandlers.GetUserByID)
			users.PUT("/:id/role", userHandlers.UpdateUserRole)
			users.DELETE("/:id", userHandlers.DeleteUserByID)
		}

		// Orders management (admin routes)
		orders := admin.Group("/orders")
		{
			// Add your admin order routes here when ready
			orders.GET("", ordersHandlers.GetAllOrders)
			orders.PUT("/:id/status", ordersHandlers.UpdateOrderStatus)
		}
	}
}
