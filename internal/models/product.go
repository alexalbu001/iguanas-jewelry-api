package models

import "time"

// The Product struct shouldn't know anything about databases.
type Product struct {
	ID            string    `json:"id"`
	Name          string    `json:"name" binding:"required,min=1,max=100"`
	Price         float64   `json:"price" binding:"required,gt=0"`
	Description   string    `json:"description" binding:"max=500"`
	Category      string    `json:"category" binding:"required,oneof=rings earrings bracelets necklaces"`
	StockQuantity int       `json:"stock_quantity"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
