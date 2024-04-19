package repository

import (
	"context"
	"database/sql"

	"github.com/jailtonjunior94/financial/internal/category/domain/entities"
	"github.com/jailtonjunior94/financial/internal/category/domain/interfaces"
)

type categoryRepository struct {
	db *sql.DB
}

func NewCategoryRepository(db *sql.DB) interfaces.CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) Find(ctx context.Context, userID string) ([]*entities.Category, error) {
	return nil, nil
}

func (r *categoryRepository) FindByID(ctx context.Context, userID, id string) (*entities.Category, error) {
	return nil, nil
}

func (r *categoryRepository) Insert(ctx context.Context, category *entities.Category) (*entities.Category, error) {
	stmt, err := r.db.Prepare("insert into categories (id, user_id, name, sequence, created_at, updated_at, active) values (?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return nil, err
	}

	_, err = stmt.Exec(
		category.ID,
		category.UserID,
		category.Name,
		category.Sequence,
		category.CreatedAt,
		category.UpdatedAt,
		category.Active,
	)
	if err != nil {
		return nil, err
	}
	return category, nil
}

func (r *categoryRepository) Update(ctx context.Context, category *entities.Category) (*entities.Category, error) {
	return nil, nil
}
