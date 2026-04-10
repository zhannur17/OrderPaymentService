package app

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/lib/pq"

	"order-service/internal/repository"
	transporthttp "order-service/internal/transport/http"
	"order-service/internal/usecase"
)

type Config struct {
	DBConnStr      string
	PaymentBaseURL string
	Port           string
}

func Run(cfg Config) error {
	db, err := sql.Open("postgres", cfg.DBConnStr)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	if err := db.Ping(); err != nil {
		return fmt.Errorf("ping db: %w", err)
	}
	log.Println("Connected to database")

	orderRepo := repository.NewPostgresOrderRepository(db)

	httpClient := &http.Client{Timeout: 2 * time.Second}
	paymentClient := repository.NewHTTPPaymentClient(httpClient, cfg.PaymentBaseURL)

	orderUC := usecase.NewOrderUseCase(orderRepo, paymentClient)
	orderHandler := transporthttp.NewOrderHandler(orderUC)

	router := transporthttp.NewRouter(orderHandler)

	addr := ":" + cfg.Port
	log.Printf("Order Service listening on %s", addr)
	return router.Run(addr)
}
