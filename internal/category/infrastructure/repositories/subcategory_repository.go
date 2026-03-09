package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jailtonjunior94/financial/internal/category/domain/entities"
	"github.com/jailtonjunior94/financial/internal/category/domain/interfaces"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"

	"github.com/JailtonJunior94/devkit-go/pkg/database"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
)

type subcategoryRepository struct {
	db   database.DBTX
	o11y observability.Observability
	fm   *metrics.FinancialMetrics
}

func NewSubcategoryRepository(db database.DBTX, o11y observability.Observability, fm *metrics.FinancialMetrics) interfaces.SubcategoryRepository {
	return &subcategoryRepository{db: db, o11y: o11y, fm: fm}
}

func (r *subcategoryRepository) FindByID(ctx context.Context, userID, categoryID, id vos.UUID) (*entities.Subcategory, error) {
	start := time.Now()
	ctx, span := r.o11y.Tracer().Start(ctx, "subcategory_repository.find_by_id")
	defer span.End()

	query := `SELECT id, category_id, user_id, name, sequence, created_at, updated_at, deleted_at
          FROM subcategories
          WHERE id = $1 AND user_id = $2 AND category_id = $3 AND deleted_at IS NULL`

	var s entities.Subcategory
	err := r.db.QueryRowContext(ctx, query, id.String(), userID.String(), categoryID.String()).Scan(
		&s.ID.Value,
		&s.CategoryID.Value,
		&s.UserID.Value,
		&s.Name.Value,
		&s.Sequence.Sequence,
		&s.CreatedAt,
		&s.UpdatedAt,
		&s.DeletedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		r.fm.RecordRepositoryQuery(ctx, "find_by_id", "subcategory", time.Since(start))
		return nil, nil
	}
	if err != nil {
		span.RecordError(err)
		r.fm.RecordRepositoryFailure(ctx, "find_by_id", "subcategory", "infra", time.Since(start))
		return nil, fmt.Errorf("subcategory_repository.find_by_id: %w", err)
	}
	r.fm.RecordRepositoryQuery(ctx, "find_by_id", "subcategory", time.Since(start))
	return &s, nil
}

func (r *subcategoryRepository) FindByCategoryID(ctx context.Context, categoryID vos.UUID) ([]*entities.Subcategory, error) {
	start := time.Now()
	ctx, span := r.o11y.Tracer().Start(ctx, "subcategory_repository.find_by_category_id")
	defer span.End()

	query := `SELECT id, category_id, user_id, name, sequence, created_at, updated_at, deleted_at
          FROM subcategories
          WHERE category_id = $1 AND deleted_at IS NULL
          ORDER BY sequence, id`

	rows, err := r.db.QueryContext(ctx, query, categoryID.String())
	if err != nil {
		span.RecordError(err)
		r.fm.RecordRepositoryFailure(ctx, "find_by_category_id", "subcategory", "infra", time.Since(start))
		return nil, fmt.Errorf("subcategory_repository.find_by_category_id: %w", err)
	}
	defer func() { _ = rows.Close() }()

	subcategories := make([]*entities.Subcategory, 0)
	for rows.Next() {
		var s entities.Subcategory
		if err := rows.Scan(
			&s.ID.Value,
			&s.CategoryID.Value,
			&s.UserID.Value,
			&s.Name.Value,
			&s.Sequence.Sequence,
			&s.CreatedAt,
			&s.UpdatedAt,
			&s.DeletedAt,
		); err != nil {
			span.RecordError(err)
			r.fm.RecordRepositoryFailure(ctx, "find_by_category_id", "subcategory", "infra", time.Since(start))
			return nil, fmt.Errorf("subcategory_repository.find_by_category_id: %w", err)
		}
		subcategories = append(subcategories, &s)
	}
	if err := rows.Err(); err != nil {
		span.RecordError(err)
		r.fm.RecordRepositoryFailure(ctx, "find_by_category_id", "subcategory", "infra", time.Since(start))
		return nil, fmt.Errorf("subcategory_repository.find_by_category_id: %w", err)
	}

	r.fm.RecordRepositoryQuery(ctx, "find_by_category_id", "subcategory", time.Since(start))
	return subcategories, nil
}

func (r *subcategoryRepository) ListPaginated(ctx context.Context, params interfaces.ListSubcategoriesParams) ([]*entities.Subcategory, error) {
	start := time.Now()
	ctx, span := r.o11y.Tracer().Start(ctx, "subcategory_repository.list_paginated")
	defer span.End()

	whereClause := "category_id = $1 AND user_id = $2 AND deleted_at IS NULL"
	args := []interface{}{params.CategoryID.String(), params.UserID.String()}

	cursorSequence, hasSeq := params.Cursor.GetInt("sequence")
	cursorID, hasID := params.Cursor.GetString("id")

	if hasSeq && hasID && cursorID != "" {
		whereClause += ` AND (sequence > $3 OR (sequence = $3 AND id > $4))`
		args = append(args, cursorSequence, cursorID)
	}

	query := fmt.Sprintf(`
SELECT id, category_id, user_id, name, sequence, created_at, updated_at, deleted_at
FROM subcategories
WHERE %s
ORDER BY sequence ASC, id ASC
LIMIT $%d`, whereClause, len(args)+1)

	args = append(args, params.Limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		span.RecordError(err)
		r.fm.RecordRepositoryFailure(ctx, "list_paginated", "subcategory", "infra", time.Since(start))
		return nil, fmt.Errorf("subcategory_repository.list_paginated: %w", err)
	}
	defer func() { _ = rows.Close() }()

	subcategories := make([]*entities.Subcategory, 0)
	for rows.Next() {
		var s entities.Subcategory
		if err := rows.Scan(
			&s.ID.Value,
			&s.CategoryID.Value,
			&s.UserID.Value,
			&s.Name.Value,
			&s.Sequence.Sequence,
			&s.CreatedAt,
			&s.UpdatedAt,
			&s.DeletedAt,
		); err != nil {
			span.RecordError(err)
			r.fm.RecordRepositoryFailure(ctx, "list_paginated", "subcategory", "infra", time.Since(start))
			return nil, fmt.Errorf("subcategory_repository.list_paginated: %w", err)
		}
		subcategories = append(subcategories, &s)
	}
	if err := rows.Err(); err != nil {
		span.RecordError(err)
		r.fm.RecordRepositoryFailure(ctx, "list_paginated", "subcategory", "infra", time.Since(start))
		return nil, fmt.Errorf("subcategory_repository.list_paginated: %w", err)
	}

	r.fm.RecordRepositoryQuery(ctx, "list_paginated", "subcategory", time.Since(start))
	return subcategories, nil
}

func (r *subcategoryRepository) Save(ctx context.Context, subcategory *entities.Subcategory) error {
	start := time.Now()
	ctx, span := r.o11y.Tracer().Start(ctx, "subcategory_repository.save")
	defer span.End()

	query := `INSERT INTO subcategories (id, category_id, user_id, name, sequence, created_at)
          VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := r.db.ExecContext(ctx, query,
		subcategory.ID.Value,
		subcategory.CategoryID.Value,
		subcategory.UserID.Value,
		subcategory.Name.Value,
		subcategory.Sequence.Sequence,
		subcategory.CreatedAt.Ptr(),
	)
	if err != nil {
		span.RecordError(err)
		r.fm.RecordRepositoryFailure(ctx, "save", "subcategory", "infra", time.Since(start))
		return fmt.Errorf("subcategory_repository.save: %w", err)
	}

	r.fm.RecordRepositoryQuery(ctx, "save", "subcategory", time.Since(start))
	return nil
}

func (r *subcategoryRepository) Update(ctx context.Context, subcategory *entities.Subcategory) error {
	start := time.Now()
	ctx, span := r.o11y.Tracer().Start(ctx, "subcategory_repository.update")
	defer span.End()

	query := `UPDATE subcategories SET name = $1, sequence = $2, updated_at = $3
          WHERE id = $4 AND deleted_at IS NULL`

	_, err := r.db.ExecContext(ctx, query,
		subcategory.Name.Value,
		subcategory.Sequence.Sequence,
		subcategory.UpdatedAt.Ptr(),
		subcategory.ID.Value,
	)
	if err != nil {
		span.RecordError(err)
		r.fm.RecordRepositoryFailure(ctx, "update", "subcategory", "infra", time.Since(start))
		return fmt.Errorf("subcategory_repository.update: %w", err)
	}

	r.fm.RecordRepositoryQuery(ctx, "update", "subcategory", time.Since(start))
	return nil
}

func (r *subcategoryRepository) SoftDelete(ctx context.Context, id vos.UUID) error {
	start := time.Now()
	ctx, span := r.o11y.Tracer().Start(ctx, "subcategory_repository.soft_delete")
	defer span.End()

	_, err := r.db.ExecContext(ctx,
		`UPDATE subcategories SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`,
		id,
	)
	if err != nil {
		span.RecordError(err)
		r.fm.RecordRepositoryFailure(ctx, "soft_delete", "subcategory", "infra", time.Since(start))
		return fmt.Errorf("subcategory_repository.soft_delete: %w", err)
	}

	r.fm.RecordRepositoryQuery(ctx, "soft_delete", "subcategory", time.Since(start))
	return nil
}

func (r *subcategoryRepository) SoftDeleteByCategoryID(ctx context.Context, categoryID vos.UUID) error {
	start := time.Now()
	ctx, span := r.o11y.Tracer().Start(ctx, "subcategory_repository.soft_delete_by_category_id")
	defer span.End()

	_, err := r.db.ExecContext(ctx,
		`UPDATE subcategories SET deleted_at = NOW(), updated_at = NOW() WHERE category_id = $1 AND deleted_at IS NULL`,
		categoryID,
	)
	if err != nil {
		span.RecordError(err)
		r.fm.RecordRepositoryFailure(ctx, "soft_delete_by_category_id", "subcategory", "infra", time.Since(start))
		return fmt.Errorf("subcategory_repository.soft_delete_by_category_id: %w", err)
	}

	r.fm.RecordRepositoryQuery(ctx, "soft_delete_by_category_id", "subcategory", time.Since(start))
	return nil
}
