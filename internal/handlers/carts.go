package handlers

import (
	"log"
	"net/http"

	"github.com/alexalbu001/iguanas-jewelry/internal/models"
	"github.com/alexalbu001/iguanas-jewelry/internal/responses"
	"github.com/gin-gonic/gin"
)

type CartsStore interface {
	GetOrCreateCartByUserID(id string) (models.Cart, error)
	EmptyCart(userID string) error
	GetCartItemByID(id string) (models.CartItems, error)
	AddItemToCart(cartID, productID string, quantity int) (models.CartItems, error)
	GetCartItems(cartID string) ([]models.CartItems, error)
	UpdateCartItemQuantity(itemID string, newQuantity int) error
	DeleteCartItem(cartItemID string) error
}

// type CartsHandlers struct {
//
//	    Carts *store.CartsStore
//	}
type CartsHandlers struct {
	Carts    CartsStore
	Products ProductsStore
}

type AddToCartRequest struct {
	ProductID string `json:"product_id" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required,min=1"`
}

type QuantityRequest struct {
	Quantity int `json:"quantity" binding:"required,min=0"`
}

func (d *CartsHandlers) GetUserCart(c *gin.Context) { //Get cart and items from the cart
	userID, exists := c.Get("userID")
	if !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	cart, err := d.Carts.GetOrCreateCartByUserID(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}

	cartItems, err := d.Carts.GetCartItems(cart.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}

	var cartResponse responses.CartResponse
	cartResponse.CartID = cart.ID
	cartResponse.Items = []responses.CartItemResponse{}

	for _, cartItem := range cartItems { // loop won't run if cart is empty because cartItems would be an empty slice. The loop just wouldn't execute - no error.
		cartProducts, err := d.Products.GetByID(cartItem.ProductID)
		if err != nil {
			deleteErr := d.Carts.DeleteCartItem(cartItem.ID)
			if deleteErr != nil {
				// Log it but don't fail the whole request
				log.Printf("Failed to delete invalid cart item %s: %v", cartItem.ID, deleteErr)
			}
			continue
		}
		cartItemResponse := responses.CartItemResponse{
			CartItemID:  cartItem.ID,
			ProductID:   cartProducts.ID,
			ProductName: cartProducts.Name,
			Price:       cartProducts.Price,
			Quantity:    cartItem.Quantity,
			Subtotal:    cartProducts.Price * float64(cartItem.Quantity),
		}

		cartResponse.Items = append(cartResponse.Items, cartItemResponse)
	}
	for _, v := range cartResponse.Items {
		cartResponse.Total += v.Subtotal
	}

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

	cart, err := d.Carts.GetOrCreateCartByUserID(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}

	item, err := d.Carts.AddItemToCart(cart.ID, addToCartRequest.ProductID, addToCartRequest.Quantity)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, item)
}

func (d *CartsHandlers) UpdateCartItem(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}
	itemID := c.Param("item_id")

	cartItem, err := d.Carts.GetCartItemByID(itemID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}

	cart, err := d.Carts.GetOrCreateCartByUserID(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}

	if cart.ID != cartItem.CartID {
		c.JSON(http.StatusNotFound, gin.H{"error": "cart item not found"})
		return
	}

	var quantityRequest QuantityRequest
	if err := c.ShouldBindBodyWithJSON(&quantityRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	err = d.Carts.UpdateCartItemQuantity(itemID, quantityRequest.Quantity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}

	if quantityRequest.Quantity == 0 {
		err = d.Carts.DeleteCartItem(itemID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
	}
	c.JSON(http.StatusAccepted, cartItem.Quantity)
}

func (d *CartsHandlers) RemoveFromCart(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	itemID := c.Param("item_id")

	cartItem, err := d.Carts.GetCartItemByID(itemID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}

	cart, err := d.Carts.GetOrCreateCartByUserID(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}

	if cart.ID != cartItem.CartID {
		c.JSON(http.StatusNotFound, gin.H{"error": "cart item not found"})
		return
	}

	err = d.Carts.DeleteCartItem(itemID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully", "id": itemID})
}

func (d *CartsHandlers) ClearCart(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	err := d.Carts.EmptyCart(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Cart cleared successfully"})
}

func NewCartsHandler(cartsStore CartsStore, productsStore ProductsStore) *CartsHandlers {
	return &CartsHandlers{
		Carts:    cartsStore,
		Products: productsStore,
	}
}
