package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/mnntn/ecommerce-project/order-service/internal/domain"
	"github.com/mnntn/ecommerce-project/order-service/internal/outbox"
	"github.com/segmentio/kafka-go"
)

const (
	OrderCreatedTopic = "orders"
)

type Producer struct {
	writer  *kafka.Writer
	brokers []string
}

func NewProducer(brokers []string) *Producer {
	// A writer is created for each topic. For now, we only have one.
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  brokers,
		Topic:    OrderCreatedTopic,
		Balancer: &kafka.LeastBytes{},
	})

	return &Producer{
		writer:  writer,
		brokers: brokers,
	}
}

func (p *Producer) PublishOrderCreated(ctx context.Context, event domain.OrderCreatedEvent) error {
	valueBytes, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal OrderCreatedEvent: %w", err)
	}

	msg := kafka.Message{
		Key:   []byte(event.OrderID.String()),
		Value: valueBytes,
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("failed to write message to topic %s: %w", OrderCreatedTopic, err)
	}

	log.Printf("Event %T sent to Kafka: %s", event, string(valueBytes))
	return nil
}

// SendMessage is a generic message sender, will be deprecated.
func (p *Producer) SendMessage(ctx context.Context, key string, value interface{}) error {
	valueBytes, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	msg := kafka.Message{
		Key:   []byte(key),
		Value: valueBytes,
	}

	// This generic method needs a topic, let's default to the old one for now.
	// This highlights the need for specific publish methods.
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  p.brokers,
		Topic:    "orders",
		Balancer: &kafka.LeastBytes{},
	})
	defer writer.Close()

	if err := writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	log.Printf("Message sent to Kafka: %s", string(valueBytes))
	return nil
}

func (p *Producer) Close() error {
	return p.writer.Close()
}

// Publish реализует интерфейс OutboxPublisher
func (p *Producer) Publish(ctx context.Context, message *outbox.OutboxMessage) error {
	msg := kafka.Message{
		Key:   message.ID[:],
		Value: message.Payload,
	}
	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("failed to write message to topic %s: %w", OrderCreatedTopic, err)
	}
	log.Printf("Outbox message sent to Kafka: %s", string(message.Payload))
	return nil
}
