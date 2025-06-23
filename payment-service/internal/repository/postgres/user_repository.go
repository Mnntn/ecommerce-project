package postgres

import (
	"database/sql"

	"github.com/google/uuid"
	"github.com/mnntn/ecommerce-project/payment-service/internal/domain"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *domain.User) error {
	query := `
		INSERT INTO users (id, name, created_at)
		VALUES ($1, $2, $3)
	`
	_, err := r.db.Exec(query, user.ID, user.Name, user.CreatedAt)
	return err
}

func (r *UserRepository) GetByID(id uuid.UUID) (*domain.User, error) {
	query := `
		SELECT id, name, created_at FROM users WHERE id = $1
	`
	user := &domain.User{}
	err := r.db.QueryRow(query, id).Scan(&user.ID, &user.Name, &user.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) GetByName(name string) (*domain.User, error) {
	query := `
		SELECT id, name, created_at FROM users WHERE name = $1
	`
	user := &domain.User{}
	err := r.db.QueryRow(query, name).Scan(&user.ID, &user.Name, &user.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) GetAll() ([]*domain.User, error) {
	rows, err := r.db.Query("SELECT id, name, created_at FROM users ORDER BY created_at ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		user := &domain.User{}
		if err := rows.Scan(&user.ID, &user.Name, &user.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}
