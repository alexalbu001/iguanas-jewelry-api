package models

import "time"

type ProductImage struct {
	ID           string    `json:"id"`
	ProductID    string    `json:"product_id"`
	ImageKey     string    `json:"-"`         // hidden from JSON output
	ImageURL     string    `json:"image_url"` // Generated dynamically, NOT stored in DB
	ContentType  string    `json:"content_type"`
	IsMain       bool      `json:"is_main"`
	DisplayOrder int       `json:"display_order"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
