package usecase

import (
	"fmt"

	"order-service/internal/domain"

	"github.com/google/uuid"
)

type orderRepo interface {
	domain.OrderRepository
	SaveWithIdempotencyKey(order *domain.Order, key string) error
}

type OrderUseCase struct {
	repo          orderRepo
	paymentClient domain.PaymentClient
}

func NewOrderUseCase(repo orderRepo, paymentClient domain.PaymentClient) *OrderUseCase {
	return &OrderUseCase{repo: repo, paymentClient: paymentClient}
}

type CreateOrderInput struct {
	CustomerID     string
	ItemName       string
	Amount         int64
	IdempotencyKey string
}

type CreateOrderOutput struct {
	Order *domain.Order
}

func (uc *OrderUseCase) CreateOrder(input CreateOrderInput) (*CreateOrderOutput, error) {
	// Idempotency: return existing order if key already used
	if input.IdempotencyKey != "" {
		existing, err := uc.repo.FindByIdempotencyKey(input.IdempotencyKey)
		if err == nil && existing != nil {
			return &CreateOrderOutput{Order: existing}, nil
		}
	}

	order, err := domain.NewOrder(uuid.New().String(), input.CustomerID, input.ItemName, input.Amount)
	if err != nil {
		return nil, fmt.Errorf("invalid order: %w", err)
	}

	if input.IdempotencyKey != "" {
		if err := uc.repo.SaveWithIdempotencyKey(order, input.IdempotencyKey); err != nil {
			return nil, fmt.Errorf("save order: %w", err)
		}
	} else {
		if err := uc.repo.Save(order); err != nil {
			return nil, fmt.Errorf("save order: %w", err)
		}
	}

	payResp, err := uc.paymentClient.Authorize(domain.PaymentRequest{
		OrderID: order.ID,
		Amount:  order.Amount,
	})
	if err != nil {
		order.MarkFailed()
		_ = uc.repo.Update(order)
		return nil, fmt.Errorf("payment service unavailable: %w", err)
	}

	if payResp.Status == "Authorized" {
		order.MarkPaid()
	} else {
		order.MarkFailed()
	}

	if err := uc.repo.Update(order); err != nil {
		return nil, fmt.Errorf("update order: %w", err)
	}

	return &CreateOrderOutput{Order: order}, nil
}

func (uc *OrderUseCase) GetOrder(id string) (*domain.Order, error) {
	return uc.repo.FindByID(id)
}

func (uc *OrderUseCase) CancelOrder(id string) (*domain.Order, error) {
	order, err := uc.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if err := order.Cancel(); err != nil {
		return nil, err
	}
	if err := uc.repo.Update(order); err != nil {
		return nil, fmt.Errorf("update order: %w", err)
	}
	return order, nil
}
