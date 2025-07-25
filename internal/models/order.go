package models

import "time"

type Order struct {
	ID          string  `json:"id"`
	UserID      string  `json:"user_id"`
	TotalAmount float64 `json:"total_amount"`
	Status      string  `json:"status"`

	// Shipping Address (complete)
	ShippingName         string `json:"shipping_name"`
	ShippingEmail        string `json:"shipping_email"`         // For delivery notifications
	ShippingPhone        string `json:"shipping_phone"`         // For delivery contact
	ShippingAddressLine1 string `json:"shipping_address_line1"` // Street address
	ShippingAddressLine2 string `json:"shipping_address_line2"` // Apt/Suite/Building
	ShippingCity         string `json:"shipping_city"`
	ShippingState        string `json:"shipping_state"`
	ShippingPostalCode   string `json:"shipping_postal_code"`
	ShippingCountry      string `json:"shipping_country"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
