package app

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/redis/go-redis/v9"

	"notification-service/internal/adapter"
	"notification-service/internal/domain"
	"notification-service/internal/messaging"
)

type Config struct {
	AmqpURL      string
	RedisURL     string
	ProviderMode string
	SmtpHost     string
	SmtpPort     string
	SmtpFrom     string
}

func Run(cfg Config) error {
	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.RedisURL,
	})
	log.Println("[Notification] Connected to Redis")

	var sender domain.EmailSender
	if cfg.ProviderMode == "REAL" {
		sender = adapter.NewRealEmailSender(cfg.SmtpHost, cfg.SmtpPort, cfg.SmtpFrom)
		log.Println("[Notification] Using REAL email provider")
	} else {
		sender = adapter.NewSimulatedEmailSender()
		log.Println("[Notification] Using SIMULATED email provider")
	}

	consumer, err := messaging.NewRabbitMQConsumer(cfg.AmqpURL, redisClient, sender)
	if err != nil {
		return err
	}
	defer consumer.Close()

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
