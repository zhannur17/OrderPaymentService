package main

import (
	"log"
	"os"

	"order-service/internal/app"
)

func main() {
	cfg := app.Config{
		DBConnStr:       getEnv("ORDER_DB_DSN", "postgres://postgres:0000@localhost:5432/order_db?sslmode=disable"),
		PaymentGRPCAddr: getEnv("PAYMENT_GRPC_ADDR", "localhost:50051"),
		Port:            getEnv("ORDER_PORT", "8080"),
		RedisURL:        getEnv("REDIS_URL", "localhost:6379"),
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
