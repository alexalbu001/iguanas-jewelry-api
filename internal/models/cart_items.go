package models

import "time"

type CartItems struct {
	ID        string    `json:"id"`
	ProductID string    `json:"product_id"`
	CartID    string    `json:"cart_id"`
	Quantity  int       `json:"quantity"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
