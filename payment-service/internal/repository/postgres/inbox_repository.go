package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/mnntn/ecommerce-project/payment-service/internal/inbox"
)

type InboxRepository struct {
	db *sql.DB
}

func NewInboxRepository(db *sql.DB) *InboxRepository {
	return &InboxRepository{db: db}
}

func (r *InboxRepository) Save(ctx context.Context, message *inbox.InboxMessage) error {
	query := `
		INSERT INTO inbox_messages (id, type, payload, status, created_at, updated_at)
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

func (r *InboxRepository) GetPending(ctx context.Context, limit int) ([]*inbox.InboxMessage, error) {
	query := `
		SELECT id, type, payload, status, created_at, updated_at
		FROM inbox_messages
		WHERE status = 'pending'
		ORDER BY created_at ASC
		LIMIT $1
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*inbox.InboxMessage
	for rows.Next() {
		msg := &inbox.InboxMessage{}
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

func (r *InboxRepository) MarkAsProcessed(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE inbox_messages
		SET status = 'processed', updated_at = $1
		WHERE id = $2
	`

	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	return err
}
