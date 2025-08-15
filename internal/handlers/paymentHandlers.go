package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/alexalbu001/iguanas-jewelry/internal/models"
	"github.com/alexalbu001/iguanas-jewelry/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/webhook"
)

type PaymentHandler struct {
	paymentService *service.PaymentService
	ordersService  *service.OrdersService
}

func NewPaymentHandler(paymentService *service.PaymentService, ordersService *service.OrdersService) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
		ordersService:  ordersService,
	}
}

func (p *PaymentHandler) RetryOrderPayment(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}
	orderID := c.Param("order_id")

	orderSummary, err := p.ordersService.GetOrderByIDAdmin(c.Request.Context(), orderID)
	if err != nil {
		c.Error(err)
		return
	}

	if userID != orderSummary.UserID {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
	}

	if orderSummary.Status == "paid" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order already paid"})
		return
	}
	if orderSummary.Status == "cancelled" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Cannot retry cancelled order",
		})
		return
	}

	payments, err := p.paymentService.GetPaymentsByOrderID(c.Request.Context(), orderID)
	if err != nil {
		c.Error(err)
		return
	}

	if len(payments) > 3 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Too many unsuccessful payments occured. Contact customer support"})
		return
	}

	idempotencyKey := uuid.NewString()

	clientSecret, err := p.paymentService.CreatePaymentIntent(c.Request.Context(), orderID, idempotencyKey)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"client_secret": clientSecret,
	})
}

func (p *PaymentHandler) HandleWebhook(c *gin.Context) {
	webhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.Error(fmt.Errorf("error reading request body: %w", err))
		return
	}

	signature := c.GetHeader("Stripe-Signature")
	event, err := webhook.ConstructEvent(payload, signature, webhookSecret)
	if err != nil {
		c.Error(fmt.Errorf("webhook signature verification failed: %w", err))
		return
	}

	switch event.Type {
	case "payment_intent.succeeded":
		var paymentIntent stripe.PaymentIntent
		err := json.Unmarshal(event.Data.Raw, &paymentIntent)
		if err != nil {
			c.Error(err)
			return
		}

		orderID := paymentIntent.Metadata["order_id"]

		if orderID == "" {
			c.Error(fmt.Errorf("missing order_id in payment metadata"))
			return
		}

		orderSummary, err := p.ordersService.GetOrderByIDAdmin(c.Request.Context(), orderID)
		if err != nil {
			c.Error(err)
			return
		}

		payment := models.Payment{
			OrderID:         orderID,
			UserID:          orderSummary.UserID,
			StripePaymentID: &paymentIntent.ID,
			Amount:          float64(paymentIntent.Amount) / 100, // they are in cents
			Status:          "succeeded",
		}

		_, err = p.paymentService.CreatePayment(c.Request.Context(), payment)
		if err != nil {
			c.Error(err)
			return
		}

		err = p.ordersService.UpdateOrderStatus(c.Request.Context(), "paid", orderID)
		if err != nil {
			c.Error(err)
			return
		}

		fmt.Printf("Payment successful for order: %s !", orderID)
	case "payment_intent.payment_failed":
		var paymentIntent stripe.PaymentIntent
		err := json.Unmarshal(event.Data.Raw, &paymentIntent)
		if err != nil {
			c.Error(err)
			return
		}
		var failureMessage *string
		if paymentIntent.LastPaymentError != nil {
			failureMessage = &paymentIntent.LastPaymentError.Msg
		}

		orderID := paymentIntent.Metadata["order_id"]

		if orderID == "" {
			c.Error(fmt.Errorf("missing order_id in payment metadata"))
			return
		}

		orderSummary, err := p.ordersService.GetOrderByIDAdmin(c.Request.Context(), orderID)
		if err != nil {
			c.Error(err)
			return
		}

		payment := models.Payment{
			OrderID:         orderID,
			UserID:          orderSummary.UserID,
			StripePaymentID: &paymentIntent.ID,
			Amount:          float64(paymentIntent.Amount) / 100,
			Status:          "failed",
			FailureReason:   failureMessage,
		}

		_, err = p.paymentService.CreatePayment(c.Request.Context(), payment)
		if err != nil {
			c.Error(err)
			return
		}
	default:
		fmt.Printf("Unhandled event type: %s\n", event.Type)
	}

	c.Status(http.StatusOK)
}
