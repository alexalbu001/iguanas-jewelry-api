package responses

type CartItemResponse struct {
	CartItemID  string  `json:"cart_item_id"`
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	Price       float64 `json:"price"`
	Quantity    int     `json:"quantity"`
	Subtotal    float64 `json:"subtotal"`
}

type CartResponse struct {
	CartID string             `json:"cart_id"`
	Items  []CartItemResponse `json:"items"`
	Total  float64            `json:"total"`
}
