package handlers

import (
	"net/http"

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

func (d *CartsHandlers) GetUserCart(c *gin.Context) { //Get cart and items from the cart
	logger, err := GetComponentLogger(c, "carts")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
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

func (d *CartsHandlers) AddToCart(c *gin.Context) {
	logger, err := GetComponentLogger(c, "carts")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	userID, exists := c.Get("userID")
	if !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	var addToCartRequest AddToCartRequest
	if err := c.ShouldBindBodyWithJSON(&addToCartRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
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

func (d *CartsHandlers) UpdateCartItem(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}
	itemID := c.Param("id")

	var quantityRequest QuantityRequest
	if err := c.ShouldBindBodyWithJSON(&quantityRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
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

func (d *CartsHandlers) RemoveFromCart(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
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

func (d *CartsHandlers) ClearCart(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
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
