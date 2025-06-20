package models

import "time"

type User struct {
	ID        string    `json:"id"`
	GoogleID  string    `json:"googleid"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
