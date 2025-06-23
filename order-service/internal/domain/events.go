package domain

import "github.com/google/uuid"

// OrderCreatedEvent is published when a new order is successfully created.
type OrderCreatedEvent struct {
	OrderID     uuid.UUID `json:"order_id"`
	UserID      string    `json:"user_id"`
	TotalAmount float64   `json:"total_amount"`
}

// OrderStatusUpdatedEvent order status update event
type OrderStatusUpdatedEvent struct {
	OrderID string `json:"order_id"`
	Status  string `json:"status"`
	Reason  string `json:"reason,omitempty"`
}
