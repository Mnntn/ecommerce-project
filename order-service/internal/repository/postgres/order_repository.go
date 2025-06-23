package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/mnntn/ecommerce-project/order-service/internal/domain"
	"github.com/mnntn/ecommerce-project/order-service/internal/outbox"
)

type OrderRepository struct {
	db *sqlx.DB
}

func NewOrderRepository(db *sqlx.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Create(ctx context.Context, order *domain.Order) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() // Rollback is ignored if tx is committed

	orderQuery := `
		INSERT INTO orders (id, user_id, status, total_amount, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING created_at, updated_at
	`
	err = tx.QueryRowxContext(ctx, orderQuery,
		order.ID, order.UserID, order.Status, order.TotalAmount, order.Description,
	).Scan(&order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return err
	}

	itemQuery := `
		INSERT INTO order_items (order_id, product_id, quantity, price)
		VALUES ($1, $2, $3, $4)
	`
	for _, item := range order.Items {
		_, err := tx.ExecContext(ctx, itemQuery, order.ID, item.ProductID, item.Quantity, item.Price)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *OrderRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	query := `SELECT id, user_id, status, total_amount, description, created_at, updated_at FROM orders WHERE id = $1`
	order := &domain.Order{}
	err := r.db.GetContext(ctx, order, query, id)
	if err == sql.ErrNoRows {
		return nil, nil // Not found is not an error
	}
	if err != nil {
		return nil, err
	}

	// TODO: Fetch order items as well
	return order, nil
}

func (r *OrderRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.Order, error) {
	query := `SELECT id, user_id, status, total_amount, description, created_at, updated_at FROM orders WHERE user_id = $1 ORDER BY created_at DESC`
	var orders []*domain.Order
	err := r.db.SelectContext(ctx, &orders, query, userID)
	if err != nil {
		return nil, err
	}
	// TODO: Fetch order items for each order
	return orders, nil
}

func (r *OrderRepository) GetAll(ctx context.Context) ([]*domain.Order, error) {
	query := `SELECT id, user_id, status, total_amount, description, created_at, updated_at FROM orders ORDER BY created_at DESC`
	var orders []*domain.Order
	err := r.db.SelectContext(ctx, &orders, query)
	if err != nil {
		return nil, err
	}
	// TODO: Fetch order items for each order
	return orders, nil
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.OrderStatus) error {
	query := `UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, status, id)
	return err
}

func (r *OrderRepository) CreateWithOutbox(ctx context.Context, order *domain.Order, outboxMsg *outbox.OutboxMessage) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() // Rollback is ignored if tx is committed

	orderQuery := `
		INSERT INTO orders (id, user_id, status, total_amount, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING created_at, updated_at
	`
	err = tx.QueryRowxContext(ctx, orderQuery,
		order.ID, order.UserID, order.Status, order.TotalAmount, order.Description,
	).Scan(&order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return err
	}

	itemQuery := `
		INSERT INTO order_items (order_id, product_id, quantity, price)
		VALUES ($1, $2, $3, $4)
	`
	for _, item := range order.Items {
		_, err := tx.ExecContext(ctx, itemQuery, order.ID, item.ProductID, item.Quantity, item.Price)
		if err != nil {
			return err
		}
	}

	// Сохраняем outbox сообщение
	outboxQuery := `
		INSERT INTO outbox_messages (id, type, payload, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err = tx.ExecContext(ctx, outboxQuery,
		outboxMsg.ID,
		outboxMsg.Type,
		outboxMsg.Payload,
		outboxMsg.Status,
		outboxMsg.CreatedAt,
		outboxMsg.UpdatedAt,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}
