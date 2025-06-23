package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/mnntn/ecommerce-project/payment-service/internal/domain"
	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafka.Reader
}

func NewConsumer(brokers []string, groupID string) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    "orders",
		GroupID:  groupID,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	return &Consumer{
		reader: reader,
	}
}

func (c *Consumer) ReadMessage(ctx context.Context) ([]byte, error) {
	msg, err := c.reader.ReadMessage(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read message: %w", err)
	}

	log.Printf("Message received from Kafka: %s", string(msg.Value))
	return msg.Value, nil
}

func (c *Consumer) ReadOrderCreatedEvent(ctx context.Context) (*domain.OrderCreatedEvent, error) {
	msg, err := c.reader.ReadMessage(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read message: %w", err)
	}

	var event domain.OrderCreatedEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return nil, fmt.Errorf("failed to unmarshal OrderCreatedEvent: %w", err)
	}

	log.Printf("OrderCreatedEvent received: OrderID=%s, UserID=%d, Amount=%.2f",
		event.OrderID, event.UserID, event.TotalAmount)

	return &event, nil
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
