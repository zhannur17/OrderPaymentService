package messaging

import (
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"

	"notification-service/internal/domain"
)

const QueueName = "payment.completed"

type RabbitMQConsumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	worker  *Worker
}

func NewRabbitMQConsumer(amqpURL string, redisClient *redis.Client, sender domain.EmailSender) (*RabbitMQConsumer, error) {
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

	worker := NewWorker(sender, redisClient)

	log.Println("[Consumer] Connected to RabbitMQ, listening on:", QueueName)

	return &RabbitMQConsumer{
		conn:    conn,
		channel: ch,
		worker:  worker,
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

	if err := c.worker.Process(event); err != nil {
		log.Printf("[Consumer] Failed to process event %s: %v — sending NACK", event.EventID, err)
		msg.Nack(false, false)
		return
	}

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
