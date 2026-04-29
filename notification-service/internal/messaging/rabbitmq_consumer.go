package messaging

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"

	"notification-service/internal/domain"
)

const QueueName = "payment.completed"

type RabbitMQConsumer struct {
	conn         *amqp.Connection
	channel      *amqp.Channel
	mu           sync.Mutex
	processedIDs map[string]bool
}

func NewRabbitMQConsumer(amqpURL string) (*RabbitMQConsumer, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, fmt.Errorf("dial rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("open channel: %w", err)
	}

	_, err = ch.QueueDeclare(
		QueueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("declare queue: %w", err)
	}

	if err := ch.Qos(1, 0, false); err != nil {
		conn.Close()
		return nil, fmt.Errorf("set qos: %w", err)
	}

	log.Println("[Consumer] Connected to RabbitMQ, listening on:", QueueName)

	return &RabbitMQConsumer{
		conn:         conn,
		channel:      ch,
		processedIDs: make(map[string]bool),
	}, nil
}

func (c *RabbitMQConsumer) Start(quit <-chan struct{}) error {
	msgs, err := c.channel.Consume(
		QueueName,
		"",
		false, // auto-ack DISABLED
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("consume: %w", err)
	}

	log.Println("[Consumer] Waiting for messages...")

	for {
		select {
		case <-quit:
			log.Println("[Consumer] Graceful shutdown")
			return nil
		case msg, ok := <-msgs:
			if !ok {
				return fmt.Errorf("channel closed")
			}
			c.handleMessage(msg)
		}
	}
}

func (c *RabbitMQConsumer) handleMessage(msg amqp.Delivery) {
	var event domain.PaymentEvent
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		log.Printf("[Consumer] Failed to parse message: %v — sending NACK", err)
		msg.Nack(false, false)
		return
	}

	// Idempotency check
	c.mu.Lock()
	already := c.processedIDs[event.EventID]
	if !already {
		c.processedIDs[event.EventID] = true
	}
	c.mu.Unlock()

	if already {
		log.Printf("[Consumer] Duplicate event %s — skipping", event.EventID)
		msg.Ack(false)
		return
	}

	amountDollars := float64(event.Amount) / 100.0
	log.Printf("[Notification] Sent email to %s for Order #%s. Amount: $%.2f. Status: %s",
		event.CustomerEmail, event.OrderID, amountDollars, event.Status)

	msg.Ack(false)
}

func (c *RabbitMQConsumer) Close() {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
	log.Println("[Consumer] RabbitMQ connection closed")
}
