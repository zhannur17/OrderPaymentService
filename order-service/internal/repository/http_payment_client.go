package repository

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"order-service/internal/domain"
)

type paymentRequestBody struct {
	OrderID string `json:"order_id"`
	Amount  int64  `json:"amount"`
}

type paymentResponseBody struct {
	TransactionID string `json:"transaction_id"`
	Status        string `json:"status"`
	Message       string `json:"message,omitempty"`
}

type HTTPPaymentClient struct {
	client  *http.Client
	baseURL string
}

func NewHTTPPaymentClient(client *http.Client, baseURL string) *HTTPPaymentClient {
	return &HTTPPaymentClient{client: client, baseURL: baseURL}
}

func (c *HTTPPaymentClient) Authorize(req domain.PaymentRequest) (*domain.PaymentResponse, error) {
	body, err := json.Marshal(paymentRequestBody{
		OrderID: req.OrderID,
		Amount:  req.Amount,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	resp, err := c.client.Post(c.baseURL+"/payments", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("payment service call failed: %w", err)
	}
	defer resp.Body.Close()

	var result paymentResponseBody
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &domain.PaymentResponse{
		TransactionID: result.TransactionID,
		Status:        result.Status,
	}, nil
}
