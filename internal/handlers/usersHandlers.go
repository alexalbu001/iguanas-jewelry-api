package handlers

import (
	"net/http"

	"github.com/alexalbu001/iguanas-jewelry/internal/models"
	"github.com/alexalbu001/iguanas-jewelry/internal/responses"
	"github.com/alexalbu001/iguanas-jewelry/internal/service"
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

func (u *UserHandlers) GetUsers(c *gin.Context) {
	users, err := u.UserService.GetUsers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}

func (u *UserHandlers) GetUserByID(c *gin.Context) {
	id := c.Param("id")
	user, err := u.UserService.GetUserByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"Error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, user)
}

func (u *UserHandlers) UpdateUserByID(c *gin.Context) {
	id := c.Param("id")

	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	updatedUser, err := u.UserService.UpdateUserByID(c.Request.Context(), id, user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"Error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, updatedUser)
}

func (u *UserHandlers) DeleteUserByID(c *gin.Context) {
	id := c.Param("id")
	err := u.UserService.DeleteUserByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully", "id": id})
}

func (u *UserHandlers) UpdateUserRole(c *gin.Context) {
	id := c.Param("id")

	var roleUpdate struct {
		Role string `json:"role" binding:"required,oneof=admin customer"`
	}
	if err := c.ShouldBindJSON(&roleUpdate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	err := u.UserService.UpdateUserRole(c.Request.Context(), id, roleUpdate.Role)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"Error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{
		"message":  "User role updated successfully",
		"user_id":  id,
		"new_role": roleUpdate.Role,
	})
}

func (u *UserHandlers) GetMyProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	user, err := u.UserService.GetUserByID(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"Error": err.Error()})
		return
	}

	profileResponse := responses.UserProfileResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.String(),
	}
	c.JSON(http.StatusOK, profileResponse)
}

func (u *UserHandlers) DeleteMyAccount(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	u.UserService.DeleteUserByID(c.Request.Context(), userID.(string))
}
