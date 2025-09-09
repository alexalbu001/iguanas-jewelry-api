package models

import "time"

type ProductImage struct {
	ID           string    `json:"id"`
	ProductID    string    `json:"product_id"`
	ImageURL     string    `json:"image_url"`
	IsMain       bool      `json:"is_main"`
	DisplayOrder int       `json:"display_order"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
