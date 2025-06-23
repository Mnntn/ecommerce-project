package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
	"github.com/mnntn/ecommerce-project/payment-service/internal/domain"
)

var ErrAccountAlreadyExists = errors.New("account already exists")

type AccountRepository struct {
	db *sql.DB
}

func NewAccountRepository(db *sql.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

func (r *AccountRepository) Create(ctx context.Context, account *domain.Account) error {
	query := `
		INSERT INTO accounts (id, user_id, balance, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	now := time.Now()
	account.CreatedAt = now
	account.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, query,
		account.ID,
		account.UserID,
		account.Balance,
		account.CreatedAt,
		account.UpdatedAt,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return ErrAccountAlreadyExists
		}
		return err
	}

	return nil
}

func (r *AccountRepository) GetByUserID(ctx context.Context, userID string) (*domain.Account, error) {
	query := `
		SELECT id, user_id, balance, created_at, updated_at
		FROM accounts
		WHERE user_id = $1
	`

	account := &domain.Account{}
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&account.ID,
		&account.UserID,
		&account.Balance,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return account, nil
}

func (r *AccountRepository) UpdateBalance(ctx context.Context, userID string, amount float64) error {
	query := `
		UPDATE accounts
		SET balance = balance + $1, updated_at = $2
		WHERE user_id = $3
	`

	_, err := r.db.ExecContext(ctx, query, amount, time.Now(), userID)
	return err
}

func (r *AccountRepository) Update(ctx context.Context, account *domain.Account) error {
	query := `
		UPDATE accounts
		SET balance = $1, updated_at = $2
		WHERE user_id = $3
	`
	_, err := r.db.ExecContext(ctx, query, account.Balance, time.Now(), account.UserID)
	return err
}
