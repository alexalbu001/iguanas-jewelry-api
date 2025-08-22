package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/alexalbu001/iguanas-jewelry/internal/models"
	"github.com/alexalbu001/iguanas-jewelry/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/webhook"
)

type PaymentHandler struct {
	paymentService      *service.PaymentService
	ordersService       *service.OrdersService
	stripeWebhookSecret string
}

func NewPaymentHandler(paymentService *service.PaymentService, ordersService *service.OrdersService, stripeWebhookSecret string) *PaymentHandler {
	return &PaymentHandler{
		paymentService:      paymentService,
		ordersService:       ordersService,
		stripeWebhookSecret: stripeWebhookSecret,
	}
}

func (p *PaymentHandler) RetryOrderPayment(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}
	orderID := c.Param("order_id")

	idempotencyKey := uuid.NewString()

	clientSecret, err := p.paymentService.RetryOrderPayment(c.Request.Context(), userID.(string), orderID, idempotencyKey)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"client_secret": clientSecret,
	})
}

// This will get called by STRIPE
func (p *PaymentHandler) HandleWebhook(c *gin.Context) {
	logger, err := GetComponentLogger(c, "payment")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	webhookSecret := p.stripeWebhookSecret
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

		logRequest(logger, "create payment", "order_id", orderID)
		_, err = p.paymentService.CreatePayment(c.Request.Context(), payment)
		if err != nil {
			logError(logger, "failed to create payment", err, "order_id", orderSummary.ID)
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

		logRequest(logger, "create failed payment", "order_id", orderID)
		_, err = p.paymentService.CreatePayment(c.Request.Context(), payment)
		if err != nil {
			logError(logger, "create failed payment", err, "order_id", orderID)
			c.Error(err)
			return
		}
	default:
		fmt.Printf("Unhandled event type: %s\n", event.Type)
	}

	c.Status(http.StatusOK)
}
