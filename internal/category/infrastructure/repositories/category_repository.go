package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jailtonjunior94/financial/internal/category/domain/entities"
	"github.com/jailtonjunior94/financial/internal/category/domain/interfaces"

	"github.com/JailtonJunior94/devkit-go/pkg/database"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
)

type categoryRepository struct {
	db   database.DBTX
	o11y observability.Observability
}

func NewCategoryRepository(db database.DBTX, o11y observability.Observability) interfaces.CategoryRepository {
	return &categoryRepository{
		db:   db,
		o11y: o11y,
	}
}

func (r *categoryRepository) List(ctx context.Context, userID vos.UUID) ([]*entities.Category, error) {
	// Removido WithTimeout: confiar no timeout do contexto pai (HTTP request ou transação)
	ctx, span := r.o11y.Tracer().Start(ctx, "category_repository.list")
	defer span.End()

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
			return nil, err
		}
		categories = append(categories, &category)
	}

	if err := rows.Err(); err != nil {
		span.RecordError(err)
		return nil, err
	}

	return categories, nil
}

// ListPaginated lista categorias de um usuário com paginação cursor-based.
func (r *categoryRepository) ListPaginated(ctx context.Context, params interfaces.ListCategoriesParams) ([]*entities.Category, error) {
	ctx, span := r.o11y.Tracer().Start(ctx, "category_repository.list_paginated")
	defer span.End()

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
			return nil, err
		}
		categories = append(categories, &category)
	}

	if err := rows.Err(); err != nil {
		span.RecordError(err)
		return nil, err
	}

	return categories, nil
}

func (r *categoryRepository) FindByID(ctx context.Context, userID, id vos.UUID) (*entities.Category, error) {
	ctx, span := r.o11y.Tracer().Start(ctx, "category_repository.find_by_id")
	defer span.End()

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
	err := r.db.QueryRowContext(ctx, queryCategory, userID.String(), id.String()).Scan(
		&category.ID.Value,
		&category.UserID.Value,
		&category.ParentID,
		&category.Name.Value,
		&category.Sequence.Sequence,
		&category.CreatedAt,
		&category.UpdatedAt,
		&category.DeletedAt,
	)

	if err == sql.ErrNoRows {
		// Categoria não encontrada
		return nil, nil
	}
	if err != nil {
		span.RecordError(err)
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
			return nil, err
		}
		children = append(children, child)
	}

	if err := rows.Err(); err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Adicionar filhos se houver
	if len(children) > 0 {
		category.AddChildrens(children)
	}

	return &category, nil
}

func (r *categoryRepository) Save(ctx context.Context, category *entities.Category) error {
	// Removido WithTimeout: confiar no timeout do contexto pai (HTTP request ou transação)
	ctx, span := r.o11y.Tracer().Start(ctx, "category_repository.save")
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
					($1, $2, $3, $4, $5, $6, $7, $8)`

	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		span.AddEvent(
			"error preparing insert category",
			observability.Field{Key: "user_id", Value: category.UserID},
			observability.Field{Key: "error", Value: err},
		)

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
		span.AddEvent(
			"error inserting category",
			observability.Field{Key: "user_id", Value: category.UserID},
			observability.Field{Key: "error", Value: err},
		)

		return err
	}
	return nil
}

func (r *categoryRepository) Update(ctx context.Context, category *entities.Category) error {
	// Removido WithTimeout: confiar no timeout do contexto pai (HTTP request ou transação)
	ctx, span := r.o11y.Tracer().Start(ctx, "category_repository.update")
	defer span.End()

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
		span.AddEvent(
			"error preparing update category",
			observability.Field{Key: "user_id", Value: category.UserID},
			observability.Field{Key: "error", Value: err},
		)

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
		span.AddEvent(
			"error updating category",
			observability.Field{Key: "user_id", Value: category.UserID},
			observability.Field{Key: "error", Value: err},
		)

		return err
	}

	return nil
}

func (r *categoryRepository) CheckCycleExists(ctx context.Context, userID, categoryID, parentID vos.UUID) (bool, error) {
	// Removido WithTimeout: confiar no timeout do contexto pai (HTTP request ou transação)
	ctx, span := r.o11y.Tracer().Start(ctx, "category_repository.check_cycle_exists")
	defer span.End()

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
		span.AddEvent(
			"error checking category cycle",
			observability.Field{Key: "user_id", Value: userID.String()},
			observability.Field{Key: "category_id", Value: categoryID.String()},
			observability.Field{Key: "parent_id", Value: parentID.String()},
			observability.Field{Key: "error", Value: err},
		)
		r.o11y.Logger().Error(ctx, "error checking category cycle",
			observability.Error(err),
			observability.String("user_id", userID.String()),
			observability.String("category_id", categoryID.String()),
			observability.String("parent_id", parentID.String()))
		return false, err
	}

	return cycleExists, nil
}
