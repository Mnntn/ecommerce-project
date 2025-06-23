package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mnntn/ecommerce-project/payment-service/internal/domain"
	"github.com/mnntn/ecommerce-project/payment-service/internal/inbox"
)

type AccountRepository interface {
	GetByUserID(ctx context.Context, userID string) (*domain.Account, error)
	Update(ctx context.Context, account *domain.Account) error
	Create(ctx context.Context, account *domain.Account) error
}

type AccountService struct {
	repo     AccountRepository
	userRepo domain.UserRepository
}

func NewAccountService(repo AccountRepository, userRepo domain.UserRepository) *AccountService {
	return &AccountService{
		repo:     repo,
		userRepo: userRepo,
	}
}

func (s *AccountService) GetBalance(ctx context.Context, userID string) (float64, error) {
	account, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return 0, err
	}
	return account.Balance, nil
}

func (s *AccountService) Deposit(ctx context.Context, userID string, amount float64) error {
	account, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}

	account.Balance += amount
	account.UpdatedAt = time.Now()

	return s.repo.Update(ctx, account)
}

func (s *AccountService) Withdraw(ctx context.Context, userID string, amount float64) error {
	account, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}

	if account.Balance < amount {
		return fmt.Errorf("insufficient funds")
	}

	account.Balance -= amount
	account.UpdatedAt = time.Now()

	return s.repo.Update(ctx, account)
}

func (s *AccountService) HandleMessage(ctx context.Context, message *inbox.InboxMessage) error {
	switch message.Type {
	case "payment_request":
		var payload struct {
			UserID string  `json:"user_id"`
			Amount float64 `json:"amount"`
		}
		if err := json.Unmarshal([]byte(message.Payload), &payload); err != nil {
			return fmt.Errorf("failed to unmarshal payment request: %w", err)
		}

		if err := s.Deposit(ctx, payload.UserID, payload.Amount); err != nil {
			return fmt.Errorf("failed to process payment: %w", err)
		}

		return nil
	default:
		return fmt.Errorf("unknown message type: %s", message.Type)
	}
}

func (s *AccountService) CreateAccount(ctx context.Context, userID string) (*domain.Account, error) {
	account := &domain.Account{
		ID:        uuid.New(),
		UserID:    userID,
		Balance:   0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := s.repo.Create(ctx, account); err != nil {
		return nil, err
	}
	return account, nil
}

func (s *AccountService) GetAccount(ctx context.Context, userID string) (*domain.Account, error) {
	return s.repo.GetByUserID(ctx, userID)
}

func (s *AccountService) GetUser(ctx context.Context, userID string) (*domain.User, error) {
	return s.userRepo.GetByID(uuid.MustParse(userID))
}

func (s *AccountService) CreateUser(ctx context.Context, name string) (*domain.User, error) {
	user := &domain.User{
		ID:        uuid.New(),
		Name:      name,
		CreatedAt: time.Now(),
	}
	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *AccountService) GetAllUsers(ctx context.Context) ([]*domain.User, error) {
	if getter, ok := s.userRepo.(interface {
		GetAll() ([]*domain.User, error)
	}); ok {
		return getter.GetAll()
	}
	return nil, fmt.Errorf("GetAll not implemented in userRepo")
}
