package responses

type ErrorResponse struct {
	Message string                 `json:"message" example:"Product not found"`
	Code    string                 `json:"code" example:"PRODUCT_NOT_FOUND"`
	Details map[string]interface{} `json:"details,omitempty"`
}
