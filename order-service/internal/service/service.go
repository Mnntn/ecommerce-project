package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mnntn/ecommerce-project/order-service/internal/domain"
	"github.com/mnntn/ecommerce-project/order-service/internal/kafka"
	"github.com/mnntn/ecommerce-project/order-service/internal/outbox"
	"github.com/mnntn/ecommerce-project/order-service/internal/repository"
)

// Service encapsulates all business logic for the order service.
type Service struct {
	orderRepo   repository.OrderRepository
	productRepo repository.ProductRepository
	producer    *kafka.Producer
	outboxRepo  outbox.OutboxRepository
}

func New(orderRepo repository.OrderRepository, productRepo repository.ProductRepository, producer *kafka.Producer, outboxRepo outbox.OutboxRepository) *Service {
	return &Service{
		orderRepo:   orderRepo,
		productRepo: productRepo,
		producer:    producer,
		outboxRepo:  outboxRepo,
	}
}

// CreateOrder handles the business logic of creating a new order.
func (s *Service) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*domain.Order, error) {
	if len(req.Items) == 0 {
		return nil, fmt.Errorf("order must contain at least one item")
	}

	productIDs := make([]int64, len(req.Items))
	for i, item := range req.Items {
		productIDs[i] = item.ProductID
	}

	products, err := s.productRepo.GetProductsByIDs(ctx, productIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get products: %w", err)
	}

	if len(products) != len(productIDs) {
		return nil, fmt.Errorf("one or more products not found")
	}

	productsMap := make(map[int64]*domain.Product)
	for _, p := range products {
		productsMap[p.ID] = p
	}

	// Создаем описание заказа
	description := "Order with items: "
	for i, item := range req.Items {
		product := productsMap[item.ProductID]
		if i > 0 {
			description += ", "
		}
		description += fmt.Sprintf("%s (x%d)", product.Name, item.Quantity)
	}

	order := &domain.Order{
		ID:          uuid.New(),
		UserID:      req.UserID,
		Status:      domain.StatusNew,
		Description: description,
		Items:       make([]domain.OrderItem, len(req.Items)),
		TotalAmount: 0,
	}

	for i, item := range req.Items {
		product := productsMap[item.ProductID]
		itemTotal := product.Price * float64(item.Quantity)
		order.TotalAmount += itemTotal
		order.Items[i] = domain.OrderItem{
			ProductID: product.ID,
			Quantity:  item.Quantity,
			Price:     product.Price,
		}
	}

	// Формируем outbox сообщение
	outboxEvent := domain.OrderCreatedEvent{
		OrderID:     order.ID,
		UserID:      order.UserID,
		TotalAmount: order.TotalAmount,
	}
	payload, err := json.Marshal(outboxEvent)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal outbox event: %w", err)
	}
	outboxMsg := &outbox.OutboxMessage{
		ID:        uuid.New(),
		Type:      "order_created",
		Payload:   payload,
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Сохраняем заказ и outbox сообщение в одной транзакции
	if err := s.orderRepo.CreateWithOutbox(ctx, order, outboxMsg); err != nil {
		return nil, fmt.Errorf("failed to create order and outbox message in db: %w", err)
	}

	return order, nil
}

// ListProducts returns all available products.
func (s *Service) ListProducts(ctx context.Context) ([]*domain.Product, error) {
	return s.productRepo.GetAll(ctx)
}

// GetOrderByID retrieves an order by its ID.
func (s *Service) GetOrderByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	// This should also retrieve order items. We'll update the repo method later.
	return s.orderRepo.GetByID(ctx, id)
}

// GetUserOrders retrieves all orders for a specific user.
func (s *Service) GetUserOrders(ctx context.Context, userID string) ([]*domain.Order, error) {
	return s.orderRepo.GetByUserID(ctx, userID)
}

// GetAllOrders retrieves all orders.
func (s *Service) GetAllOrders(ctx context.Context) ([]*domain.Order, error) {
	return s.orderRepo.GetAll(ctx)
}

// UpdateOrderStatus updates the status of an order. Used by Kafka consumer.
func (s *Service) UpdateOrderStatus(ctx context.Context, orderID uuid.UUID, status domain.OrderStatus) error {
	return s.orderRepo.UpdateStatus(ctx, orderID, status)
}
