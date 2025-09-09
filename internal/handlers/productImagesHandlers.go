package handlers

import (
	"net/http"

	customerrors "github.com/alexalbu001/iguanas-jewelry-api/internal/customErrors"
	"github.com/alexalbu001/iguanas-jewelry-api/internal/models"
	"github.com/alexalbu001/iguanas-jewelry-api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ProductImagesHandlers struct {
	productImagesService *service.ProductImagesService
}

func NewProductImagesHandlers(productImagesService *service.ProductImagesService) *ProductImagesHandlers {
	return &ProductImagesHandlers{
		productImagesService: productImagesService,
	}
}

// @Summary Add product image
// @Description Add a new image to a product (Admin only)
// @Tags product-images
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Param image body models.ProductImage true "Product image information"
// @Success 201 {object} models.ProductImage
// @Failure 400 {object} responses.ErrorResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 403 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v1/admin/products/{id}/images [post]
func (p *ProductImagesHandlers) AddProductImage(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.Error(&customerrors.ErrUserNotFound)
		return
	}

	productID := c.Param("id")
	if err := uuid.Validate(productID); err != nil {
		c.Error(&customerrors.ErrInvalidUUID)
		return
	}

	var newImage models.ProductImage
	if err := c.ShouldBindBodyWithJSON(&newImage); err != nil {
		c.Error(&customerrors.ErrInvalidJSON)
		return
	}

	insertedImage, err := p.productImagesService.InsertProductImage(c.Request.Context(), newImage, userID.(string), productID)
	if err != nil {
		c.Error(err)
		return
	}

	c.IndentedJSON(http.StatusCreated, insertedImage)
}

// @Summary Remove product image
// @Description Remove an image from a product (Admin only)
// @Tags product-images
// @Produce json
// @Param id path string true "Product ID"
// @Param imageID path string true "Image ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} responses.ErrorResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 403 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v1/admin/products/{id}/images/{imageID} [delete]
func (p *ProductImagesHandlers) RemoveProductImage(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.Error(&customerrors.ErrUserNotFound)
		return
	}

	productID := c.Param("id")
	if err := uuid.Validate(productID); err != nil {
		c.Error(&customerrors.ErrInvalidUUID)
		return
	}

	imageID := c.Param("imageID")
	if err := uuid.Validate(imageID); err != nil {
		c.Error(&customerrors.ErrInvalidUUID)
		return
	}

	err := p.productImagesService.DeleteProductImage(c.Request.Context(), imageID, productID, userID.(string))
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product image deleted successfully", "id": imageID})
}

// @Summary Set primary image
// @Description Set an image as the primary image for a product (Admin only)
// @Tags product-images
// @Produce json
// @Param id path string true "Product ID"
// @Param imageID path string true "Image ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} responses.ErrorResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 403 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v1/admin/products/{id}/images/{imageID} [put]
func (p *ProductImagesHandlers) SetPrimaryImage(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.Error(&customerrors.ErrUserNotFound)
		return
	}

	productID := c.Param("id")
	if err := uuid.Validate(productID); err != nil {
		c.Error(&customerrors.ErrInvalidUUID)
		return
	}

	imageID := c.Param("imageID")
	if err := uuid.Validate(imageID); err != nil {
		c.Error(&customerrors.ErrInvalidUUID)
		return
	}

	err := p.productImagesService.SetPrimaryImage(c.Request.Context(), productID, imageID, userID.(string))
	if err != nil {
		c.Error(err)
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "Product image set as main successfully", "id": imageID})
}

// @Summary Reorder product images
// @Description Change the order of images for a product (Admin only)
// @Tags product-images
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Param imageOrder body []string true "Array of image IDs in desired order"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} responses.ErrorResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 403 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v1/admin/products/{id}/images/reorder [put]
func (p *ProductImagesHandlers) ReorderImages(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.Error(&customerrors.ErrUserNotFound)
		return
	}

	productID := c.Param("id")
	if err := uuid.Validate(productID); err != nil {
		c.Error(&customerrors.ErrInvalidUUID)
		return
	}

	var imageOrder []string
	if err := c.ShouldBindJSON(&imageOrder); err != nil {
		c.Error(&customerrors.ErrInvalidJSON)
		return
	}

	err := p.productImagesService.ReorderImages(c.Request.Context(), productID, imageOrder, userID.(string))
	if err != nil {
		c.Error(err)
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "Product image order changed successfully", "id": productID})
}
