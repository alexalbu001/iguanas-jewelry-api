package handlers

import (
	"net/http"

	customerrors "github.com/alexalbu001/iguanas-jewelry/internal/customErrors"
	"github.com/alexalbu001/iguanas-jewelry/internal/service"
	"github.com/gin-gonic/gin"
)

// type CartsHandlers struct {
//
//	    Carts *store.CartsStore
//	}
type CartsHandlers struct {
	CartsService    *service.CartsService
	ProductsService *service.ProductsService
}

type AddToCartRequest struct {
	ProductID string `json:"product_id" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required,min=1"`
}

type QuantityRequest struct {
	Quantity int `json:"quantity" binding:"min=0"`
}

func NewCartsHandler(cartsService *service.CartsService, productsService *service.ProductsService) *CartsHandlers {
	return &CartsHandlers{
		CartsService:    cartsService,
		ProductsService: productsService,
	}
}

// @Summary Get user cart
// @Description Retrieves the current user's cart with all items
// @Tags carts
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v1/cart [get]
func (d *CartsHandlers) GetUserCart(c *gin.Context) { //Get cart and items from the cart
	logger, err := GetComponentLogger(c, "carts")
	if err != nil {
		c.Error(err)
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.Error(&customerrors.ErrUserNotFound)
		return
	}
	logRequest(logger, "get user cart", "user_id", userID)
	cartSummary, err := d.CartsService.GetUserCart(c.Request.Context(), userID.(string))
	if err != nil {
		logError(logger, "failed to get user cart", err, "user_id", userID)
		c.Error(err)
		return
	}
	cartResponse := convertToCartResponse(cartSummary)

	c.JSON(http.StatusOK, cartResponse)
}

// @Summary Add item to cart
// @Description Adds a product to the user's cart with specified quantity
// @Tags carts
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body AddToCartRequest true "Add to cart request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} responses.ErrorResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v1/cart [post]
func (d *CartsHandlers) AddToCart(c *gin.Context) {
	logger, err := GetComponentLogger(c, "carts")
	if err != nil {
		c.Error(err)
		return
	}
	userID, exists := c.Get("userID")
	if !exists {
		c.Error(&customerrors.ErrUserNotFound)
		return
	}

	var addToCartRequest AddToCartRequest
	if err := c.ShouldBindBodyWithJSON(&addToCartRequest); err != nil {
		c.Error(&customerrors.ErrInvalidJSON)
		return
	}

	logRequest(logger, "add to cart", "user_id", userID)
	cartSummary, err := d.CartsService.AddToCart(c.Request.Context(), userID.(string), addToCartRequest.ProductID, addToCartRequest.Quantity)
	if err != nil {
		logError(logger, "failed to add to cart", err, "user_id", userID)
		c.Error(err)
		return
	}
	response := convertToCartResponse(cartSummary)

	c.JSON(http.StatusOK, response)
}

// @Summary Update cart item quantity
// @Description Updates the quantity of a specific item in the user's cart
// @Tags carts
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Cart item ID"
// @Param request body QuantityRequest true "Quantity update request"
// @Success 202 {object} map[string]interface{}
// @Failure 400 {object} responses.ErrorResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v1/cart/{id} [put]
func (d *CartsHandlers) UpdateCartItem(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.Error(&customerrors.ErrUserNotFound)
		return
	}
	itemID := c.Param("id")

	var quantityRequest QuantityRequest
	if err := c.ShouldBindBodyWithJSON(&quantityRequest); err != nil {
		c.Error(&customerrors.ErrInvalidJSON)
		return
	}

	cartSummary, err := d.CartsService.UpdateCartItemQuantity(c.Request.Context(), userID.(string), itemID, quantityRequest.Quantity)
	if err != nil {
		c.Error(err)
		return
	}

	response := convertToCartResponse(cartSummary)

	c.JSON(http.StatusAccepted, response)

}

// @Summary Remove item from cart
// @Description Removes a specific item from the user's cart
// @Tags carts
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Cart item ID"
// @Success 202 {object} map[string]interface{}
// @Failure 401 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v1/cart/{id} [delete]
func (d *CartsHandlers) RemoveFromCart(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.Error(&customerrors.ErrUserNotFound)
		return
	}

	itemID := c.Param("id")

	cartSummary, err := d.CartsService.RemoveFromCart(c.Request.Context(), userID.(string), itemID)
	if err != nil {
		c.Error(err) //log error
		return
	}

	response := convertToCartResponse(cartSummary)

	c.JSON(http.StatusAccepted, response)
}

// @Summary Clear cart
// @Description Removes all items from the user's cart
// @Tags carts
// @Produce json
// @Security ApiKeyAuth
// @Success 202 {object} map[string]interface{}
// @Failure 401 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v1/cart [delete]
func (d *CartsHandlers) ClearCart(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.Error(&customerrors.ErrUserNotFound)
		return
	}

	cartSummary, err := d.CartsService.ClearCart(c.Request.Context(), userID.(string))
	if err != nil {
		c.Error(err)
		return
	}
	response := convertToCartResponse(cartSummary)

	c.JSON(http.StatusAccepted, response)
}
