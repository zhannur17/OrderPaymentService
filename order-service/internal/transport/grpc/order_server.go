package grpc

import (
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"order-service/internal/domain"

	orderv1 "github.com/zhannur17/ap2-generated/order/v1"
)

type OrderServer struct {
	orderv1.UnimplementedOrderServiceServer
	repo domain.OrderRepository
}

func NewOrderServer(repo domain.OrderRepository) *OrderServer {
	return &OrderServer{repo: repo}
}

func (s *OrderServer) SubscribeToOrderUpdates(
	req *orderv1.OrderRequest,
	stream orderv1.OrderService_SubscribeToOrderUpdatesServer,
) error {
	var lastStatus string

	for {
		select {
		case <-stream.Context().Done():
			return status.Error(codes.Canceled, "client disconnected")
		default:
		}

		order, err := s.repo.FindByID(req.OrderId)
		if err != nil {
			return status.Errorf(codes.NotFound, "order not found: %v", err)
		}

		if order.Status != lastStatus {
			lastStatus = order.Status
			if err := stream.Send(&orderv1.OrderStatusUpdate{
				OrderId: order.ID,
				Status:  order.Status,
			}); err != nil {
				return err
			}
		}

		if order.Status == domain.StatusPaid ||
			order.Status == domain.StatusFailed ||
			order.Status == domain.StatusCancelled {
			return nil
		}

		time.Sleep(500 * time.Millisecond)
	}
}
