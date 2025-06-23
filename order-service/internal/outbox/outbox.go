package outbox

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type OutboxMessage struct {
	ID        uuid.UUID       `json:"id"`
	Type      string          `json:"type"`
	Payload   json.RawMessage `json:"payload"`
	Status    string          `json:"status"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type OutboxRepository interface {
	Save(ctx context.Context, message *OutboxMessage) error
	GetPending(ctx context.Context, limit int) ([]*OutboxMessage, error)
	MarkAsProcessed(ctx context.Context, id uuid.UUID) error
}

type OutboxPublisher interface {
	Publish(ctx context.Context, message *OutboxMessage) error
}

type OutboxProcessor struct {
	repo      OutboxRepository
	publisher OutboxPublisher
}

func NewOutboxProcessor(repo OutboxRepository, publisher OutboxPublisher) *OutboxProcessor {
	return &OutboxProcessor{
		repo:      repo,
		publisher: publisher,
	}
}

func (p *OutboxProcessor) Start() {
	go func() {
		for {
			if err := p.ProcessPendingMessages(context.Background()); err != nil {
				// Log error but continue processing
				continue
			}
			time.Sleep(time.Second)
		}
	}()
}

func (p *OutboxProcessor) ProcessPendingMessages(ctx context.Context) error {
	messages, err := p.repo.GetPending(ctx, 10)
	if err != nil {
		return err
	}

	for _, message := range messages {
		if err := p.publisher.Publish(ctx, message); err != nil {
			return err
		}

		if err := p.repo.MarkAsProcessed(ctx, message.ID); err != nil {
			return err
		}
	}

	return nil
}

func CreatePaymentMessage(orderID uuid.UUID, userID string, amount float64) (*OutboxMessage, error) {
	payload := struct {
		OrderID uuid.UUID `json:"order_id"`
		UserID  string    `json:"user_id"`
		Amount  float64   `json:"amount"`
	}{
		OrderID: orderID,
		UserID:  userID,
		Amount:  amount,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return &OutboxMessage{
		ID:        uuid.New(),
		Type:      "payment_request",
		Payload:   payloadBytes,
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}
