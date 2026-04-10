package main

import (
	"log"
	"os"

	"order-service/internal/app"
)

func main() {
	cfg := app.Config{
		DBConnStr:      getEnv("ORDER_DB_DSN", "postgres://postgres:postgres@localhost:5432/orders_db?sslmode=disable"),
		PaymentBaseURL: getEnv("PAYMENT_SERVICE_URL", "http://localhost:8081"),
		Port:           getEnv("ORDER_PORT", "8080"),
	}

	if err := app.Run(cfg); err != nil {
		log.Fatalf("Order service error: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
