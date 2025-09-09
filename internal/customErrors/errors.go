package customerrors

import (
	"fmt"
	"net/http"
)

type HTTPError interface {
	Error() string
	StatusCode() int
	Code() string
	Details() map[string]interface{} // interface{} = return any type because any type implements interface{}
}
type APIError struct {
	Message    string `json:"message"`
	ErrorCode  string `json:"error_code"`
	HTTPStatus int    `json:"-"`
}

func (e *APIError) Error() string {
	return e.Message
}

func (e *APIError) StatusCode() int {
	return e.HTTPStatus
}

func (e *APIError) Code() string {
	return e.ErrorCode
}

func (e *APIError) Details() map[string]interface{} {
	return map[string]interface{}{}
}

var (
	ErrCartEmpty                 = APIError{"Cart is empty", "CART_EMPTY", http.StatusBadRequest}
	ErrInsufficientStockStatic   = APIError{"Not enough products in stock", "INSUFFICIENT_STOCK", http.StatusBadRequest}
	ErrOrderNotFound             = APIError{"Order not found", "ORDER_NOT_FOUND", http.StatusNotFound}
	ErrCartItemNotFound          = APIError{"Cart item not found", "CART_ITEM_NOT_FOUND", http.StatusNotFound}
	ErrUserNotFound              = APIError{"User not found", "USER_NOT_FOUND", http.StatusNotFound}
	ErrProductNotFound           = APIError{"Product not found", "Product_NOT_FOUND", http.StatusNotFound}
	ErrOrderNotOwned             = APIError{"Order does not belong to user", "ORDER_NOT_OWNED", http.StatusForbidden}
	ErrOrderAlreadyPaid          = APIError{"Order already paid", "ORDER_ALREADY_PAID", http.StatusForbidden}
	ErrOrderCancelled            = APIError{"Order is cancelled", "ORDER_CANCELLED", http.StatusForbidden}
	ErrCannotCancel              = APIError{"Order cannot be cancelled", "CANNOT_CANCEL", http.StatusBadRequest}
	ErrCannotChangeStatus        = APIError{"Order status cannot be changed", "CANNOT_CHANGE_STATUS", http.StatusBadRequest}
	ErrInvalidPrice              = APIError{"Price must be greater than 0", "INVALID_PRICE", http.StatusBadRequest}
	ErrMissingCategory           = APIError{"Product category is required", "MISSING_CATEGORY", http.StatusBadRequest}
	ErrMissingName               = APIError{"Product name is required", "MISSING_NAME", http.StatusBadRequest}
	ErrInvalidProductID          = APIError{"Invalid product ID format", "INVALID_PRODUCT_ID", http.StatusBadRequest}
	ErrEmptyProductID            = APIError{"Product ID cannot be empty", "EMPTY_PRODUCT_ID", http.StatusBadRequest}
	ErrPaymentNotFound           = APIError{"Payment not found", "ORDER_NOT_FOUND", http.StatusNotFound}
	ErrPaymentCardDeclined       = APIError{"Card was declined", "CARD_DECLINED", http.StatusBadRequest}
	ErrPaymentCardExpired        = APIError{"Card is expired", "CARD_EXPIRED", http.StatusBadRequest}
	ErrPaymentIncorrectCVC       = APIError{"Incorrect card CVC", "INCORRECT_CVC", http.StatusBadRequest}
	ErrPaymentProcessingFailed   = APIError{"Payment processing failed", "PAYMENT_PROCESSING_FAILED", http.StatusInternalServerError}
	ErrPaymentsTooManyRetries    = APIError{"Too many unsuccessful payments occured. Contact customer support", "PAYMENTS_TOO_MANY_RETRIES", http.StatusBadRequest}
	ErrInvalidShippingAddress    = APIError{"Invalid shipping address line", "INVALID_SHIPPING_ADDRESS", http.StatusBadRequest}
	ErrFieldsMissing             = APIError{"Required missing fields", "MISSING_FIELDS", http.StatusBadRequest}
	ErrShippingNameTooLong       = APIError{"Name is too long", "NAME_TOO_LONG", http.StatusBadRequest}
	ErrShippingAddressTooLong    = APIError{"Address is too long", "ADDRESS_TOO_LONG", http.StatusBadRequest}
	ErrInvalidEmail              = APIError{"Invalid email", "INVALID_EMAIL", http.StatusBadRequest}
	ErrInvalidUUID               = APIError{"Invalid UUID format", "INVALID_UUID", http.StatusBadRequest}
	ErrInvalidJSON               = APIError{"Invalid JSON payload", "INVALID_JSON", http.StatusBadRequest}
	ErrUserUnauthorized          = APIError{"User unauthorized", "USER_UNAUTHORIZED", http.StatusUnauthorized}
	ErrProductImageNotFound      = APIError{"Product image not found", "PRODUCT_IMAGE_NOT_FOUND", http.StatusNotFound}
	ErrMissingImageURL           = APIError{"Image URL is required", "MISSING_IMAGE_URL", http.StatusBadRequest}
	ErrProductImagesLimitReached = APIError{"Product images limit reached", "PRODUCT_IMAGES_LIMIT_REACHED", http.StatusBadRequest}
	ErrEmptyImageURL             = APIError{"Image URL cannot be empty", "EMPTY_IMAGE_URL", http.StatusBadRequest}
	ErrEmptyImageOrder           = APIError{"Image order cannot be empty", "EMPTY_IMAGE_ORDER", http.StatusBadRequest}
	// Generic validation
	ErrInvalidInput   = APIError{"Invalid input data", "INVALID_INPUT", http.StatusBadRequest}
	ErrInternalServer = APIError{"Internal Server Error", "INTERNAL_SERVER_ERROR", http.StatusInternalServerError}
)

type ErrInsufficientStock struct {
	ProductID      string
	RequestedQty   int
	AvailableStock int
	CurrentCartQty int
}

func (eis *ErrInsufficientStock) Error() string {
	return fmt.Sprintf("insufficient stock: requested %d, only %d available",
		eis.RequestedQty, eis.AvailableStock)
}

func (eis *ErrInsufficientStock) StatusCode() int {
	return http.StatusUnprocessableEntity
}

func (eis *ErrInsufficientStock) Code() string {
	return "INSUFFICENT_STOCK"
}

func (eis *ErrInsufficientStock) Details() map[string]interface{} {
	return map[string]interface{}{
		"product_id":            eis.ProductID,
		"requested_quantity":    eis.RequestedQty,
		"available_stock":       eis.AvailableStock,
		"current_cart_quantity": eis.CurrentCartQty,
	}
}

func NewErrInsufficientStock(productID string, requestedQty int, availableStock int, currentCartQty int) *ErrInsufficientStock {
	return &ErrInsufficientStock{
		ProductID:      productID,
		RequestedQty:   requestedQty,
		AvailableStock: availableStock,
		CurrentCartQty: currentCartQty,
	}
}
