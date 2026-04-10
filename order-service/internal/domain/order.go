package domain

import (
	"errors"
	"time"
)

const (
	StatusPending   = "Pending"
	StatusPaid      = "Paid"
	StatusFailed    = "Failed"
	StatusCancelled = "Cancelled"
)

type Order struct {
	ID             string
	CustomerID     string
	ItemName       string
	Amount         int64 // in cents
	Status         string
	CreatedAt      time.Time
	IdempotencyKey string
}

var (
	ErrOrderNotFound  = errors.New("order not found")
	ErrInvalidAmount  = errors.New("amount must be greater than 0")
	ErrCannotCancel   = errors.New("only pending orders can be cancelled")
	ErrDuplicateOrder = errors.New("duplicate order")
)

func NewOrder(id, customerID, itemName string, amount int64) (*Order, error) {
	if amount <= 0 {
		return nil, ErrInvalidAmount
	}
	return &Order{
		ID:             id,
		CustomerID:     customerID,
		ItemName:       itemName,
		Amount:         amount,
		Status:         StatusPending,
		CreatedAt:      time.Now(),
		IdempotencyKey: "",
	}, nil
}

func (o *Order) Cancel() error {
	if o.Status != StatusPending {
		return ErrCannotCancel
	}
	o.Status = StatusCancelled
	return nil
}

func (o *Order) MarkPaid() {
	o.Status = StatusPaid
}

func (o *Order) MarkFailed() {
	o.Status = StatusFailed
}
