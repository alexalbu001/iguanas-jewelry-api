package routes

import (
	"github.com/alexalbu001/iguanas-jewelry/internal/auth"
	"github.com/alexalbu001/iguanas-jewelry/internal/handlers"
	"github.com/alexalbu001/iguanas-jewelry/internal/middleware"
	"github.com/gin-gonic/gin"
)

// func main() {

// 	r := gin.Default()
// 	r.GET("/ping", func(c *gin.Context) {
// 		c.JSON(200, gin.H{
// 			"message": "pong",
// 		})
// 	})

// 	r.GET("/", func(c *gin.Context) {
// 		c.JSON(200, gin.H{
// 			"message": "Hello World",
// 		})
// 	})

// 	r.GET("/api/v1/products", handlers.ProductHandlers.GetAll)
// 	r.GET("/api/v1/products:id", store.GetProductByID)
// 	r.POST("/api/v1/products", store.PostProducts)
// 	r.Run(":" + os.Getenv("PORT"))
// }

func SetupRoutes(r *gin.Engine, h *handlers.ProductHandlers, u *handlers.UserHandlers, c *handlers.CartsHandlers, a *auth.AuthHandlers, m *middleware.AuthMiddleware, n *middleware.AdminMiddleware) {
	api := r.Group("/api/v1")
	{
		api.GET("/products", h.GetProducts)
		api.GET("/products/:id", h.GetProductByID)
		api.POST("/products", m.RequireAuth(), n.RequireAdmin(), h.PostProducts)
		api.PUT("/products/:id", m.RequireAuth(), n.RequireAdmin(), h.UpdateProductByID)
		api.DELETE("/products/:id", m.RequireAuth(), n.RequireAdmin(), h.DeleteProductByID)
		// Users
		api.GET("/users", m.RequireAuth(), n.RequireAdmin(), u.GetUsers)
		api.GET("/users/:id", m.RequireAuth(), n.RequireAdmin(), u.GetUserByID)
		api.PUT("/users/:id/role", m.RequireAuth(), n.RequireAdmin(), u.UpdateUserRole)
		api.DELETE("/user/:id", m.RequireAuth(), n.RequireAdmin(), u.DeleteUserByID)
		// Carts
		api.GET("/cart", m.RequireAuth(), c.GetUserCart)
		api.POST("/cart", m.RequireAuth(), c.AddToCart)
		api.PUT("/cart/items/:item_id", m.RequireAuth(), c.UpdateCartItem)
		api.DELETE("/cart/items/:item_id", m.RequireAuth(), c.RemoveFromCart)
		api.DELETE("/cart", m.RequireAuth(), c.ClearCart)
	}
	r.GET("/auth/google", a.GoogleLogin)
	r.GET("/auth/google/callback", a.GoogleCallback)

}
