package handlers

import (
	"net/http"

	"github.com/alexalbu001/iguanas-jewelry/internal/models"
	"github.com/gin-gonic/gin"
)

type UsersStore interface {
	GetUsers() ([]models.User, error)
	GetUserByGoogleID(googleID string) (models.User, error)
	GetUserByID(id string) (models.User, error)
	AddUser(user models.User) (models.User, error)
	DeleteUser(id string) error
	UpdateUser(id string, user models.User) (models.User, error)
	UpdateUserRole(id string, role string) error
}

type UserHandlers struct {
	User UsersStore
}

func (u *UserHandlers) GetUsers(c *gin.Context) {
	users, err := u.User.GetUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}

func (u *UserHandlers) GetUserByID(c *gin.Context) {
	id := c.Param("id")
	user, err := u.User.GetUserByID(id)
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

	updatedUser, err := u.User.UpdateUser(id, user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"Error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, updatedUser)
}

func (u *UserHandlers) DeleteUserByID(c *gin.Context) {
	id := c.Param("id")
	err := u.User.DeleteUser(id)
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

	err := u.User.UpdateUserRole(id, roleUpdate.Role)
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

func NewUserHandler(userStore UsersStore) *UserHandlers {
	return &UserHandlers{
		User: userStore,
	}
}
