package repositories

import (
	"context"
	"database/sql"

	"github.com/jailtonjunior94/financial/internal/category/domain/entities"
	"github.com/jailtonjunior94/financial/internal/category/domain/interfaces"

	"github.com/JailtonJunior94/devkit-go/pkg/o11y"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
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

func (r *categoryRepository) Find(ctx context.Context, userID vos.UUID) ([]*entities.Category, error) {
	ctx, span := r.o11y.Start(ctx, "category_repository.find_by_id")
	defer span.End()

	query := `select
				id,
				user_id,
				parent_id,
				name,
				sequence,
				created_at,
				updated_at,
				deleted_at
			from
				categories c
			where
				user_id = ?
				and deleted_at is null
			order by
				sequence;`

	rows, err := r.db.QueryContext(ctx, query, userID.String())
	if err != nil {
		span.AddAttributes(ctx, o11y.Error, "error finding categories", o11y.Attributes{Key: "user_id", Value: userID.String()})
		return nil, err
	}
	defer rows.Close()

	var categories []*entities.Category
	for rows.Next() {
		var category entities.Category
		err := rows.Scan(
			&category.ID.Value,
			&category.UserID.Value,
			&category.ParentID,
			&category.Name,
			&category.Sequence,
			&category.CreatedAt,
			&category.UpdatedAt.Time,
			&category.DeletedAt.Time,
		)
		if err != nil {
			span.AddAttributes(ctx, o11y.Error, "error scanning categories", o11y.Attributes{Key: "user_id", Value: userID.String()})
			return nil, err
		}
		categories = append(categories, &category)
	}
	return categories, nil
}

func (r *categoryRepository) FindByID(ctx context.Context, userID, id vos.UUID) (*entities.Category, error) {
	ctx, span := r.o11y.Start(ctx, "category_repository.find_by_id")
	defer span.End()

	query := `select
				id,
				user_id,
				parent_id,
				name,
				sequence,
				created_at,
				updated_at,
				deleted_at
			  from
				categories c
			  where
				c.user_id = ?
				and c.id = ?;`

	var category entities.Category
	err := r.db.QueryRowContext(ctx, query, userID.String(), id.String()).Scan(
		&category.ID.Value,
		&category.UserID.Value,
		&category.ParentID,
		&category.Name,
		&category.Sequence,
		&category.CreatedAt,
		&category.UpdatedAt.Time,
		&category.DeletedAt.Time,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			span.AddAttributes(ctx, o11y.Ok, "category not found", o11y.Attributes{Key: "id", Value: id.String()})
			return nil, nil
		}
		span.AddAttributes(ctx, o11y.Error, "error finding category", o11y.Attributes{Key: "id", Value: id.String()})
		return nil, err
	}
	return &category, nil
}

func (r *categoryRepository) Insert(ctx context.Context, category *entities.Category) (*entities.Category, error) {
	ctx, span := r.o11y.Start(ctx, "category_repository.insert")
	defer span.End()

	query := `insert into
				categories (
					id,
					user_id,
					parent_id,
					name,
					sequence,
					created_at,
					updated_at,
					deleted_at
				)
				values
					(?, ?, ?, ?, ?, ?, ?, ?)`

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
		category.ID.Value,
		category.UserID.Value,
		category.ParentID.SafeUUID(),
		category.Name,
		category.Sequence,
		category.CreatedAt,
		category.UpdatedAt.Time,
		category.DeletedAt.Time,
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
