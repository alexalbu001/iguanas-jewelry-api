package routes

import (
	"github.com/alexalbu001/iguanas-jewelry/internal/auth"
	"github.com/alexalbu001/iguanas-jewelry/internal/handlers"
	"github.com/alexalbu001/iguanas-jewelry/internal/middleware"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, productHandlers *handlers.ProductHandlers, userHandlers *handlers.UserHandlers, cartsHandlers *handlers.CartsHandlers, ordersHandlers *handlers.OrdersHandlers, paymentHandlers *handlers.PaymentHandler, authHandlers *auth.AuthHandlers, authMiddleware *middleware.AuthMiddleware, adminMiddleware *middleware.AdminMiddleware, loggingMiddleware *middleware.LoggingMiddleware) {
	r.Use(loggingMiddleware.RequestLogging())
	// Auth routes (no /api/v1 prefix for OAuth)
	auth := r.Group("/auth")
	{
		auth.GET("/google", authHandlers.GoogleLogin)
		auth.GET("/google/callback", authHandlers.GoogleCallback)
	}
	r.POST("/webhooks/stripe", paymentHandlers.HandleWebhook)
	// API v1 routes
	api := r.Group("/api/v1")

	// Public routes (no authentication required)
	public := api.Group("")
	{
		public.GET("/products", productHandlers.GetProducts)
		public.GET("/products/:id", productHandlers.GetProductByID)
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
