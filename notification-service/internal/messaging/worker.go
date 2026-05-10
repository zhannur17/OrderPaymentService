package messaging

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"

	"notification-service/internal/domain"
)

const (
	maxRetries = 3
	baseDelay  = 2 * time.Second
)

type Worker struct {
	sender      domain.EmailSender
	redisClient *redis.Client
}

func NewWorker(sender domain.EmailSender, redisClient *redis.Client) *Worker {
	return &Worker{sender: sender, redisClient: redisClient}
}

func (w *Worker) Process(event domain.PaymentEvent) error {
	ctx := context.Background()

	key := fmt.Sprintf("notification:processed:%s", event.EventID)
	exists, err := w.redisClient.Exists(ctx, key).Result()
	if err == nil && exists > 0 {
		log.Printf("[Worker] Event %s already processed — skipping", event.EventID)
		return nil
	}

	// Retry with Exponential Backoff
	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := w.sender.Send(
			event.CustomerEmail,
			event.OrderID,
			event.Amount,
			event.Status,
		)
		if err == nil {
			w.redisClient.Set(ctx, key, "done", 24*time.Hour)
			log.Printf("[Worker] Event %s processed successfully on attempt %d", event.EventID, attempt)
			return nil
		}

		lastErr = err
		delay := baseDelay * time.Duration(1<<(attempt-1)) // 2s, 4s, 8s
		log.Printf("[Worker] Attempt %d failed: %v. Retrying in %s...", attempt, err, delay)
		time.Sleep(delay)
	}

	return fmt.Errorf("all %d attempts failed: %w", maxRetries, lastErr)
}
