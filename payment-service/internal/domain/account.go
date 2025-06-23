package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Account struct {
	ID        uuid.UUID `json:"id"`
	UserID    string    `json:"user_id"`
	Balance   float64   `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AccountRepository interface {
	Create(ctx context.Context, account *Account) error
	GetByUserID(ctx context.Context, userID string) (*Account, error)
	UpdateBalance(ctx context.Context, userID string, amount float64) error
}

type AccountService interface {
	CreateAccount(userID string) (*Account, error)
	GetAccount(userID string) (*Account, error)
	Deposit(userID string, amount float64) error
	Withdraw(userID string, amount float64) error
}

// User represents a user in the system
// (added for user/account management)
type User struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type UserRepository interface {
	Create(user *User) error
	GetByID(id uuid.UUID) (*User, error)
	GetByName(name string) (*User, error)
	GetAll() ([]*User, error)
}
