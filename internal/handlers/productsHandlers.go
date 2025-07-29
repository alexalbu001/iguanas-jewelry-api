package handlers

import (
	"net/http"

	"github.com/alexalbu001/iguanas-jewelry/internal/models"
	"github.com/alexalbu001/iguanas-jewelry/internal/service"
	"github.com/gin-gonic/gin"
)

type ProductHandlers struct {
	ProductHandler *service.ProductsService
}

func NewProductHandlers(productHandler *service.ProductsService) *ProductHandlers {
	return &ProductHandlers{
		ProductHandler: productHandler,
	}
}

func (h *ProductHandlers) GetProducts(c *gin.Context) {
	products, err := h.ProductHandler.GetProducts(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, products)
}

func (h *ProductHandlers) PostProducts(c *gin.Context) {
	var newProduct models.Product
	if err := c.ShouldBindJSON(&newProduct); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	createdProduct, err := h.ProductHandler.PostProduct(c.Request.Context(), newProduct)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusCreated, createdProduct)
}

// func PostProducts(c *gin.Context) {

// 	// Call BindJSON to bind the received JSON to newProduct
// 	var newProduct models.Product
// 	if err := c.BindJSON(&newProduct); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	ProductHandler.Products = append(ProductHandler.Products, newProduct)
// 	c.IndentedJSON(http.StatusCreated, newProduct)
// }

func (h *ProductHandlers) GetProductByID(c *gin.Context) {
	id := c.Param("id")

	foundProduct, err := h.ProductHandler.GetProductByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"Error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, foundProduct)
}

// func GetProductByID(c *gin.Context) {
// 	id := c.Param("id")

// 	for _, p := range ProductHandler.Products {
// 		if p.ID == id {
// 			c.IndentedJSON(http.StatusOK, p)
// 			return
// 		}
// 	}
// 	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "product not found"})
// }

func (h *ProductHandlers) UpdateProductByID(c *gin.Context) {
	id := c.Param("id")
	var newProduct models.Product
	if err := c.ShouldBindJSON(&newProduct); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	updatedProduct, err := h.ProductHandler.UpdateProductByID(c.Request.Context(), id, newProduct)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"Error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, updatedProduct)
}

// func UpdateProductsByID(c *gin.Context) {
// 	id := c.Param("id")

// 	for _, p := range ProductHandler.Products {
// 		if p.ID == id {

// 		}
// 	}
// }

func (h *ProductHandlers) DeleteProductByID(c *gin.Context) {
	id := c.Param("id")
	err := h.ProductHandler.DeleteProductByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully", "id": id})
}

// func DeleteProductsByID(c *gin.Context) {

// }
