package domain

import (
	"time"

	"github.com/google/uuid"
)

type OrderStatus string

const (
	StatusNew       OrderStatus = "NEW"
	StatusFinished  OrderStatus = "FINISHED"
	StatusCancelled OrderStatus = "CANCELLED"
)

type Order struct {
	ID          uuid.UUID   `json:"id" db:"id"`
	UserID      string      `json:"user_id" db:"user_id"`
	Items       []OrderItem `json:"items"`
	TotalAmount float64     `json:"total_amount" db:"total_amount"`
	Description string      `json:"description" db:"description"`
	Status      OrderStatus `json:"status" db:"status"`
	CreatedAt   time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at" db:"updated_at"`
}

type OrderItem struct {
	ID        int64     `json:"id" db:"id"`
	OrderID   uuid.UUID `json:"order_id" db:"order_id"`
	ProductID int64     `json:"product_id" db:"product_id"`
	Quantity  int       `json:"quantity" db:"quantity"`
	Price     float64   `json:"price" db:"price"` // Price at the time of order
}

type OrderRepository interface {
	Create(order *Order) error
	GetByID(id uuid.UUID) (*Order, error)
	GetByUserID(userID string) ([]*Order, error)
	GetAll() ([]*Order, error)
	UpdateStatus(id uuid.UUID, status OrderStatus) error
}

type OrderService interface {
	CreateOrder(userID string, amount float64, description string) (*Order, error)
	GetOrder(id uuid.UUID) (*Order, error)
	GetUserOrders(userID string) ([]*Order, error)
	GetAllOrders() ([]*Order, error)
	UpdateOrderStatus(id uuid.UUID, status OrderStatus) error
}
