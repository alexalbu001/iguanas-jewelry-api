package handlers

import (
	"net/http"

	"github.com/alexalbu001/iguanas-jewelry/internal/responses"
	"github.com/alexalbu001/iguanas-jewelry/internal/service"
	"github.com/gin-gonic/gin"
)

type OrdersHandlers struct {
	ordersService *service.OrdersService
}

func NewOrdersHandlers(ordersService *service.OrdersService) *OrdersHandlers {
	return &OrdersHandlers{
		ordersService: ordersService,
	}
}

type AddShippingInfoToOrder struct {
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

func buildShippingInfoResponse(orderSummary service.OrderSummary) responses.ShippingInfoResponse {
	return responses.ShippingInfoResponse{
		Name:         orderSummary.ShippingName,
		Email:        orderSummary.ShippingAddress.Email,
		Phone:        orderSummary.ShippingAddress.Phone,
		AddressLine1: orderSummary.ShippingAddress.AddressLine1,
		AddressLine2: orderSummary.ShippingAddress.AddressLine2,
		City:         orderSummary.ShippingAddress.City,
		State:        orderSummary.ShippingAddress.State,
		PostalCode:   orderSummary.ShippingAddress.PostalCode,
		Country:      orderSummary.ShippingAddress.Country,
	}
}

func buildOrderItemResponse(orderSummary service.OrderSummary, orderItem service.OrderItemSummary) responses.OrderItemResponse {
	return responses.OrderItemResponse{
		ID:          orderItem.ID,
		OrderID:     orderSummary.ID,
		ProductID:   orderItem.ProductID,
		ProductName: orderItem.ProductName,
		Quantity:    orderItem.Quantity,
		Price:       orderItem.Price,
		Subtotal:    orderItem.Subtotal,
	}
}

func buildOrderResponse(orderSummary service.OrderSummary) responses.OrderResponse {
	var orderItemsResponse []responses.OrderItemResponse
	for _, orderItem := range orderSummary.OrderItems {
		orderItemResponse := buildOrderItemResponse(orderSummary, orderItem)
		orderItemsResponse = append(orderItemsResponse, orderItemResponse)
	}

	shippingInfoResponse := buildShippingInfoResponse(orderSummary)

	orderResponse := responses.OrderResponse{
		OrderID:      orderSummary.ID,
		Status:       orderSummary.Status,
		Total:        orderSummary.Total,
		CreatedDate:  orderSummary.CreatedDate.String(),
		Items:        orderItemsResponse,
		ShippingInfo: shippingInfoResponse,
	}
	return orderResponse
}

func (oh *OrdersHandlers) CreateOrder(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	var addShippingInfoToOrder AddShippingInfoToOrder

	if err := c.ShouldBindBodyWithJSON(&addShippingInfoToOrder); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	shippingInfo := service.ShippingInfo{
		Name:         addShippingInfoToOrder.Name,
		Email:        addShippingInfoToOrder.Email,
		Phone:        addShippingInfoToOrder.Phone,
		AddressLine1: addShippingInfoToOrder.AddressLine1,
		AddressLine2: addShippingInfoToOrder.AddressLine2,
		City:         addShippingInfoToOrder.City,
		State:        addShippingInfoToOrder.State,
		PostalCode:   addShippingInfoToOrder.PostalCode,
		Country:      addShippingInfoToOrder.Country,
	}

	orderSummary, err := oh.ordersService.CreateOrderFromCart(c.Request.Context(), userID.(string), shippingInfo)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, orderSummary)
}

func (oh *OrdersHandlers) ViewOrderHistory(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	orderSummaries, err := oh.ordersService.GetOrdersHistory(c.Request.Context(), userID.(string))
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, orderSummaries)
}

func (oh *OrdersHandlers) CancelOrder(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	orderID := c.Param("id")

	err := oh.ordersService.CancelOrder(c.Request.Context(), userID.(string), orderID)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message":  "Order cancelled successfully",
		"order_id": orderID,
	})
}

func (oh *OrdersHandlers) GetOrderInfo(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	orderID := c.Param("id")

	orderSummary, err := oh.ordersService.GetOrderInfo(c.Request.Context(), userID.(string), orderID)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, orderSummary)
}

func (oh *OrdersHandlers) GetAllOrders(c *gin.Context) {
	orderSummaries, err := oh.ordersService.GetAllOrders(c.Request.Context())
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, orderSummaries)
}

func (oh *OrdersHandlers) GetOrdersByStatus(c *gin.Context) {
	status := c.Query("status")
	if status != "" {
		orderSummaries, err := oh.ordersService.GetOrdersByStatus(c.Request.Context(), status)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, orderSummaries)
	} else {

		orderSummaries, err := oh.ordersService.GetAllOrders(c.Request.Context())
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, orderSummaries)
	}
}

func (oh *OrdersHandlers) UpdateOrderStatus(c *gin.Context) {

	orderID := c.Param("id")
	status := c.Query("status")

	err := oh.ordersService.UpdateOrderStatus(c.Request.Context(), status, orderID)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusAccepted, gin.H{
		"order_id": orderID,
		"message":  "Status modified",
	})
}
