package postgres

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/mnntn/ecommerce-project/order-service/internal/domain"
)

type ProductRepository struct {
	db *sqlx.DB
}

func NewProductRepository(db *sqlx.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) GetAll(ctx context.Context) ([]*domain.Product, error) {
	var products []*domain.Product
	query := "SELECT id, name, description, price, created_at, updated_at FROM products"
	err := r.db.SelectContext(ctx, &products, query)
	if err != nil {
		return nil, err
	}
	return products, nil
}

func (r *ProductRepository) GetProductsByIDs(ctx context.Context, ids []int64) ([]*domain.Product, error) {
	var products []*domain.Product
	query, args, err := sqlx.In("SELECT id, name, description, price, created_at, updated_at FROM products WHERE id IN (?)", ids)
	if err != nil {
		return nil, err
	}
	query = r.db.Rebind(query)
	err = r.db.SelectContext(ctx, &products, query, args...)
	if err != nil {
		return nil, err
	}
	return products, nil
}
