package postgres

import (
	"context"
	"log"
	"time"
)

type OutboxPublisher interface {
	Publish(ctx context.Context, message *OutboxMessage) error
}

type OutboxProcessor struct {
	repo      *OutboxRepository
	publisher OutboxPublisher
}

func NewOutboxProcessor(repo *OutboxRepository, publisher OutboxPublisher) *OutboxProcessor {
	return &OutboxProcessor{
		repo:      repo,
		publisher: publisher,
	}
}

func (p *OutboxProcessor) Start() {
	go func() {
		for {
			if err := p.ProcessPendingMessages(context.Background()); err != nil {
				log.Printf("OutboxProcessor error: %v", err)
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
	for _, msg := range messages {
		if err := p.publisher.Publish(ctx, msg); err != nil {
			return err
		}
		if err := p.repo.MarkAsProcessed(ctx, msg.ID); err != nil {
			return err
		}
	}
	return nil
}
