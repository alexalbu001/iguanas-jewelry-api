package responses

// Carts
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

//Orders

type CreateOrderResponse struct {
	OrderID     string  `json:"order_id"`
	Status      string  `json:"status"`
	Total       float64 `json:"total"`
	CreatedDate string  `json:"created_date"`
	Message     string  `json:"message"` // "Order created successfully"
}

type ShippingInfoResponse struct {
	Name         string `json:"name" binding:"required"`
	Email        string `json:"email" binding:"required,email"`
	Phone        string `json:"phone" binding:"required"`
	AddressLine1 string `json:"address_line1" binding:"required"`
	AddressLine2 string `json:"address_line2"` // Optional
	City         string `json:"city" binding:"required"`
	State        string `json:"state" binding:"required"`
	PostalCode   string `json:"postal_code" binding:"required"`
	Country      string `json:"country" binding:"required"`
}

type OrderItemResponse struct {
	ID          string  `json:"id"`
	OrderID     string  `json:"order_id"`
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"` // Price at time of purchase (snapshot!)
	Subtotal    float64 `json:"subtotal"`
}
type OrderResponse struct {
	OrderID      string              `json:"order_id"`
	Status       string              `json:"status"`
	Total        float64             `json:"total"`
	CreatedDate  string              `json:"created_date"`
	Items        []OrderItemResponse `json:"items"`
	ShippingInfo ShippingInfoResponse
}
type OrdersListResponse struct {
	Orders []OrderResponse `json:"orders"`
}

type ChangeOrderStatusResponse struct {
	OrderID string `json:"order_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
	Success bool   `json:"success"`
}

//Users

type UserProfileResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	CreatedAt string `json:"created_at"`
}
