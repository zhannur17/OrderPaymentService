package http

import "github.com/gin-gonic/gin"

func NewRouter(handler *PaymentHandler) *gin.Engine {
	r := gin.Default()

	r.POST("/payments", handler.Authorize)
	r.GET("/payments/:order_id", handler.GetByOrderID)

	return r
}
