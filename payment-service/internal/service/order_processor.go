package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/mnntn/ecommerce-project/payment-service/internal/domain"
	"github.com/mnntn/ecommerce-project/payment-service/internal/inbox"
	"github.com/mnntn/ecommerce-project/payment-service/internal/kafka"
	"github.com/mnntn/ecommerce-project/payment-service/internal/repository/postgres"
)

type OrderProcessor struct {
	accountRepo domain.AccountRepository
	producer    *kafka.Producer
	inboxRepo   *postgres.InboxRepository
	outboxRepo  *postgres.OutboxRepository
	db          *sql.DB
}

func NewOrderProcessor(accountRepo domain.AccountRepository, producer *kafka.Producer, inboxRepo *postgres.InboxRepository, outboxRepo *postgres.OutboxRepository, db *sql.DB) *OrderProcessor {
	return &OrderProcessor{
		accountRepo: accountRepo,
		producer:    producer,
		inboxRepo:   inboxRepo,
		outboxRepo:  outboxRepo,
		db:          db,
	}
}

func (p *OrderProcessor) ProcessOrderCreated(ctx context.Context, event *domain.OrderCreatedEvent) error {
	// Проверяем, не обработано ли уже событие (по id заказа)
	inboxID := uuid.New() // Можно использовать event.OrderID, если он UUID
	payload, _ := json.Marshal(event)
	inboxMsg := &inbox.InboxMessage{
		ID:        inboxID,
		Type:      "order_created",
		Payload:   payload,
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Проверяем, есть ли уже такое событие в inbox
	var exists bool
	err = tx.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM inbox_messages WHERE id = $1)", inboxMsg.ID).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return nil // Уже обработано
	}

	// Сохраняем событие в inbox
	_, err = tx.ExecContext(ctx, `INSERT INTO inbox_messages (id, type, payload, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`,
		inboxMsg.ID, inboxMsg.Type, inboxMsg.Payload, inboxMsg.Status, inboxMsg.CreatedAt, inboxMsg.UpdatedAt)
	if err != nil {
		return err
	}

	// Получаем аккаунт пользователя
	var balance float64
	err = tx.QueryRowContext(ctx, "SELECT balance FROM accounts WHERE user_id = $1 FOR UPDATE", event.UserID).Scan(&balance)
	if err != nil {
		p.markInboxProcessedTx(ctx, tx, inboxMsg.ID)
		return p.saveOutboxAndCommitTx(ctx, tx, event.OrderID.String(), "CANCELLED", "Account not found", inboxMsg.ID)
	}

	if balance < event.TotalAmount {
		p.markInboxProcessedTx(ctx, tx, inboxMsg.ID)
		return p.saveOutboxAndCommitTx(ctx, tx, event.OrderID.String(), "CANCELLED", "Insufficient balance", inboxMsg.ID)
	}

	// Списываем средства
	_, err = tx.ExecContext(ctx, "UPDATE accounts SET balance = balance - $1, updated_at = $2 WHERE user_id = $3", event.TotalAmount, time.Now(), event.UserID)
	if err != nil {
		p.markInboxProcessedTx(ctx, tx, inboxMsg.ID)
		return p.saveOutboxAndCommitTx(ctx, tx, event.OrderID.String(), "CANCELLED", "Failed to withdraw funds", inboxMsg.ID)
	}

	// Всё успешно — формируем событие FINISHED
	return p.saveOutboxAndCommitTx(ctx, tx, event.OrderID.String(), "FINISHED", "Payment successful", inboxMsg.ID)
}

func (p *OrderProcessor) saveOutboxAndCommitTx(ctx context.Context, tx *sql.Tx, orderID, status, reason string, inboxID uuid.UUID) error {
	// Формируем событие
	event := domain.OrderStatusUpdatedEvent{
		OrderID: orderID,
		Status:  status,
		Reason:  reason,
	}
	payload, _ := json.Marshal(event)
	outboxMsg := &postgres.OutboxMessage{
		ID:        uuid.New(),
		Type:      "order_status_updated",
		Payload:   payload,
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_, err := tx.ExecContext(ctx, `INSERT INTO outbox_messages (id, type, payload, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`,
		outboxMsg.ID, outboxMsg.Type, outboxMsg.Payload, outboxMsg.Status, outboxMsg.CreatedAt, outboxMsg.UpdatedAt)
	if err != nil {
		return err
	}
	// Помечаем inbox processed
	if err := p.markInboxProcessedTx(ctx, tx, inboxID); err != nil {
		return err
	}
	return tx.Commit()
}

func (p *OrderProcessor) markInboxProcessedTx(ctx context.Context, tx *sql.Tx, id uuid.UUID) error {
	_, err := tx.ExecContext(ctx, `UPDATE inbox_messages SET status = 'processed', updated_at = $1 WHERE id = $2`, time.Now(), id)
	return err
}

func (p *OrderProcessor) publishOrderStatusUpdate(ctx context.Context, orderID string, status, reason string) error {
	event := domain.OrderStatusUpdatedEvent{
		OrderID: orderID,
		Status:  status,
		Reason:  reason,
	}

	if err := p.producer.PublishOrderStatusUpdated(ctx, event); err != nil {
		log.Printf("Failed to publish order status update: %v", err)
		return err
	}

	log.Printf("Published order status update: OrderID=%s, Status=%s, Reason=%s", orderID, status, reason)
	return nil
}
