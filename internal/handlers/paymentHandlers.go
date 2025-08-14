package handlers

import (
	"net/http"

	"github.com/alexalbu001/iguanas-jewelry/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PaymentHandler struct {
	PaymentService *service.PaymentService
}

func NewPaymentHandler(paymentService *service.PaymentService) *PaymentHandler {
	return &PaymentHandler{
		PaymentService: paymentService,
	}
}

func (p *PaymentHandler) CreatePaymentIntent(c *gin.Context) {
	orderID := c.Param("order_id")
	idempotencyKey := uuid.NewString()

	clientSecret, err := p.PaymentService.CreatePaymentIntent(c.Request.Context(), orderID, idempotencyKey)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"client_secret": clientSecret,
	})

}
