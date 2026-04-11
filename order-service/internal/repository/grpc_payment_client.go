package repository

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"order-service/internal/domain"

	paymentv1 "github.com/zhannur17/ap2-generated/payment/v1"
)

type GRPCPaymentClient struct {
	client paymentv1.PaymentServiceClient
}

func NewGRPCPaymentClient(addr string) (*GRPCPaymentClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("grpc dial: %w", err)
	}
	return &GRPCPaymentClient{client: paymentv1.NewPaymentServiceClient(conn)}, nil
}

func (c *GRPCPaymentClient) Authorize(req domain.PaymentRequest) (*domain.PaymentResponse, error) {
	resp, err := c.client.ProcessPayment(context.Background(), &paymentv1.PaymentRequest{
		OrderId: req.OrderID,
		Amount:  req.Amount,
	})
	if err != nil {
		return nil, fmt.Errorf("grpc payment: %w", err)
	}
	return &domain.PaymentResponse{
		TransactionID: resp.TransactionId,
		Status:        resp.Status,
	}, nil
}
