package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/mnntn/ecommerce-project/order-service/internal/domain"
	"github.com/mnntn/ecommerce-project/order-service/internal/outbox"
)

type OrderRepository interface {
	Create(ctx context.Context, order *domain.Order) error
	CreateWithOutbox(ctx context.Context, order *domain.Order, outboxMsg *outbox.OutboxMessage) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error)
	GetByUserID(ctx context.Context, userID string) ([]*domain.Order, error)
	GetAll(ctx context.Context) ([]*domain.Order, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.OrderStatus) error
}

type ProductRepository interface {
	GetProductsByIDs(ctx context.Context, ids []int64) ([]*domain.Product, error)
	GetAll(ctx context.Context) ([]*domain.Product, error)
}
