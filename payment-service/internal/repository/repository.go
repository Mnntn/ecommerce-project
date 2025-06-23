package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type PaymentRepository interface {
	CreatePayment(ctx context.Context, payment *Payment) error
	GetPayment(ctx context.Context, id string) (*Payment, error)
	UpdatePayment(ctx context.Context, payment *Payment) error
}

type Payment struct {
	ID        string    `db:"id"`
	OrderID   string    `db:"order_id"`
	Amount    float64   `db:"amount"`
	Status    string    `db:"status"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type PostgresPaymentRepository struct {
	db *sqlx.DB
}

func NewPaymentRepository(connString string) PaymentRepository {
	db, err := sqlx.Connect("postgres", connString)
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to database: %v", err))
	}
	return &PostgresPaymentRepository{db: db}
}

func (r *PostgresPaymentRepository) CreatePayment(ctx context.Context, payment *Payment) error {
	query := `
		INSERT INTO payments (id, order_id, amount, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.ExecContext(ctx, query,
		payment.ID,
		payment.OrderID,
		payment.Amount,
		payment.Status,
		payment.CreatedAt,
		payment.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create payment: %w", err)
	}
	return nil
}

func (r *PostgresPaymentRepository) GetPayment(ctx context.Context, id string) (*Payment, error) {
	var payment Payment
	query := `SELECT * FROM payments WHERE id = $1`
	err := r.db.GetContext(ctx, &payment, query, id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("payment not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}
	return &payment, nil
}

func (r *PostgresPaymentRepository) UpdatePayment(ctx context.Context, payment *Payment) error {
	query := `
		UPDATE payments
		SET status = $1, updated_at = $2
		WHERE id = $3
	`
	_, err := r.db.ExecContext(ctx, query,
		payment.Status,
		payment.UpdatedAt,
		payment.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}
	return nil
}
