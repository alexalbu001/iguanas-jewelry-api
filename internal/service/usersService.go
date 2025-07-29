package service

import (
	"context"
	"fmt"

	"github.com/alexalbu001/iguanas-jewelry/internal/models"
	"github.com/google/uuid"
)

type UsersStore interface {
	GetUsers(ctx context.Context) ([]models.User, error)
	GetUserByGoogleID(ctx context.Context, googleID string) (models.User, error)
	GetUserByID(ctx context.Context, id string) (models.User, error)
	AddUser(ctx context.Context, user models.User) (models.User, error)
	DeleteUser(ctx context.Context, id string) error
	UpdateUser(ctx context.Context, id string, user models.User) (models.User, error)
	UpdateUserRole(ctx context.Context, id string, role string) error
}

type UserService struct {
	UserStore UsersStore
}

func NewUserService(userStore UsersStore) *UserService {
	return &UserService{
		UserStore: userStore,
	}
}

func (u *UserService) GetUsers(ctx context.Context) ([]models.User, error) {
	users, err := u.UserStore.GetUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch users: %w", err)
	}

	return users, nil
}

func (u *UserService) GetUserByID(ctx context.Context, userID string) (models.User, error) {
	err := uuid.Validate(userID)
	if err != nil {
		return models.User{}, fmt.Errorf("User id is invalid")
	}
	user, err := u.UserStore.GetUserByID(ctx, userID)
	if err != nil {
		return models.User{}, fmt.Errorf("No user with this id found")
	}
	return user, nil
}

func (u *UserService) UpdateUserByID(ctx context.Context, userID string, user models.User) (models.User, error) {
	err := uuid.Validate(userID)
	if err != nil {
		return models.User{}, fmt.Errorf("User id is invalid")
	}
	if user.Name == "" {
		return models.User{}, fmt.Errorf("User name can't be empty")
	}

	updatedUser, err := u.UserStore.UpdateUser(ctx, userID, user)
	if err != nil {
		return models.User{}, fmt.Errorf("Failed to update user, %w", err)
	}
	return updatedUser, nil
}

func (u *UserService) UpdateUserRole(ctx context.Context, userID, role string) error {
	err := uuid.Validate(userID)
	if err != nil {
		fmt.Errorf("User id is invalid")
	}
	if role != "admin" && role != "customer" {
		fmt.Errorf("User role must be either admin or customer")
	}

	err = u.UserStore.UpdateUserRole(ctx, userID, role)
	if err != nil {
		fmt.Errorf("Failed to update user, %w", err)
	}
	return nil
}

func (u *UserService) DeleteUserByID(ctx context.Context, userID string) error {
	err := uuid.Validate(userID)
	if err != nil {
		fmt.Errorf("User id is invalid")
	}
	err = u.UserStore.DeleteUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("Failed to delete user: %w", err)
	}
	return nil
}

func (u *UserService) AddUser(ctx context.Context, user models.User) (models.User, error) {
	if user.Name == "" {
		return models.User{}, fmt.Errorf("User name can't be empty")
	}
	if user.GoogleID == "" {
		return models.User{}, fmt.Errorf("Failed to create user with missing google id")
	}
	if user.Email == "" {
		return models.User{}, fmt.Errorf("Failed to create user with missing google id")
	}
	createdUser, err := u.UserStore.AddUser(ctx, user)
	if err != nil {
		return models.User{}, fmt.Errorf("Failed to create user: %w", err)
	}
	return createdUser, nil
}
