package domain

type PaymentRepository interface {
	Save(payment *Payment) error
	FindByOrderID(orderID string) (*Payment, error)
	ListByStatus(status string) ([]*Payment, error)
}
