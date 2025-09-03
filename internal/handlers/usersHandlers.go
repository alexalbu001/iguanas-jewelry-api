package handlers

import (
	"net/http"

	customerrors "github.com/alexalbu001/iguanas-jewelry-api/internal/customErrors"
	"github.com/alexalbu001/iguanas-jewelry-api/internal/models"
	"github.com/alexalbu001/iguanas-jewelry-api/internal/service"
	"github.com/gin-gonic/gin"
)

type UserHandlers struct {
	UserService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandlers {
	return &UserHandlers{
		UserService: userService,
	}
}

// @Summary Get a list of all users
// @Description Fetches all users
// @Tags users
// @Produce json
// @Success 200 {array} models.User
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v1/users [get]
func (u *UserHandlers) GetUsers(c *gin.Context) {
	users, err := u.UserService.GetUsers(c.Request.Context())
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, users)
}

// @Summary Get user by ID
// @Description Retrieves a specific user by their ID
// @Tags users
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} models.User
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v1/users/{id} [get]
func (u *UserHandlers) GetUserByID(c *gin.Context) {
	id := c.Param("id")
	user, err := u.UserService.GetUserByID(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}
	c.IndentedJSON(http.StatusOK, user)
}

// @Summary Update user by ID
// @Description Updates an existing user's information
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param user body models.User true "User information to update"
// @Success 200 {object} models.User
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v1/users/{id} [put]
func (u *UserHandlers) UpdateUserByID(c *gin.Context) {
	id := c.Param("id")

	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.Error(&customerrors.ErrInvalidJSON)
		return
	}

	updatedUser, err := u.UserService.UpdateUserByID(c.Request.Context(), id, user)
	if err != nil {
		c.Error(err)
		return
	}
	c.IndentedJSON(http.StatusOK, updatedUser)
}

// @Summary Delete user by ID
// @Description Removes a user from the system
// @Tags users
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v1/users/{id} [delete]
func (u *UserHandlers) DeleteUserByID(c *gin.Context) {
	id := c.Param("id")
	err := u.UserService.DeleteUserByID(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully", "id": id})
}

// @Summary Update user role
// @Description Updates a user's role (admin or customer)
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param roleUpdate body object true "Role update request" schema(object{role=string})
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v1/users/{id}/role [put]
func (u *UserHandlers) UpdateUserRole(c *gin.Context) {
	id := c.Param("id")

	var roleUpdate struct {
		Role string `json:"role" binding:"required,oneof=admin customer"`
	}
	if err := c.ShouldBindJSON(&roleUpdate); err != nil {
		c.Error(&customerrors.ErrInvalidJSON)
		return
	}

	err := u.UserService.UpdateUserRole(c.Request.Context(), id, roleUpdate.Role)
	if err != nil {
		c.Error(err)
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{
		"message":  "User role updated successfully",
		"user_id":  id,
		"new_role": roleUpdate.Role,
	})
}

// @Summary Get my profile
// @Description Retrieves the authenticated user's profile information
// @Tags users
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} models.User
// @Failure 401 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v1/users/profile [get]
func (u *UserHandlers) GetMyProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.Error(&customerrors.ErrUserNotFound)
		return
	}

	user, err := u.UserService.GetUserByID(c.Request.Context(), userID.(string))
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, user)
}

// @Summary Delete my account
// @Description Allows authenticated users to delete their own account
// @Tags users
// @Produce json
// @Security ApiKeyAuth
// @Success 202 {object} map[string]interface{}
// @Failure 401 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v1/users/profile [delete]
func (u *UserHandlers) DeleteMyAccount(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.Error(&customerrors.ErrUserNotFound)
		return
	}

	err := u.UserService.DeleteUserByID(c.Request.Context(), userID.(string))
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"user_id": userID,
		"message": "The account has been deleted",
	})
}
