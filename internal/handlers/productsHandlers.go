package handlers

import (
	"net/http"

	customerrors "github.com/alexalbu001/iguanas-jewelry/internal/customErrors"
	"github.com/alexalbu001/iguanas-jewelry/internal/models"
	"github.com/alexalbu001/iguanas-jewelry/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ProductHandlers struct {
	ProductHandler *service.ProductsService
}

func NewProductHandlers(productHandler *service.ProductsService) *ProductHandlers {
	return &ProductHandlers{
		ProductHandler: productHandler,
	}
}

// @Summary Fetch all products
// @Description Retrieves a list of all available products
// @Tags products
// @Produce json
// @Success 200 {array} models.Product
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v1/products [get]
func (h *ProductHandlers) GetProducts(c *gin.Context) {
	products, err := h.ProductHandler.GetProducts(c.Request.Context())
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, products)
}

// @Summary Post a Product
// @Description Creates a product based on json input
// @Tags products
// @Accept json
// @Produce json
// @Param product body models.Product true "Product information"
// @Success 201 {object} models.Product
// @Failure 400 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v1/products [post]
func (h *ProductHandlers) PostProducts(c *gin.Context) {
	var newProduct models.Product
	if err := c.ShouldBindJSON(&newProduct); err != nil {
		c.Error(&customerrors.ErrInvalidJSON)
		return
	}
	createdProduct, err := h.ProductHandler.PostProduct(c.Request.Context(), newProduct)
	if err != nil {
		c.Error(err)
		return
	}
	c.IndentedJSON(http.StatusCreated, createdProduct)
}

// @Summary Show a product
// @Description Retrieve a product by ID
// @Tags products
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} models.Product
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v1/products/{id} [get]
func (h *ProductHandlers) GetProductByID(c *gin.Context) {
	id := c.Param("id")
	if err := uuid.Validate(id); err != nil {
		c.Error(&customerrors.ErrInvalidUUID)
		return
	}

	foundProduct, err := h.ProductHandler.GetProductByID(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}
	c.IndentedJSON(http.StatusOK, foundProduct)
}

// @Summary Change a product
// @Description Update a product
// @Tags products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Param product body models.Product true "Product information to update"
// @Success 200 {object} models.Product
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v1/products/{id} [put]
func (h *ProductHandlers) UpdateProductByID(c *gin.Context) {
	id := c.Param("id")
	if err := uuid.Validate(id); err != nil {
		c.Error(&customerrors.ErrInvalidUUID)
		return
	}
	var newProduct models.Product
	if err := c.ShouldBindJSON(&newProduct); err != nil {
		c.Error(&customerrors.ErrInvalidJSON)
		return
	}
	updatedProduct, err := h.ProductHandler.UpdateProductByID(c.Request.Context(), id, newProduct)
	if err != nil {
		c.Error(err)
		return
	}
	c.IndentedJSON(http.StatusOK, updatedProduct)
}

// @Summary Remove a product
// @Description Remove a product
// @Tags products
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v1/products/{id} [delete]
func (h *ProductHandlers) DeleteProductByID(c *gin.Context) {
	id := c.Param("id")
	if err := uuid.Validate(id); err != nil {
		c.Error(&customerrors.ErrInvalidUUID)
		return
	}
	err := h.ProductHandler.DeleteProductByID(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully", "id": id})
}
