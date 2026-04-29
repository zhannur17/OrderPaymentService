package domain

import "errors"

const (
	StatusAuthorized = "Authorized"
	StatusDeclined   = "Declined"

	MaxAllowedAmount int64 = 100000
)

var (
	ErrPaymentNotFound = errors.New("payment not found")
	ErrAmountExceeded  = errors.New("amount exceeds limit")
)

type Payment struct {
	ID            string
	OrderID       string
	TransactionID string
	Amount        int64
	Status        string
}

func NewPayment(id, orderID, transactionID string, amount int64) (*Payment, error) {
	status := StatusAuthorized
	if amount > MaxAllowedAmount {
		status = StatusDeclined
	}
	return &Payment{
		ID:            id,
		OrderID:       orderID,
		TransactionID: transactionID,
		Amount:        amount,
		Status:        status,
	}, nil
}
