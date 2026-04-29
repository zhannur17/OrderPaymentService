package usecase

import (
	"fmt"
	"log"

	"payment-service/internal/domain"

	"github.com/google/uuid"
)

type PaymentUseCase struct {
	repo      domain.PaymentRepository
	publisher domain.EventPublisher
}

func NewPaymentUseCase(repo domain.PaymentRepository, publisher domain.EventPublisher) *PaymentUseCase {
	return &PaymentUseCase{repo: repo, publisher: publisher}
}

type AuthorizeInput struct {
	OrderID       string
	Amount        int64
	CustomerEmail string
}

type AuthorizeOutput struct {
	Payment *domain.Payment
}

func (uc *PaymentUseCase) Authorize(input AuthorizeInput) (*AuthorizeOutput, error) {
	payment, err := domain.NewPayment(
		uuid.New().String(),
		input.OrderID,
		uuid.New().String(),
		input.Amount,
	)
	if err != nil {
		return nil, fmt.Errorf("create payment: %w", err)
	}

	if err := uc.repo.Save(payment); err != nil {
		return nil, fmt.Errorf("save payment: %w", err)
	}

	// Publish event AFTER successful DB save
	email := input.CustomerEmail
	if email == "" {
		email = "user@example.com"
	}

	event := domain.PaymentEvent{
		EventID:       uuid.New().String(),
		OrderID:       payment.OrderID,
		Amount:        payment.Amount,
		CustomerEmail: email,
		Status:        payment.Status,
	}

	if err := uc.publisher.PublishPaymentEvent(event); err != nil {
		log.Printf("[PaymentUseCase] Warning: failed to publish event: %v", err)
	}

	return &AuthorizeOutput{Payment: payment}, nil
}

func (uc *PaymentUseCase) GetByOrderID(orderID string) (*domain.Payment, error) {
	p, err := uc.repo.FindByOrderID(orderID)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (uc *PaymentUseCase) ListByStatus(status string) ([]*domain.Payment, error) {
	payments, err := uc.repo.ListByStatus(status)
	if err != nil {
		return nil, fmt.Errorf("list payments by status: %w", err)
	}
	return payments, nil
}
