package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/mnntn/ecommerce-project/payment-service/internal/domain"
	"github.com/mnntn/ecommerce-project/payment-service/internal/repository/postgres"
	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(brokers []string) *Producer {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(brokers[0]),
		Topic:    "payments",
		Balancer: &kafka.LeastBytes{},
	}

	return &Producer{
		writer: writer,
	}
}

func (p *Producer) SendMessage(ctx context.Context, key string, value interface{}) error {
	valueBytes, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	msg := kafka.Message{
		Key:   []byte(key),
		Value: valueBytes,
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	log.Printf("Message sent to Kafka: %s", string(valueBytes))
	return nil
}

func (p *Producer) PublishOrderStatusUpdated(ctx context.Context, event domain.OrderStatusUpdatedEvent) error {
	return p.SendMessage(ctx, event.OrderID, event)
}

func (p *Producer) Close() error {
	return p.writer.Close()
}

// Publish реализует интерфейс OutboxPublisher
func (p *Producer) Publish(ctx context.Context, message *postgres.OutboxMessage) error {
	msg := kafka.Message{
		Key:   message.ID[:],
		Value: message.Payload,
	}
	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("failed to write message to topic: %w", err)
	}
	log.Printf("Outbox message sent to Kafka: %s", string(message.Payload))
	return nil
}
