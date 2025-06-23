package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/mnntn/ecommerce-project/order-service/internal/domain"
	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafka.Reader
}

func NewConsumer(brokers []string, groupID string) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    "payments",
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

func (c *Consumer) ReadOrderStatusUpdatedEvent(ctx context.Context) (*domain.OrderStatusUpdatedEvent, error) {
	msg, err := c.reader.ReadMessage(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read message: %w", err)
	}

	var event domain.OrderStatusUpdatedEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return nil, fmt.Errorf("failed to unmarshal OrderStatusUpdatedEvent: %w", err)
	}

	log.Printf("OrderStatusUpdatedEvent received: OrderID=%s, Status=%s, Reason=%s",
		event.OrderID, event.Status, event.Reason)

	return &event, nil
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
