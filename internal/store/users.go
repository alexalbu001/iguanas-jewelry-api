package store

import (
	"context"
	"fmt"
	"time"

	"github.com/alexalbu001/iguanas-jewelry/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UsersStore struct {
	dbpool *pgxpool.Pool
}

func (h *UsersStore) GetUsers(ctx context.Context) ([]models.User, error) {
	sql := `
	SELECT id, googleid, email, name, role, created_at, updated_at
	FROM users
	ORDER BY created_at DESC
	`

	rows, err := h.dbpool.Query(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("Error querying users: %w", err)
	}
	defer rows.Close()
	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID,
			&user.GoogleID,
			&user.Email,
			&user.Name,
			&user.Role,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("Error scanning users row: %w", err)
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("Error iterating users rows: %w", err)
	}
	return users, nil
}

func (h *UsersStore) GetUserByGoogleID(ctx context.Context, googleID string) (models.User, error) {
	sql := `
SELECT id, googleid, email, name, role, created_at, updated_at
FROM users
WHERE googleid=$1`

	row := h.dbpool.QueryRow(ctx, sql, googleID)

	var user models.User
	err := row.Scan(&user.ID, &user.GoogleID, &user.Email, &user.Name, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return models.User{}, fmt.Errorf("User not found with id %s", googleID)
		}
		return models.User{}, fmt.Errorf("Error scanning user row: %w", err)
	}

	return user, nil
}

func (h *UsersStore) GetUserByID(ctx context.Context, id string) (models.User, error) {
	sql := `
SELECT id, googleid, email, name, role, created_at, updated_at
FROM users
WHERE id=$1`

	row := h.dbpool.QueryRow(ctx, sql, id)

	var user models.User
	err := row.Scan(&user.ID, &user.GoogleID, &user.Email, &user.Name, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return models.User{}, fmt.Errorf("User not found with id %s", id)
		}
		return models.User{}, fmt.Errorf("Error scanning user row: %w", err)
	}

	return user, nil
}

func (h *UsersStore) AddUser(ctx context.Context, user models.User) (models.User, error) {
	user.ID = uuid.NewString()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	sql := `
	INSERT INTO users (id, googleid, email, name, role, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := h.dbpool.Exec(ctx, sql, user.ID, user.GoogleID, user.Email, user.Name, user.Role, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return models.User{}, fmt.Errorf("User could not be created, %w", err)
	}
	return user, nil
}

func (h *UsersStore) DeleteUser(ctx context.Context, id string) error {
	sql := `
	DELETE FROM users
	WHERE id=$1`

	commandTag, err := h.dbpool.Exec(ctx, sql, id)
	if err != nil {
		return fmt.Errorf("Error deleting user: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("User not found with id: %s", id)
	}
	return nil
}

func (h *UsersStore) UpdateUser(ctx context.Context, id string, user models.User) (models.User, error) {
	user.UpdatedAt = time.Now()
	sql := `
	UPDATE users
	SET name=$1, role=$2, updated_at=$3
	WHERE id=$4
	RETURNING id,google_id,email, created_at`

	row := h.dbpool.QueryRow(ctx, sql, user.Name, user.Role, user.UpdatedAt, id)

	var newUser models.User

	err := row.Scan(&newUser.ID, &newUser.GoogleID, &newUser.Email, &newUser.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return models.User{}, fmt.Errorf("No user with id: %s", id)
		}
		return models.User{}, fmt.Errorf("Error scanning row: %w", err)
	}
	newUser.Name = user.Name
	newUser.Role = user.Role
	newUser.UpdatedAt = user.UpdatedAt

	return newUser, nil
}

func (h *UsersStore) UpdateUserRole(ctx context.Context, id string, role string) error {
	updatedAt := time.Now()
	sql := `
	UPDATE users
	SET role=$1, updated_at=$2
	WHERE id=$3
	`
	commandTag, err := h.dbpool.Exec(ctx, sql, role, updatedAt, id)
	if err != nil {
		return fmt.Errorf("Error updating user role: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("User not found with id: %s", id)
	}
	return nil
}

func NewUsersStore(connection *pgxpool.Pool) *UsersStore {
	return &UsersStore{
		dbpool: connection,
	}
}
