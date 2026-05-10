package domain

type EmailSender interface {
	Send(to string, orderID string, amount int64, status string) error
}
