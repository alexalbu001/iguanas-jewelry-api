package handlers

import (
	"fmt"
	"net/http"

	"github.com/alexalbu001/iguanas-jewelry-api/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ImageStorageHandlers struct {
	imageStorage storage.ImageStorage
}

func NewImageStorageHandlers(imageStorage storage.ImageStorage) *ImageStorageHandlers {
	return &ImageStorageHandlers{
		imageStorage: imageStorage,
	}
}

// @Summary Serve image
// @Description Serve an image file from storage
// @Tags images
// @Produce image/*
// @Param filepath path string true "Image file path"
// @Success 200 {file} binary "Image file"
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /images/{filepath} [get]
func (h *ImageStorageHandlers) ServeImage(c *gin.Context) {
	filepath := c.Param("filepath")

	data, err := h.imageStorage.Get(c.Request.Context(), filepath)
	if err != nil {
		c.Error(err)
		return
	}

	contentType := http.DetectContentType(data)
	c.Header("Content-Type", contentType)
	c.Data(http.StatusOK, contentType, data)
}

// @Summary Upload image (Local Development)
// @Description Upload an image file to local storage (Admin only, Local Development only)
// @Tags images
// @Accept image/*
// @Produce json
// @Param filepath path string true "Image file path"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} responses.ErrorResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 403 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v1/admin/uploads/{filepath} [put]
func (h *ImageStorageHandlers) HandleLocalUpload(c *gin.Context) {
	filepath := c.Param("filepath")

	data, err := c.GetRawData()
	if err != nil {
		c.Error(err)
		return
	}

	err = h.imageStorage.Store(c.Request.Context(), filepath, data)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Image uploaded successfully",
		"filepath": filepath,
	})
}

// @Summary Generate upload URL
// @Description Generate a presigned URL for uploading images to storage
// @Tags images
// @Accept json
// @Produce json
// @Param content_type query string true "Content type of the image (e.g., image/jpeg, image/png)"
// @Param product_id query string false "Product ID for organizing images"
// @Param type query string false "Image type (main, thumbnail, gallery)" default(main)
// @Success 200 {object} map[string]interface{} "Upload URL and metadata"
// @Failure 400 {object} responses.ErrorResponse "Bad request - missing required parameters"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /api/v1/images/upload-url [get]
func (h *ImageStorageHandlers) GetUploadURL(c *gin.Context) {
	contentType := c.Query("content_type")
	if contentType == "" {
		c.JSON(400, gin.H{"error": "content_type is required"})
		return
	}

	productID := c.Query("product_id")          // optional
	imageType := c.DefaultQuery("type", "main") // default to main image

	// Generate unique key
	key := fmt.Sprintf("products/%s/%s/%s.%s",
		productID, imageType, uuid.New().String(), contentType)

	url, err := h.imageStorage.GenerateUploadURL(c.Request.Context(), key, contentType)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(200, gin.H{
		"upload_url": url,
		"key":        key,
		"expires_in": 900, // 15 minutes
	})
}
