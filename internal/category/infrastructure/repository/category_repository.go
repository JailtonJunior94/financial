package repository

import (
	"context"
	"database/sql"

	"github.com/jailtonjunior94/financial/internal/category/domain/entities"
	"github.com/jailtonjunior94/financial/internal/category/domain/interfaces"
	"github.com/jailtonjunior94/financial/pkg/o11y"
)

type categoryRepository struct {
	db   *sql.DB
	o11y o11y.Observability
}

func NewCategoryRepository(db *sql.DB, o11y o11y.Observability) interfaces.CategoryRepository {
	return &categoryRepository{
		db:   db,
		o11y: o11y,
	}
}

func (r *categoryRepository) Find(ctx context.Context, userID string) ([]*entities.Category, error) {
	return nil, nil
}

func (r *categoryRepository) FindByID(ctx context.Context, userID, id string) (*entities.Category, error) {
	return nil, nil
}

func (r *categoryRepository) Insert(ctx context.Context, category *entities.Category) (*entities.Category, error) {
	ctx, span := r.o11y.Start(ctx, "category_repository.insert")
	defer span.End()

	query := `insert into
				categories (
					id,
					user_id,
					name,
					sequence,
					created_at,
					updated_at,
					active
				)
				values
					(?, ?, ?, ?, ?, ?, ?)`

	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		span.AddAttributes(
			ctx, o11y.Error, "error creating category",
			o11y.Attributes{Key: "user_id", Value: category.UserID},
			o11y.Attributes{Key: "error", Value: err},
		)
		return nil, err
	}

	_, err = stmt.ExecContext(
		ctx,
		category.ID,
		category.UserID,
		category.Name,
		category.Sequence,
		category.CreatedAt,
		category.UpdatedAt,
		category.Active,
	)
	if err != nil {
		span.AddAttributes(
			ctx, o11y.Error, "error creating category",
			o11y.Attributes{Key: "user_id", Value: category.UserID},
			o11y.Attributes{Key: "error", Value: err},
		)
		return nil, err
	}
	return category, nil
}

func (r *categoryRepository) Update(ctx context.Context, category *entities.Category) (*entities.Category, error) {
	return nil, nil
}
