package main

import (
	"log"
	"os"

	"payment-service/internal/app"
)

func main() {
	cfg := app.Config{
		DBConnStr: getEnv("PAYMENT_DB_DSN", "postgres://postgres:0000@localhost:5432/payments_db?sslmode=disable"),
		Port:      getEnv("PAYMENT_PORT", "8081"),
	}

	if err := app.Run(cfg); err != nil {
		log.Fatalf("Payment service error: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
