package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jailtonjunior94/financial/internal/category/domain/entities"
	"github.com/jailtonjunior94/financial/internal/category/domain/interfaces"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"

	"github.com/JailtonJunior94/devkit-go/pkg/database"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
)

type categoryRepository struct {
	db  database.DBTX
	o11y observability.Observability
	fm  *metrics.FinancialMetrics
}

func NewCategoryRepository(db database.DBTX, o11y observability.Observability, fm *metrics.FinancialMetrics) interfaces.CategoryRepository {
	return &categoryRepository{
		db:   db,
		o11y: o11y,
		fm:   fm,
	}
}

func (r *categoryRepository) List(ctx context.Context, userID vos.UUID) ([]*entities.Category, error) {
	start := time.Now()
	ctx, span := r.o11y.Tracer().Start(ctx, "category_repository.list")
	defer span.End()

	r.o11y.Logger().Debug(ctx, "query_started",
		observability.String("operation", "list"),
		observability.String("layer", "repository"),
		observability.String("entity", "category"),
		observability.String("user_id", userID.String()),
	)

	query := `select
				id,
				user_id,
				name,
				sequence,
				created_at,
				updated_at,
				deleted_at
			from
				categories c
			where
				user_id = $1
				and deleted_at is null
				and parent_id is null
			order by
				sequence;`

	rows, err := r.db.QueryContext(ctx, query, userID.String())
	if err != nil {
		span.RecordError(err)
		r.o11y.Logger().Error(ctx, "query_failed",
			observability.String("operation", "list"),
			observability.String("layer", "repository"),
			observability.String("entity", "category"),
			observability.String("user_id", userID.String()),
			observability.Error(err),
		)
		r.fm.RecordRepositoryFailure(ctx, "list", "category", "infra", time.Since(start))
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var categories []*entities.Category
	for rows.Next() {
		var category entities.Category
		err := rows.Scan(
			&category.ID.Value,
			&category.UserID.Value,
			&category.Name.Value,
			&category.Sequence.Sequence,
			&category.CreatedAt,
			&category.UpdatedAt,
			&category.DeletedAt,
		)
		if err != nil {
			span.RecordError(err)
			r.o11y.Logger().Error(ctx, "query_failed",
				observability.String("operation", "list"),
				observability.String("layer", "repository"),
				observability.String("entity", "category"),
				observability.String("user_id", userID.String()),
				observability.Error(err),
			)
			r.fm.RecordRepositoryFailure(ctx, "list", "category", "infra", time.Since(start))
			return nil, err
		}
		categories = append(categories, &category)
	}

	if err := rows.Err(); err != nil {
		span.RecordError(err)
		r.o11y.Logger().Error(ctx, "query_failed",
			observability.String("operation", "list"),
			observability.String("layer", "repository"),
			observability.String("entity", "category"),
			observability.String("user_id", userID.String()),
			observability.Error(err),
		)
		r.fm.RecordRepositoryFailure(ctx, "list", "category", "infra", time.Since(start))
		return nil, err
	}

	r.o11y.Logger().Debug(ctx, "query_completed",
		observability.String("operation", "list"),
		observability.String("layer", "repository"),
		observability.String("entity", "category"),
		observability.String("user_id", userID.String()),
	)
	r.fm.RecordRepositoryQuery(ctx, "list", "category", time.Since(start))

	return categories, nil
}

// ListPaginated lista categorias de um usuário com paginação cursor-based.
func (r *categoryRepository) ListPaginated(ctx context.Context, params interfaces.ListCategoriesParams) ([]*entities.Category, error) {
	start := time.Now()
	ctx, span := r.o11y.Tracer().Start(ctx, "category_repository.list_paginated")
	defer span.End()

	r.o11y.Logger().Debug(ctx, "query_started",
		observability.String("operation", "list_paginated"),
		observability.String("layer", "repository"),
		observability.String("entity", "category"),
		observability.String("user_id", params.UserID.String()),
	)

	// Build WHERE clause with cursor
	whereClause := "user_id = $1 AND deleted_at IS NULL AND parent_id IS NULL"
	args := []interface{}{params.UserID.String()}

	cursorSequence, hasSeq := params.Cursor.GetInt("sequence")
	cursorID, hasID := params.Cursor.GetString("id")

	if hasSeq && hasID && cursorID != "" {
		// Keyset pagination: WHERE (sequence, id) > (cursor_seq, cursor_id)
		whereClause += ` AND (
			sequence > $2
			OR (sequence = $2 AND id > $3)
		)`
		args = append(args, cursorSequence, cursorID)
	}

	query := fmt.Sprintf(`
		SELECT
			id,
			user_id,
			name,
			sequence,
			created_at,
			updated_at,
			deleted_at
		FROM categories
		WHERE %s
		ORDER BY sequence ASC, id ASC
		LIMIT $%d`, whereClause, len(args)+1)

	args = append(args, params.Limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		span.RecordError(err)
		r.o11y.Logger().Error(ctx, "query_failed",
			observability.String("operation", "list_paginated"),
			observability.String("layer", "repository"),
			observability.String("entity", "category"),
			observability.String("user_id", params.UserID.String()),
			observability.Error(err),
		)
		r.fm.RecordRepositoryFailure(ctx, "list_paginated", "category", "infra", time.Since(start))
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	categories := make([]*entities.Category, 0)
	for rows.Next() {
		var category entities.Category
		err := rows.Scan(
			&category.ID.Value,
			&category.UserID.Value,
			&category.Name.Value,
			&category.Sequence.Sequence,
			&category.CreatedAt,
			&category.UpdatedAt,
			&category.DeletedAt,
		)
		if err != nil {
			span.RecordError(err)
			r.o11y.Logger().Error(ctx, "query_failed",
				observability.String("operation", "list_paginated"),
				observability.String("layer", "repository"),
				observability.String("entity", "category"),
				observability.String("user_id", params.UserID.String()),
				observability.Error(err),
			)
			r.fm.RecordRepositoryFailure(ctx, "list_paginated", "category", "infra", time.Since(start))
			return nil, err
		}
		categories = append(categories, &category)
	}

	if err := rows.Err(); err != nil {
		span.RecordError(err)
		r.o11y.Logger().Error(ctx, "query_failed",
			observability.String("operation", "list_paginated"),
			observability.String("layer", "repository"),
			observability.String("entity", "category"),
			observability.String("user_id", params.UserID.String()),
			observability.Error(err),
		)
		r.fm.RecordRepositoryFailure(ctx, "list_paginated", "category", "infra", time.Since(start))
		return nil, err
	}

	r.o11y.Logger().Debug(ctx, "query_completed",
		observability.String("operation", "list_paginated"),
		observability.String("layer", "repository"),
		observability.String("entity", "category"),
		observability.String("user_id", params.UserID.String()),
	)
	r.fm.RecordRepositoryQuery(ctx, "list_paginated", "category", time.Since(start))

	return categories, nil
}

func (r *categoryRepository) FindByID(ctx context.Context, userID, id vos.UUID) (*entities.Category, error) {
	start := time.Now()
	ctx, span := r.o11y.Tracer().Start(ctx, "category_repository.find_by_id")
	defer span.End()

	r.o11y.Logger().Debug(ctx, "query_started",
		observability.String("operation", "find_by_id"),
		observability.String("layer", "repository"),
		observability.String("entity", "category"),
		observability.String("user_id", userID.String()),
		observability.String("category_id", id.String()),
	)

	// 1. Buscar categoria principal
	queryCategory := `
		SELECT
			id,
			user_id,
			parent_id,
			name,
			sequence,
			created_at,
			updated_at,
			deleted_at
		FROM categories
		WHERE user_id = $1
			AND id = $2
			AND deleted_at IS NULL`

	var category entities.Category
	var parentIDStr sql.NullString
	err := r.db.QueryRowContext(ctx, queryCategory, userID.String(), id.String()).Scan(
		&category.ID.Value,
		&category.UserID.Value,
		&parentIDStr,
		&category.Name.Value,
		&category.Sequence.Sequence,
		&category.CreatedAt,
		&category.UpdatedAt,
		&category.DeletedAt,
	)

	// Convert sql.NullString to *vos.UUID
	if parentIDStr.Valid && parentIDStr.String != "" {
		parentUUID, parseErr := vos.NewUUIDFromString(parentIDStr.String)
		if parseErr == nil {
			category.ParentID = &parentUUID
		}
	}

	if err == sql.ErrNoRows {
		// Check if category exists but belongs to a different user
		var existsForOtherUser bool
		checkQuery := `SELECT EXISTS(SELECT 1 FROM categories WHERE id = $1)`
		_ = r.db.QueryRowContext(ctx, checkQuery, id.String()).Scan(&existsForOtherUser)

		if existsForOtherUser {
			r.o11y.Logger().Warn(ctx, "query_completed",
				observability.String("operation", "find_by_id"),
				observability.String("layer", "repository"),
				observability.String("entity", "category"),
				observability.String("user_id", userID.String()),
				observability.String("category_id", id.String()),
				observability.String("result", "category_belongs_to_different_user"),
			)
		} else {
			r.o11y.Logger().Warn(ctx, "query_completed",
				observability.String("operation", "find_by_id"),
				observability.String("layer", "repository"),
				observability.String("entity", "category"),
				observability.String("user_id", userID.String()),
				observability.String("category_id", id.String()),
				observability.String("result", "not_found"),
			)
		}
		r.fm.RecordRepositoryQuery(ctx, "find_by_id", "category", time.Since(start))
		return nil, nil
	}
	if err != nil {
		span.RecordError(err)
		r.o11y.Logger().Error(ctx, "query_failed",
			observability.String("operation", "find_by_id"),
			observability.String("layer", "repository"),
			observability.String("entity", "category"),
			observability.String("user_id", userID.String()),
			observability.String("category_id", id.String()),
			observability.Error(err),
		)
		r.fm.RecordRepositoryFailure(ctx, "find_by_id", "category", "infra", time.Since(start))
		return nil, err
	}

	// 2. Buscar subcategorias (filhos)
	queryChildren := `
		SELECT
			id,
			user_id,
			name,
			sequence,
			created_at,
			updated_at,
			deleted_at
		FROM categories
		WHERE parent_id = $1
			AND deleted_at IS NULL
		ORDER BY sequence`

	rows, err := r.db.QueryContext(ctx, queryChildren, id.String())
	if err != nil {
		span.RecordError(err)
		r.o11y.Logger().Error(ctx, "query_failed",
			observability.String("operation", "find_by_id"),
			observability.String("layer", "repository"),
			observability.String("entity", "category"),
			observability.String("user_id", userID.String()),
			observability.String("category_id", id.String()),
			observability.Error(err),
		)
		r.fm.RecordRepositoryFailure(ctx, "find_by_id", "category", "infra", time.Since(start))
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var children []entities.Category
	for rows.Next() {
		var child entities.Category
		err := rows.Scan(
			&child.ID.Value,
			&child.UserID.Value,
			&child.Name.Value,
			&child.Sequence.Sequence,
			&child.CreatedAt,
			&child.UpdatedAt,
			&child.DeletedAt,
		)
		if err != nil {
			span.RecordError(err)
			r.o11y.Logger().Error(ctx, "query_failed",
				observability.String("operation", "find_by_id"),
				observability.String("layer", "repository"),
				observability.String("entity", "category"),
				observability.String("user_id", userID.String()),
				observability.String("category_id", id.String()),
				observability.Error(err),
			)
			r.fm.RecordRepositoryFailure(ctx, "find_by_id", "category", "infra", time.Since(start))
			return nil, err
		}
		children = append(children, child)
	}

	if err := rows.Err(); err != nil {
		span.RecordError(err)
		r.o11y.Logger().Error(ctx, "query_failed",
			observability.String("operation", "find_by_id"),
			observability.String("layer", "repository"),
			observability.String("entity", "category"),
			observability.String("user_id", userID.String()),
			observability.String("category_id", id.String()),
			observability.Error(err),
		)
		r.fm.RecordRepositoryFailure(ctx, "find_by_id", "category", "infra", time.Since(start))
		return nil, err
	}

	// Adicionar filhos se houver
	if len(children) > 0 {
		category.AddChildrens(children)
	}

	r.o11y.Logger().Debug(ctx, "query_completed",
		observability.String("operation", "find_by_id"),
		observability.String("layer", "repository"),
		observability.String("entity", "category"),
		observability.String("user_id", userID.String()),
		observability.String("category_id", id.String()),
	)
	r.fm.RecordRepositoryQuery(ctx, "find_by_id", "category", time.Since(start))

	return &category, nil
}

func (r *categoryRepository) Save(ctx context.Context, category *entities.Category) error {
	start := time.Now()
	ctx, span := r.o11y.Tracer().Start(ctx, "category_repository.save")
	defer span.End()

	r.o11y.Logger().Debug(ctx, "query_started",
		observability.String("operation", "save"),
		observability.String("layer", "repository"),
		observability.String("entity", "category"),
		observability.String("user_id", category.UserID.String()),
	)

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
					($1, $2, $3, $4, $5, $6, $7, $8)`

	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		span.RecordError(err)
		r.o11y.Logger().Error(ctx, "query_failed",
			observability.String("operation", "save"),
			observability.String("layer", "repository"),
			observability.String("entity", "category"),
			observability.String("user_id", category.UserID.String()),
			observability.Error(err),
		)
		r.fm.RecordRepositoryFailure(ctx, "save", "category", "infra", time.Since(start))
		return err
	}
	defer func() { _ = stmt.Close() }()

	_, err = stmt.ExecContext(
		ctx,
		category.ID.Value,
		category.UserID.Value,
		category.ParentID.SafeUUID(),
		category.Name.Value,
		category.Sequence.Sequence,
		category.CreatedAt.Ptr(),
		category.UpdatedAt.Ptr(),
		category.DeletedAt.Ptr(),
	)
	if err != nil {
		span.RecordError(err)
		r.o11y.Logger().Error(ctx, "query_failed",
			observability.String("operation", "save"),
			observability.String("layer", "repository"),
			observability.String("entity", "category"),
			observability.String("user_id", category.UserID.String()),
			observability.Error(err),
		)
		r.fm.RecordRepositoryFailure(ctx, "save", "category", "infra", time.Since(start))
		return err
	}

	r.o11y.Logger().Debug(ctx, "query_completed",
		observability.String("operation", "save"),
		observability.String("layer", "repository"),
		observability.String("entity", "category"),
		observability.String("user_id", category.UserID.String()),
	)
	r.fm.RecordRepositoryQuery(ctx, "save", "category", time.Since(start))

	return nil
}

func (r *categoryRepository) Update(ctx context.Context, category *entities.Category) error {
	start := time.Now()
	ctx, span := r.o11y.Tracer().Start(ctx, "category_repository.update")
	defer span.End()

	r.o11y.Logger().Debug(ctx, "query_started",
		observability.String("operation", "update"),
		observability.String("layer", "repository"),
		observability.String("entity", "category"),
		observability.String("user_id", category.UserID.String()),
		observability.String("category_id", category.ID.String()),
	)

	query := `update
				categories
			set
				name = $1,
				sequence = $2,
				updated_at = $3,
				parent_id = $4,
				deleted_at = $5
			where
				id = $6
				and user_id = $7`

	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		span.RecordError(err)
		r.o11y.Logger().Error(ctx, "query_failed",
			observability.String("operation", "update"),
			observability.String("layer", "repository"),
			observability.String("entity", "category"),
			observability.String("user_id", category.UserID.String()),
			observability.String("category_id", category.ID.String()),
			observability.Error(err),
		)
		r.fm.RecordRepositoryFailure(ctx, "update", "category", "infra", time.Since(start))
		return err
	}
	defer func() { _ = stmt.Close() }()

	_, err = stmt.ExecContext(
		ctx,
		category.Name.Value,
		category.Sequence.Sequence,
		category.UpdatedAt.Ptr(),
		category.ParentID.SafeUUID(),
		category.DeletedAt.Ptr(),
		category.ID.Value,
		category.UserID.Value,
	)
	if err != nil {
		span.RecordError(err)
		r.o11y.Logger().Error(ctx, "query_failed",
			observability.String("operation", "update"),
			observability.String("layer", "repository"),
			observability.String("entity", "category"),
			observability.String("user_id", category.UserID.String()),
			observability.String("category_id", category.ID.String()),
			observability.Error(err),
		)
		r.fm.RecordRepositoryFailure(ctx, "update", "category", "infra", time.Since(start))
		return err
	}

	r.o11y.Logger().Debug(ctx, "query_completed",
		observability.String("operation", "update"),
		observability.String("layer", "repository"),
		observability.String("entity", "category"),
		observability.String("user_id", category.UserID.String()),
		observability.String("category_id", category.ID.String()),
	)
	r.fm.RecordRepositoryQuery(ctx, "update", "category", time.Since(start))

	return nil
}

func (r *categoryRepository) CheckCycleExists(ctx context.Context, userID, categoryID, parentID vos.UUID) (bool, error) {
	start := time.Now()
	ctx, span := r.o11y.Tracer().Start(ctx, "category_repository.check_cycle_exists")
	defer span.End()

	r.o11y.Logger().Debug(ctx, "query_started",
		observability.String("operation", "check_cycle_exists"),
		observability.String("layer", "repository"),
		observability.String("entity", "category"),
		observability.String("user_id", userID.String()),
		observability.String("category_id", categoryID.String()),
		observability.String("parent_id", parentID.String()),
	)

	query := `
		WITH RECURSIVE category_path AS (
			-- Caso base: começar do parent proposto
			SELECT id, parent_id, 1 as depth
			FROM categories
			WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL

			UNION ALL

			-- Caso recursivo: seguir cadeia de parents
			SELECT c.id, c.parent_id, cp.depth + 1
			FROM categories c
			INNER JOIN category_path cp ON c.id = cp.parent_id
			WHERE c.user_id = $2 AND c.deleted_at IS NULL AND cp.depth < 10
		)
		SELECT EXISTS(SELECT 1 FROM category_path WHERE id = $3) as cycle_exists`

	var cycleExists bool
	err := r.db.QueryRowContext(ctx, query, parentID.String(), userID.String(), categoryID.String()).Scan(&cycleExists)
	if err != nil {
		span.RecordError(err)
		r.o11y.Logger().Error(ctx, "query_failed",
			observability.String("operation", "check_cycle_exists"),
			observability.String("layer", "repository"),
			observability.String("entity", "category"),
			observability.String("user_id", userID.String()),
			observability.String("category_id", categoryID.String()),
			observability.String("parent_id", parentID.String()),
			observability.Error(err),
		)
		r.fm.RecordRepositoryFailure(ctx, "check_cycle_exists", "category", "infra", time.Since(start))
		return false, err
	}

	r.o11y.Logger().Debug(ctx, "query_completed",
		observability.String("operation", "check_cycle_exists"),
		observability.String("layer", "repository"),
		observability.String("entity", "category"),
		observability.String("user_id", userID.String()),
		observability.String("category_id", categoryID.String()),
		observability.String("parent_id", parentID.String()),
	)
	r.fm.RecordRepositoryQuery(ctx, "check_cycle_exists", "category", time.Since(start))

	return cycleExists, nil
}
