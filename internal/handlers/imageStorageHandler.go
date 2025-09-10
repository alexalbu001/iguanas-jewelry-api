package handlers

import (
	"net/http"

	"github.com/alexalbu001/iguanas-jewelry-api/internal/storage"
	"github.com/gin-gonic/gin"
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
