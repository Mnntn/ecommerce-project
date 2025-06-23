package domain

import (
	"github.com/google/uuid"
)

// OrderCreatedEvent событие создания заказа
type OrderCreatedEvent struct {
	OrderID     uuid.UUID `json:"order_id"`
	UserID      string    `json:"user_id"`
	TotalAmount float64   `json:"total_amount"`
}

// OrderStatusUpdatedEvent событие обновления статуса заказа
type OrderStatusUpdatedEvent struct {
	OrderID string `json:"order_id"`
	Status  string `json:"status"`
	Reason  string `json:"reason,omitempty"`
}
