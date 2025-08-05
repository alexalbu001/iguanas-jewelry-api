package customerrors

import "net/http"

type APIError struct {
	Message    string `json:"message"`
	Code       string `json:"code"`
	StatusCode int    `json:"-"`
}

func (e *APIError) Error() string {
	return e.Message
}

var (
	ErrCartEmpty          = APIError{"Cart is empty", "CART_EMPTY", http.StatusBadRequest}
	ErrInsufficientStock  = APIError{"Not enough products in stock", "INSUFFICIENT_STOCK", http.StatusBadRequest}
	ErrOrderNotFound      = APIError{"Order not found", "ORDER_NOT_FOUND", http.StatusNotFound}
	ErrCartItemNotFound   = APIError{"Cart item not found", "CART_ITEM_NOT_FOUND", http.StatusNotFound}
	ErrUserNotFound       = APIError{"User not found", "USER_NOT_FOUND", http.StatusNotFound}
	ErrProductNotFound    = APIError{"Product not found", "Product_NOT_FOUND", http.StatusNotFound}
	ErrOrderNotOwned      = APIError{"Order does not belong to user", "ORDER_NOT_OWNED", http.StatusForbidden}
	ErrCannotCancel       = APIError{"Order cannot be cancelled", "CANNOT_CANCEL", http.StatusBadRequest}
	ErrCannotChangeStatus = APIError{"Order status cannot be changed", "CANNOT_CHANGE_STATUS", http.StatusBadRequest}
	ErrInvalidPrice       = APIError{"Price must be greater than 0", "INVALID_PRICE", http.StatusBadRequest}
	ErrMissingCategory    = APIError{"Product category is required", "MISSING_CATEGORY", http.StatusBadRequest}
	ErrMissingName        = APIError{"Product name is required", "MISSING_NAME", http.StatusBadRequest}
	ErrInvalidProductID   = APIError{"Invalid product ID format", "INVALID_PRODUCT_ID", http.StatusBadRequest}
	ErrEmptyProductID     = APIError{"Product ID cannot be empty", "EMPTY_PRODUCT_ID", http.StatusBadRequest}

	// Generic validation
	ErrInvalidInput = APIError{"Invalid input data", "INVALID_INPUT", http.StatusBadRequest}
)
