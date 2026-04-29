package app

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"

	paymentv1 "github.com/zhannur17/ap2-generated/payment/v1"
	"payment-service/internal/messaging"
	"payment-service/internal/repository"
	transportgrpc "payment-service/internal/transport/grpc"
	"payment-service/internal/usecase"
)

type Config struct {
	DBConnStr string
	Port      string
	AmqpURL   string
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

	publisher, err := messaging.NewRabbitMQPublisher(cfg.AmqpURL)
	if err != nil {
		return fmt.Errorf("rabbitmq publisher: %w", err)
	}
	defer publisher.Close()

	paymentRepo := repository.NewPostgresPaymentRepository(db)
	paymentUC := usecase.NewPaymentUseCase(paymentRepo, publisher)

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

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("[Payment] Shutting down gRPC server...")
		grpcServer.GracefulStop()
	}()

	log.Printf("Payment gRPC server listening on :%s", grpcPort)
	return grpcServer.Serve(lis)
}
