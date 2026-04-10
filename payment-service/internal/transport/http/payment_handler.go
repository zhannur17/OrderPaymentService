package http

import (
	"errors"
	"net/http"

	"payment-service/internal/domain"
	"payment-service/internal/usecase"

	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	uc *usecase.PaymentUseCase
}

func NewPaymentHandler(uc *usecase.PaymentUseCase) *PaymentHandler {
	return &PaymentHandler{uc: uc}
}

type authorizeRequest struct {
	OrderID string `json:"order_id" binding:"required"`
	Amount  int64  `json:"amount"   binding:"required"`
}

func (h *PaymentHandler) Authorize(c *gin.Context) {
	var req authorizeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	out, err := h.uc.Authorize(usecase.AuthorizeInput{
		OrderID: req.OrderID,
		Amount:  req.Amount,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	statusCode := http.StatusCreated
	if out.Payment.Status == domain.StatusDeclined {
		statusCode = http.StatusUnprocessableEntity
	}

	c.JSON(statusCode, gin.H{
		"id":             out.Payment.ID,
		"order_id":       out.Payment.OrderID,
		"transaction_id": out.Payment.TransactionID,
		"amount":         out.Payment.Amount,
		"status":         out.Payment.Status,
	})
}

func (h *PaymentHandler) GetByOrderID(c *gin.Context) {
	orderID := c.Param("order_id")
	payment, err := h.uc.GetByOrderID(orderID)
	if err != nil {
		if errors.Is(err, domain.ErrPaymentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":             payment.ID,
		"order_id":       payment.OrderID,
		"transaction_id": payment.TransactionID,
		"amount":         payment.Amount,
		"status":         payment.Status,
	})
}
