package http

import (
	"errors"
	"fmt"
	"net/http"

	"order-service/internal/domain"
	"order-service/internal/usecase"

	"github.com/gin-gonic/gin"
)

// OrderUseCaseInterface allows handler to work with both OrderUseCase and CachedOrderUseCase
type OrderUseCaseInterface interface {
	CreateOrder(input usecase.CreateOrderInput) (*usecase.CreateOrderOutput, error)
	GetOrder(id string) (*domain.Order, error)
	CancelOrder(id string) (*domain.Order, error)
	GetOrdersByAmountRange(minAmount, maxAmount int64) ([]*domain.Order, error)
}

type OrderHandler struct {
	uc OrderUseCaseInterface
}

func NewOrderHandler(uc OrderUseCaseInterface) *OrderHandler {
	return &OrderHandler{uc: uc}
}

type createOrderRequest struct {
	CustomerID string `json:"customer_id" binding:"required"`
	ItemName   string `json:"item_name"   binding:"required"`
	Amount     int64  `json:"amount"      binding:"required"`
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req createOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	idempotencyKey := c.GetHeader("Idempotency-Key")

	out, err := h.uc.CreateOrder(usecase.CreateOrderInput{
		CustomerID:     req.CustomerID,
		ItemName:       req.ItemName,
		Amount:         req.Amount,
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		if errors.Is(err, domain.ErrInvalidAmount) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if isPaymentUnavailable(err) {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "payment service unavailable"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, orderResponse(out.Order))
}

func (h *OrderHandler) GetOrder(c *gin.Context) {
	id := c.Param("id")
	order, err := h.uc.GetOrder(id)
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, orderResponse(order))
}

func (h *OrderHandler) CancelOrder(c *gin.Context) {
	id := c.Param("id")
	order, err := h.uc.CancelOrder(id)
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
			return
		}
		if errors.Is(err, domain.ErrCannotCancel) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, orderResponse(order))
}

func (h *OrderHandler) ListOrdersByAmount(c *gin.Context) {
	minStr := c.Query("min_amount")
	maxStr := c.Query("max_amount")

	if minStr == "" || maxStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "min_amount and max_amount are required"})
		return
	}

	var minAmount, maxAmount int64
	if _, err := fmt.Sscan(minStr, &minAmount); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "min_amount must be a number"})
		return
	}
	if _, err := fmt.Sscan(maxStr, &maxAmount); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "max_amount must be a number"})
		return
	}
	if minAmount < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "min_amount must be >= 0"})
		return
	}
	if minAmount > maxAmount {
		c.JSON(http.StatusBadRequest, gin.H{"error": "min_amount must be <= max_amount"})
		return
	}

	orders, err := h.uc.GetOrdersByAmountRange(minAmount, maxAmount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result := make([]gin.H, 0, len(orders))
	for _, o := range orders {
		result = append(result, orderResponse(o))
	}
	c.JSON(http.StatusOK, result)
}

func orderResponse(o *domain.Order) gin.H {
	return gin.H{
		"id":          o.ID,
		"customer_id": o.CustomerID,
		"item_name":   o.ItemName,
		"amount":      o.Amount,
		"status":      o.Status,
		"created_at":  o.CreatedAt,
	}
}

func isPaymentUnavailable(err error) bool {
	if err == nil {
		return false
	}
	return containsString(err.Error(), "payment service")
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())
}
