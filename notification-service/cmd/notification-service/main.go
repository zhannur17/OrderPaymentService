package main

import (
	"log"
	"os"

	"notification-service/internal/app"
)

func main() {
	cfg := app.Config{
		AmqpURL: getEnv("AMQP_URL", "amqp://guest:guest@localhost:5672/"),
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
