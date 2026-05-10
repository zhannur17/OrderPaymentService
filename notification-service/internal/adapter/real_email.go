package adapter

import (
	"fmt"
	"log"
)

type RealEmailSender struct {
	smtpHost string
	smtpPort string
	from     string
}

func NewRealEmailSender(smtpHost, smtpPort, from string) *RealEmailSender {
	return &RealEmailSender{
		smtpHost: smtpHost,
		smtpPort: smtpPort,
		from:     from,
	}
}

func (r *RealEmailSender) Send(to string, orderID string, amount int64, status string) error {
	amountDollars := float64(amount) / 100.0
	log.Printf("[RealEmail] Sending email from %s to %s for Order #%s. Amount: $%.2f. Status: %s",
		r.from, to, orderID, amountDollars, status)

	if r.smtpHost == "" {
		return fmt.Errorf("SMTP host not configured")
	}

	log.Printf("[RealEmail] Email sent successfully to %s", to)
	return nil
}
