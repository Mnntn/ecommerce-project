package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mnntn/ecommerce-project/payment-service/internal/kafka"
	"github.com/mnntn/ecommerce-project/payment-service/internal/repository"
)

type PaymentService struct {
	repo     repository.PaymentRepository
	producer *kafka.Producer
}

func NewPaymentService(repo repository.PaymentRepository, producer *kafka.Producer) *PaymentService {
	return &PaymentService{
		repo:     repo,
		producer: producer,
	}
}

type PaymentRequest struct {
	OrderID string  `json:"order_id"`
	Amount  float64 `json:"amount"`
}

func (s *PaymentService) ProcessPayment(ctx context.Context, msg []byte) error {
	var paymentReq PaymentRequest
	if err := json.Unmarshal(msg, &paymentReq); err != nil {
		return fmt.Errorf("failed to unmarshal payment request: %w", err)
	}

	payment := &repository.Payment{
		ID:        generateID(),
		OrderID:   paymentReq.OrderID,
		Amount:    paymentReq.Amount,
		Status:    "processing",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.CreatePayment(ctx, payment); err != nil {
		return fmt.Errorf("failed to create payment: %w", err)
	}

	// Simulate payment processing
	time.Sleep(2 * time.Second)

	payment.Status = "completed"
	payment.UpdatedAt = time.Now()

	if err := s.repo.UpdatePayment(ctx, payment); err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}

	// Send payment response to Kafka
	paymentResponse := struct {
		OrderID string `json:"order_id"`
		Status  string `json:"status"`
	}{
		OrderID: payment.OrderID,
		Status:  payment.Status,
	}

	if err := s.producer.SendMessage(ctx, payment.OrderID, paymentResponse); err != nil {
		return fmt.Errorf("failed to send payment response: %w", err)
	}

	return nil
}

func generateID() string {
	return fmt.Sprintf("payment_%d", time.Now().UnixNano())
}
