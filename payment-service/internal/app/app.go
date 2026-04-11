package app

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"

	paymentv1 "github.com/zhannur17/ap2-generated/payment/v1"
	"payment-service/internal/repository"
	transportgrpc "payment-service/internal/transport/grpc"
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

	paymentRepo := repository.NewPostgresPaymentRepository(db)
	paymentUC := usecase.NewPaymentUseCase(paymentRepo)

	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
	}

	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		return fmt.Errorf("listen grpc: %w", err)
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(transportgrpc.LoggingInterceptor),
	)
	paymentv1.RegisterPaymentServiceServer(grpcServer, transportgrpc.NewPaymentServer(paymentUC))

	log.Printf("Payment gRPC server listening on :%s", grpcPort)
	return grpcServer.Serve(lis)
}
