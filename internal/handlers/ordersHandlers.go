package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	customerrors "github.com/alexalbu001/iguanas-jewelry/internal/customErrors"
	"github.com/alexalbu001/iguanas-jewelry/internal/responses"
	"github.com/alexalbu001/iguanas-jewelry/internal/service"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type OrdersHandlers struct {
	ordersService  *service.OrdersService
	paymentService *service.PaymentService
	sqsClient      *sqs.Client
	queueURL       string
}

type ExpirationMessage struct {
	OrderID   string    `json:"order_id"`
	CreatedAt time.Time `json:"created_at`
}

func NewOrdersHandlers(ordersService *service.OrdersService, paymentService *service.PaymentService, sqsClient *sqs.Client, queueURL string) *OrdersHandlers {
	return &OrdersHandlers{
		ordersService:  ordersService,
		paymentService: paymentService,
		sqsClient:      sqsClient,
		queueURL:       queueURL,
	}
}

type AddShippingInfoToOrder struct {
	Name         string `json:"name" binding:"required,max=200"`
	Email        string `json:"email" binding:"required,email"`
	Phone        string `json:"phone" binding:"required,e164"`
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

// @Summary Create order from cart
// @Description Creates a new order from the user's cart with shipping information
// @Tags orders
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param shippingInfo body AddShippingInfoToOrder true "Shipping information for the order"
// @Success 200 {object} map[string]interface{} "Order created with payment intent"
// @Failure 400 {object} responses.ErrorResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v1/orders [post]
func (oh *OrdersHandlers) CreateOrder(c *gin.Context) {
	logger, err := GetComponentLogger(c, "orders")
	if err != nil {
		c.Error(err)
		return
	}
	userID, exists := c.Get("userID")
	if !exists {
		c.Error(&customerrors.ErrUserNotFound)
		return
	}

	var addShippingInfoToOrder AddShippingInfoToOrder

	if err := c.ShouldBindBodyWithJSON(&addShippingInfoToOrder); err != nil {
		c.Error(&customerrors.ErrInvalidJSON)
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

	logRequest(logger, "create order", "user_id", userID)
	orderSummary, err := oh.ordersService.CreateOrderFromCart(c.Request.Context(), userID.(string), shippingInfo)
	if err != nil {
		logError(logger, "failed to create order", err, "user_id", userID, "order_id", orderSummary.ID)
		c.Error(err)
		return
	}

	idempotencyKey := uuid.NewString()

	logRequest(logger, "create payment intent", "user_id", userID)
	clientSecret, err := oh.paymentService.CreatePaymentIntent(c.Request.Context(), orderSummary.ID, idempotencyKey)
	if err != nil {
		logError(logger, "failed to create payment intent", err, "user_id", userID, "order_id", orderSummary.ID)
		c.Error(err)
		return
	}

	input, err := oh.CreateSQSInputMessage(orderSummary.ID)
	if err != nil {
		c.Error(err)
		return
	}
	_, err = oh.sqsClient.SendMessage(c.Request.Context(), input)
	if err != nil {
		// Log but don't fail the order (best effort approach)
		logError(logger, "failed to send message to sqs", err, "order_id", orderSummary.ID)
	}

	c.JSON(http.StatusOK, gin.H{
		"order":         orderSummary,
		"client_secret": clientSecret,
	})
}

func (oh *OrdersHandlers) CreateSQSInputMessage(orderID string) (*sqs.SendMessageInput, error) {
	expirationMsg := ExpirationMessage{
		OrderID:   orderID,
		CreatedAt: time.Now(),
	}

	messageBody, err := json.Marshal(expirationMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}
	return &sqs.SendMessageInput{
		QueueUrl:     aws.String(oh.queueURL),
		MessageBody:  aws.String(string(messageBody)),
		DelaySeconds: 900, // 15 minutes
	}, nil
}

// @Summary View order history
// @Description Retrieves the order history for the authenticated user
// @Tags orders
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {array} service.OrderSummary
// @Failure 401 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v1/orders/history [get]
func (oh *OrdersHandlers) ViewOrderHistory(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.Error(&customerrors.ErrUserNotFound)
		return
	}

	orderSummaries, err := oh.ordersService.GetOrdersHistory(c.Request.Context(), userID.(string))
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, orderSummaries)
}

// @Summary Cancel order
// @Description Cancels a specific order for the authenticated user
// @Tags orders
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Order ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} responses.ErrorResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 403 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v1/orders/{id}/cancel [post]
func (oh *OrdersHandlers) CancelOrder(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.Error(&customerrors.ErrUserNotFound)
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

// @Summary Get order information
// @Description Retrieves detailed information about a specific order
// @Tags orders
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Order ID"
// @Success 200 {object} service.OrderSummary
// @Failure 400 {object} responses.ErrorResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 403 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v1/orders/{id} [get]
func (oh *OrdersHandlers) GetOrderInfo(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.Error(&customerrors.ErrUserNotFound)
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

// @Summary Get all orders (Admin only)
// @Description Retrieves all orders in the system (admin access required)
// @Tags orders
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {array} service.OrderSummary
// @Failure 401 {object} responses.ErrorResponse
// @Failure 403 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v1/orders [get]
func (oh *OrdersHandlers) GetAllOrders(c *gin.Context) {
	orderSummaries, err := oh.ordersService.GetAllOrders(c.Request.Context())
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, orderSummaries)
}

// @Summary Get orders by status
// @Description Retrieves orders filtered by status (admin access required)
// @Tags orders
// @Produce json
// @Security ApiKeyAuth
// @Param status query string false "Order status filter"
// @Success 200 {array} service.OrderSummary
// @Failure 401 {object} responses.ErrorResponse
// @Failure 403 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v1/orders/status [get]
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

// @Summary Update order status (Admin only)
// @Description Updates the status of a specific order (admin access required)
// @Tags orders
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Order ID"
// @Param status query string true "New order status"
// @Success 202 {object} map[string]interface{}
// @Failure 400 {object} responses.ErrorResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 403 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v1/orders/{id}/status [put]
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
