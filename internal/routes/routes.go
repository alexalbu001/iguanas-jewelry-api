package routes

import (
	"github.com/alexalbu001/iguanas-jewelry/internal/handlers"
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

func SetupRoutes(r *gin.Engine, h *handlers.ProductHandlers) {
	api := r.Group("/api/v1")
	{
		api.GET("/products", h.GetProducts)
		api.GET("/products/:id", h.GetProductByID)
		api.POST("/products", h.PostProducts)
		api.PUT("/products/:id", h.UpdateProductByID)
		api.DELETE("/products/:id", h.DeleteProductByID)
	}
}
