package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	customerrors "github.com/alexalbu001/iguanas-jewelry/internal/customErrors"
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

// @Summary Retry order payment
// @Description Retries payment for a failed order by creating a new payment intent
// @Tags payments
// @Produce json
// @Security ApiKeyAuth
// @Param order_id path string true "Order ID"
// @Success 200 {object} map[string]interface{} "Payment intent created"
// @Failure 400 {object} responses.ErrorResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v1/payments/{order_id}/retry [post]
func (p *PaymentHandler) RetryOrderPayment(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.Error(&customerrors.ErrUserNotFound)
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

// @Summary Handle Stripe webhook
// @Description Processes Stripe webhook events for payment status updates
// @Tags payments
// @Accept json
// @Produce json
// @Param stripe-signature header string true "Stripe webhook signature"
// @Success 200 {string} string "Webhook processed successfully"
// @Failure 400 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v1/payments/webhook [post]
// This will get called by STRIPE
func (p *PaymentHandler) HandleWebhook(c *gin.Context) {
	logger, err := GetComponentLogger(c, "payment")
	if err != nil {
		c.Error(err)
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
