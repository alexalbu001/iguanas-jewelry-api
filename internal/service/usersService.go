package service

import (
	"fmt"

	"github.com/alexalbu001/iguanas-jewelry/internal/models"
	"github.com/google/uuid"
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

type UserService struct {
	UserStore UsersStore
}

func NewUserService(userStore UsersStore) *UserService {
	return &UserService{
		UserStore: userStore,
	}
}

func (u *UserService) GetUsers() ([]models.User, error) {
	users, err := u.GetUsers()
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch users: %w", err)
	}

	return users, nil
}

func (u *UserService) GetUserByID(userID string) (models.User, error) {
	err := uuid.Validate(userID)
	if err != nil {
		return models.User{}, fmt.Errorf("User id is invalid")
	}
	user, err := u.GetUserByID(userID)
	if err != nil {
		return models.User{}, fmt.Errorf("No user with this id found")
	}
	return user, nil
}

func (u *UserService) UpdateUserByID(userID string, user models.User) (models.User, error) {
	err := uuid.Validate(userID)
	if err != nil {
		return models.User{}, fmt.Errorf("User id is invalid")
	}
	if user.Name == "" {
		return models.User{}, fmt.Errorf("User name can't be empty")
	}

	updatedUser, err := u.UpdateUserByID(userID, user)
	if err != nil {
		return models.User{}, fmt.Errorf("Failed to update user, %w", err)
	}
	return updatedUser, nil
}

func (u *UserService) UpdateUserRole(userID, role string) error {
	err := uuid.Validate(userID)
	if err != nil {
		fmt.Errorf("User id is invalid")
	}
	if role != "admin" && role != "customer" {
		fmt.Errorf("User role must be either admin or customer")
	}

	err = u.UpdateUserRole(userID, role)
	if err != nil {
		fmt.Errorf("Failed to update user, %w", err)
	}
	return nil
}

func (u *UserService) DeleteUserByID(userID string) error {
	err := uuid.Validate(userID)
	if err != nil {
		fmt.Errorf("User id is invalid")
	}
	err = u.DeleteUserByID(userID)
	if err != nil {
		return fmt.Errorf("Failed to delete user: %w", err)
	}
	return nil
}

func (u *UserService) AddUser(user models.User) (models.User, error) {
	if user.Name == "" {
		return models.User{}, fmt.Errorf("User name can't be empty")
	}
	if user.GoogleID == "" {
		return models.User{}, fmt.Errorf("Failed to create user with missing google id")
	}
	if user.Email == "" {
		return models.User{}, fmt.Errorf("Failed to create user with missing google id")
	}
	createdUser, err := u.AddUser(user)
	if err != nil {
		return models.User{}, fmt.Errorf("Failed to create user: %w", err)
	}
	return createdUser, nil
}
