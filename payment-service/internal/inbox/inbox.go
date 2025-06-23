package inbox

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type InboxMessage struct {
	ID        uuid.UUID       `json:"id"`
	Type      string          `json:"type"`
	Payload   json.RawMessage `json:"payload"`
	Status    string          `json:"status"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type InboxRepository interface {
	Save(ctx context.Context, message *InboxMessage) error
	GetPending(ctx context.Context, limit int) ([]*InboxMessage, error)
	MarkAsProcessed(ctx context.Context, id uuid.UUID) error
}

type MessageHandler interface {
	HandleMessage(ctx context.Context, message *InboxMessage) error
}

type InboxProcessor struct {
	repo    InboxRepository
	handler MessageHandler
}

func NewInboxProcessor(repo InboxRepository, handler MessageHandler) *InboxProcessor {
	return &InboxProcessor{
		repo:    repo,
		handler: handler,
	}
}

func (p *InboxProcessor) Start() {
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

func (p *InboxProcessor) ProcessPendingMessages(ctx context.Context) error {
	messages, err := p.repo.GetPending(ctx, 10)
	if err != nil {
		return err
	}

	for _, message := range messages {
		if err := p.handler.HandleMessage(ctx, message); err != nil {
			return err
		}

		if err := p.repo.MarkAsProcessed(ctx, message.ID); err != nil {
			return err
		}
	}

	return nil
}

func CreatePaymentRequestMessage(orderID uuid.UUID, userID string, amount float64) (*InboxMessage, error) {
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

	return &InboxMessage{
		ID:        uuid.New(),
		Type:      "payment_request",
		Payload:   payloadBytes,
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}
