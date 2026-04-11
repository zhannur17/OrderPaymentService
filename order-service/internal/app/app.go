package app

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"

	orderv1 "github.com/zhannur17/ap2-generated/order/v1"
	"order-service/internal/repository"
	transportgrpc "order-service/internal/transport/grpc"
	transporthttp "order-service/internal/transport/http"
	"order-service/internal/usecase"
)

type Config struct {
	DBConnStr       string
	PaymentGRPCAddr string
	Port            string
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

	paymentClient, err := repository.NewGRPCPaymentClient(cfg.PaymentGRPCAddr)
	if err != nil {
		return fmt.Errorf("grpc payment client: %w", err)
	}

	orderUC := usecase.NewOrderUseCase(orderRepo, paymentClient)
	orderHandler := transporthttp.NewOrderHandler(orderUC)
	router := transporthttp.NewRouter(orderHandler)

	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50052"
	}

	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		return fmt.Errorf("listen grpc: %w", err)
	}

	grpcServer := grpc.NewServer()
	orderv1.RegisterOrderServiceServer(grpcServer, transportgrpc.NewOrderServer(orderRepo))

	go func() {
		log.Printf("Order gRPC server listening on :%s", grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("gRPC server error: %v", err)
		}
	}()

	addr := ":" + cfg.Port
	log.Printf("Order Service HTTP listening on %s", addr)
	return router.Run(addr)
}
