package app

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"

	"payment-service/internal/repository"
	transporthttp "payment-service/internal/transport/http"
	"payment-service/internal/usecase"
)

type Config struct {
	DBConnStr string
	Port      string
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

	// Composition Root
	paymentRepo := repository.NewPostgresPaymentRepository(db)
	paymentUC := usecase.NewPaymentUseCase(paymentRepo)
	paymentHandler := transporthttp.NewPaymentHandler(paymentUC)

	router := transporthttp.NewRouter(paymentHandler)

	addr := ":" + cfg.Port
	log.Printf("Payment Service listening on %s", addr)
	return router.Run(addr)
}
