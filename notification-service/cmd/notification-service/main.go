package main

import (
	"log"
	"os"

	"notification-service/internal/app"
)

func main() {
	cfg := app.Config{
		AmqpURL:      getEnv("AMQP_URL", "amqp://guest:guest@localhost:5672/"),
		RedisURL:     getEnv("REDIS_URL", "localhost:6379"),
		ProviderMode: getEnv("PROVIDER_MODE", "SIMULATED"),
		SmtpHost:     getEnv("SMTP_HOST", ""),
		SmtpPort:     getEnv("SMTP_PORT", "587"),
		SmtpFrom:     getEnv("SMTP_FROM", ""),
	}

	log.Println("Starting Notification Service...")
	if err := app.Run(cfg); err != nil {
		log.Fatalf("Notification service error: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
