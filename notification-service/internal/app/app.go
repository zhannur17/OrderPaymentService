package app

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"notification-service/internal/messaging"
)

type Config struct {
	AmqpURL string
}

func Run(cfg Config) error {
	consumer, err := messaging.NewRabbitMQConsumer(cfg.AmqpURL)
	if err != nil {
		return err
	}
	defer consumer.Close()

	// Graceful shutdown
	quit := make(chan struct{})
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("[Notification] Signal received, shutting down...")
		close(quit)
	}()

	return consumer.Start(quit)
}
