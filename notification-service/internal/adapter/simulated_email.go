package adapter

import (
	"fmt"
	"log"
	"math/rand"
	"time"
)

type SimulatedEmailSender struct{}

func NewSimulatedEmailSender() *SimulatedEmailSender {
	return &SimulatedEmailSender{}
}

func (s *SimulatedEmailSender) Send(to string, orderID string, amount int64, status string) error {
	time.Sleep(500 * time.Millisecond)

	if rand.Float32() < 0.3 {
		return fmt.Errorf("simulated: provider temporarily unavailable")
	}

	amountDollars := float64(amount) / 100.0
	log.Printf("[SimulatedEmail] Sent email to %s for Order #%s. Amount: $%.2f. Status: %s",
		to, orderID, amountDollars, status)

	return nil
}
