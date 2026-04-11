package grpc

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"payment-service/internal/usecase"

	paymentv1 "github.com/zhannur17/ap2-generated/payment/v1"
)

type PaymentServer struct {
	paymentv1.UnimplementedPaymentServiceServer
	uc *usecase.PaymentUseCase
}

func NewPaymentServer(uc *usecase.PaymentUseCase) *PaymentServer {
	return &PaymentServer{uc: uc}
}

func (s *PaymentServer) ProcessPayment(ctx context.Context, req *paymentv1.PaymentRequest) (*paymentv1.PaymentResponse, error) {
	if req.Amount <= 0 {
		return nil, status.Error(codes.InvalidArgument, "amount must be greater than 0")
	}

	out, err := s.uc.Authorize(usecase.AuthorizeInput{
		OrderID: req.OrderId,
		Amount:  req.Amount,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "payment failed: %v", err)
	}

	return &paymentv1.PaymentResponse{
		TransactionId: out.Payment.TransactionID,
		Status:        out.Payment.Status,
		ProcessedAt:   timestamppb.New(time.Now()),
	}, nil
}
