# AP2 Assignment 2 - gRPC Migration

## Repositories
- **Proto Repository**: https://github.com/zhannur17/ap2-protos
- **Generated Code Repository**: https://github.com/zhannur17/ap2-generated

## Architecture
- **Order Service** — REST server (Gin) on :8080, gRPC client → Payment Service, gRPC server on :50052
- **Payment Service** — gRPC server on :50051 with Logging Interceptor

## Communication
- External: REST (POST /orders, GET /orders/:id, PATCH /orders/:id/cancel)
- Internal: gRPC (Order Service → Payment Service)
- Streaming: gRPC Server-side streaming (Order Service → Client)

### Prerequisites
- Go 1.22+
- PostgreSQL

### 1. Start Payment Service
```cmd
cd payment-service
go run ./cmd/payment-service
```

### 2. Start Order Service
```cmd
cd order-service
go run ./cmd/order-service
```

### 3. Test gRPC call
```powershell
Invoke-WebRequest -Uri "http://localhost:8080/orders" -Method POST -ContentType "application/json" -Body '{"customer_id": "customer1", "item_name": "Test Item", "amount": 5000}'
```

## Environment Variables

### payment-service/.env
PAYMENT_DB_DSN=postgres://postgres:0000@localhost:5432/payments_db?sslmode=disable
GRPC_PORT=50051

### order-service/.env
ORDER_DB_DSN=postgres://postgres:0000@localhost:5432/orders_db?sslmode=disable
PAYMENT_GRPC_ADDR=localhost:50051
ORDER_PORT=8080
GRPC_PORT=50052

## Branch
- `master` — Assignment 1 (REST)
- `grpc-migration` — Assignment 2 (gRPC)


## Architecture Diagram
![Architecture](architecture.png)



## Assignment 3 — Event-Driven Architecture with RabbitMQ

### What was added
A new Notification Service that listens to payment events from RabbitMQ and simulates sending emails.

### Architecture

[Order Service] --gRPC--> [Payment Service] --RabbitMQ--> [Notification Service]
|
publishes event
payment.completed

### Event Flow
1. User sends POST /orders to Order Service
2. Order Service calls Payment Service via gRPC
3. Payment Service saves payment to DB
4. Payment Service publishes PaymentEvent to RabbitMQ queue payment.completed
5. Notification Service receives the event and logs a simulated email

### Idempotency Strategy
Each event has a unique event_id (UUID). The Notification Service stores all processed IDs in an in-memory map. If the same event arrives twice it is ACKed and skipped without printing the log again.

### ACK Logic
- Auto-ACK is disabled
- A message is ACKed only AFTER the log is successfully printed
- If processing fails NACK is sent and the message stays in the queue
- This guarantees at-least-once delivery

### Graceful Shutdown
Both Payment Service and Notification Service listen for SIGINT/SIGTERM signals. On shutdown they finish processing the current message and cleanly close RabbitMQ connections.


### RabbitMQ Management UI
Open http://localhost:15672
- Username: guest
- Password: guest