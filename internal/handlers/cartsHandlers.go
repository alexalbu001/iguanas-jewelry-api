package handlers

import (
	"net/http"

	"github.com/alexalbu001/iguanas-jewelry/internal/responses"
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
	Quantity int `json:"quantity" binding:"required,min=0"`
}

func NewCartsHandler(cartsService *service.CartsService, productsService *service.ProductsService) *CartsHandlers {
	return &CartsHandlers{
		CartsService:    cartsService,
		ProductsService: productsService,
	}
}

func (d *CartsHandlers) GetUserCart(c *gin.Context) { //Get cart and items from the cart
	userID, exists := c.Get("userID")
	if !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	cartSummary, err := d.CartsService.GetUserCart(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}

	var cartResponse responses.CartResponse
	cartResponse.CartID = cartSummary.CartID
	cartResponse.Total = cartSummary.Total

	c.JSON(http.StatusOK, cartResponse)
}

func (d *CartsHandlers) AddToCart(c *gin.Context) {
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

	addToCartResult, err := d.CartsService.AddToCart(userID.(string), addToCartRequest.ProductID, addToCartRequest.Quantity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	response := convertToCartResponse(addToCartResult.CartSummary)

	if addToCartResult.Success {
		c.JSON(http.StatusCreated, response)
	} else {
		c.JSON(http.StatusBadRequest, response)
	}
}

func (d *CartsHandlers) UpdateCartItem(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}
	itemID := c.Param("item_id")

	var quantityRequest QuantityRequest
	if err := c.ShouldBindBodyWithJSON(&quantityRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	cartOperationResult, err := d.CartsService.UpdateCartItemQuantity(userID.(string), itemID, quantityRequest.Quantity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}

	response := convertToCartResponse(cartOperationResult.CartSummary)

	if cartOperationResult.Success {
		c.JSON(http.StatusAccepted, response)
	} else {
		c.JSON(http.StatusBadRequest, response)
	}
}

func (d *CartsHandlers) RemoveFromCart(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	itemID := c.Param("item_id")

	cartOperationResult, err := d.CartsService.RemoveFromCart(userID.(string), itemID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}

	response := convertToCartResponse(cartOperationResult.CartSummary)

	if cartOperationResult.Success {
		c.JSON(http.StatusAccepted, response)
	} else {
		c.JSON(http.StatusBadRequest, response)
	}
}

func (d *CartsHandlers) ClearCart(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	cartOperationResult, err := d.CartsService.ClearCart(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	response := convertToCartResponse(cartOperationResult.CartSummary)

	if cartOperationResult.Success {
		c.JSON(http.StatusAccepted, response)
	} else {
		c.JSON(http.StatusBadRequest, response)
	}
}
