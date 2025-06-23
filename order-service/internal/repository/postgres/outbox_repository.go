package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/mnntn/ecommerce-project/order-service/internal/outbox"
)

type OutboxRepository struct {
	db *sql.DB
}

func NewOutboxRepository(db *sql.DB) *OutboxRepository {
	return &OutboxRepository{db: db}
}

func (r *OutboxRepository) Save(ctx context.Context, message *outbox.OutboxMessage) error {
	query := `
		INSERT INTO outbox_messages (id, type, payload, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.ExecContext(ctx, query,
		message.ID,
		message.Type,
		message.Payload,
		message.Status,
		message.CreatedAt,
		message.UpdatedAt,
	)

	return err
}

func (r *OutboxRepository) GetPending(ctx context.Context, limit int) ([]*outbox.OutboxMessage, error) {
	query := `
		SELECT id, type, payload, status, created_at, updated_at
		FROM outbox_messages
		WHERE status = 'pending'
		ORDER BY created_at ASC
		LIMIT $1
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*outbox.OutboxMessage
	for rows.Next() {
		msg := &outbox.OutboxMessage{}
		err := rows.Scan(
			&msg.ID,
			&msg.Type,
			&msg.Payload,
			&msg.Status,
			&msg.CreatedAt,
			&msg.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

func (r *OutboxRepository) MarkAsProcessed(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE outbox_messages
		SET status = 'processed', updated_at = $1
		WHERE id = $2
	`

	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	return err
} 