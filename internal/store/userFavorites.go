package store

import (
	"context"
	"fmt"
	"time"

	"github.com/alexalbu001/iguanas-jewelry-api/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserFavoritesStore struct {
	dbpool *pgxpool.Pool
}

func NewUserFavoritesStore(dbpool *pgxpool.Pool) *UserFavoritesStore {
	return &UserFavoritesStore{
		dbpool: dbpool,
	}
}

func (s *UserFavoritesStore) GetUserFavorites(ctx context.Context, userID string) ([]models.UserFavorites, error) {
	sql := `
	SELECT id, user_id, product_id, created_at
	FROM user_favorites
	WHERE user_id=$1
	`
	rows, err := s.dbpool.Query(ctx, sql, userID)
	if err != nil {
		return nil, fmt.Errorf("Error querying user favorites: %w", err)
	}
	defer rows.Close()
	var userFavorites []models.UserFavorites
	for rows.Next() {
		var userFavorite models.UserFavorites
		err := rows.Scan(&userFavorite.ID, &userFavorite.UserID, &userFavorite.ProductID, &userFavorite.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("Error scanning user favorites row: %w", err)
		}
		userFavorites = append(userFavorites, userFavorite)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("Error iterating user favorites rows: %w", err)
	}
	return userFavorites, nil
}

func (s *UserFavoritesStore) AddUserFavorite(ctx context.Context, userID string, productID string) error {
	id := uuid.NewString()
	sql := `
	INSERT INTO user_favorites (id, user_id, product_id, created_at)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (user_id, product_id) DO NOTHING
	`
	_, err := s.dbpool.Exec(ctx, sql, id, userID, productID, time.Now())
	if err != nil {
		return fmt.Errorf("Error adding user favorite: %w", err)
	}
	return nil
}

func (s *UserFavoritesStore) RemoveUserFavorite(ctx context.Context, userID string, productID string) error {
	sql := `
	DELETE FROM user_favorites
	WHERE user_id=$1 AND product_id=$2
	`
	_, err := s.dbpool.Exec(ctx, sql, userID, productID)
	if err != nil {
		return fmt.Errorf("Error removing user favorite: %w", err)
	}
	return nil
}

func (s *UserFavoritesStore) ClearUserFavorites(ctx context.Context, userID string) error {
	sql := `
	DELETE FROM user_favorites
	WHERE user_id=$1
	`
	_, err := s.dbpool.Exec(ctx, sql, userID)
	if err != nil {
		return fmt.Errorf("Error deleting user favorites: %w", err)
	}
	return nil
}
