package handlers

import (
	"net/http"

	customerrors "github.com/alexalbu001/iguanas-jewelry-api/internal/customErrors"
	"github.com/alexalbu001/iguanas-jewelry-api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserFavoritesHandlers struct {
	UserFavoritesService *service.UserFavoritesService
}

func NewUserFavoritesHandlers(userFavoritesService *service.UserFavoritesService) *UserFavoritesHandlers {
	return &UserFavoritesHandlers{
		UserFavoritesService: userFavoritesService,
	}
}

// @Summary Get user favorites
// @Description Retrieves the current user's favorites
// @Tags user-favorites
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {array} models.Product
// @Failure 401 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v1/user-favorites [get]
func (u *UserFavoritesHandlers) GetUserFavorites(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.Error(&customerrors.ErrUserNotFound)
		return
	}
	products, err := u.UserFavoritesService.GetUserFavorites(c.Request.Context(), userID.(string))
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, products)
}

// @Summary Add user favorite
// @Description Add a product to the current user's favorites
// @Tags user-favorites
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} responses.ErrorResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v1/user-favorites/{id} [post]
func (u *UserFavoritesHandlers) AddUserFavorite(c *gin.Context) {
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
	err := u.UserFavoritesService.AddUserFavorite(c.Request.Context(), userID.(string), productID)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Product added to favorites"})
}

// @Summary Remove user favorite
// @Description Remove a product from the current user's favorites
// @Tags user-favorites
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} responses.ErrorResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v1/user-favorites/{id} [delete]
func (u *UserFavoritesHandlers) RemoveUserFavorite(c *gin.Context) {
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
	err := u.UserFavoritesService.RemoveUserFavorite(c.Request.Context(), userID.(string), productID)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Product removed from favorites"})
}

// @Summary Clear user favorites
// @Description Remove all products from the current user's favorites
// @Tags user-favorites
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v1/user-favorites [delete]
func (u *UserFavoritesHandlers) ClearUserFavorites(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.Error(&customerrors.ErrUserNotFound)
		return
	}
	err := u.UserFavoritesService.ClearUserFavorites(c.Request.Context(), userID.(string))
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User favorites cleared"})
}
