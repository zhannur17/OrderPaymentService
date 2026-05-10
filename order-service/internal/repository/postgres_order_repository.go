package repository

import (
	"database/sql"
	"fmt"
	"time"

	"order-service/internal/domain"
)

type PostgresOrderRepository struct {
	db *sql.DB
}

func NewPostgresOrderRepository(db *sql.DB) *PostgresOrderRepository {
	return &PostgresOrderRepository{db: db}
}

func (r *PostgresOrderRepository) Save(order *domain.Order) error {
	query := `
		INSERT INTO orders (id, customer_id, item_name, amount, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Exec(query,
		order.ID,
		order.CustomerID,
		order.ItemName,
		order.Amount,
		order.Status,
		order.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert order: %w", err)
	}
	return nil
}

func (r *PostgresOrderRepository) SaveWithIdempotencyKey(order *domain.Order, key string) error {
	query := `
		INSERT INTO orders (id, customer_id, item_name, amount, status, created_at, idempotency_key)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (idempotency_key) DO NOTHING
	`
	_, err := r.db.Exec(query,
		order.ID,
		order.CustomerID,
		order.ItemName,
		order.Amount,
		order.Status,
		order.CreatedAt,
		key,
	)
	if err != nil {
		return fmt.Errorf("insert order with idempotency key: %w", err)
	}
	return nil
}

func (r *PostgresOrderRepository) FindByID(id string) (*domain.Order, error) {
	query := `
		SELECT id, customer_id, item_name, amount, status, created_at
		FROM orders WHERE id = $1
	`
	row := r.db.QueryRow(query, id)
	return scanOrder(row)
}

func (r *PostgresOrderRepository) Update(order *domain.Order) error {
	query := `UPDATE orders SET status = $1 WHERE id = $2`
	res, err := r.db.Exec(query, order.Status, order.ID)
	if err != nil {
		return fmt.Errorf("update order: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrOrderNotFound
	}
	return nil
}

func (r *PostgresOrderRepository) FindByIdempotencyKey(key string) (*domain.Order, error) {
	query := `
		SELECT id, customer_id, item_name, amount, status, created_at
		FROM orders WHERE idempotency_key = $1
	`
	row := r.db.QueryRow(query, key)
	return scanOrder(row)
}

func (r *PostgresOrderRepository) FindByAmountRange(minAmount, maxAmount int64) ([]*domain.Order, error) {
	query := `
		SELECT id, customer_id, item_name, amount, status, created_at
		FROM orders WHERE amount >= $1 AND amount <= $2
	`
	rows, err := r.db.Query(query, minAmount, maxAmount)
	if err != nil {
		return nil, fmt.Errorf("query orders by amount range: %w", err)
	}
	defer rows.Close()

	var orders []*domain.Order
	for rows.Next() {
		var o domain.Order
		var createdAt time.Time
		if err := rows.Scan(&o.ID, &o.CustomerID, &o.ItemName, &o.Amount, &o.Status, &createdAt); err != nil {
			return nil, fmt.Errorf("scan order: %w", err)
		}
		o.CreatedAt = createdAt
		orders = append(orders, &o)
	}
	if orders == nil {
		orders = []*domain.Order{}
	}
	return orders, nil
}

func scanOrder(row *sql.Row) (*domain.Order, error) {
	var o domain.Order
	var createdAt time.Time
	err := row.Scan(
		&o.ID,
		&o.CustomerID,
		&o.ItemName,
		&o.Amount,
		&o.Status,
		&createdAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrOrderNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan order: %w", err)
	}
	o.CreatedAt = createdAt
	return &o, nil
}
