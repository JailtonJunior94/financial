package repositories

import (
	"context"
	"time"

	"github.com/jailtonjunior94/financial/internal/category/domain/entities"
	"github.com/jailtonjunior94/financial/internal/category/domain/interfaces"
	"github.com/jailtonjunior94/financial/pkg/database"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
)

type categoryRepository struct {
	db   database.DBExecutor
	o11y observability.Observability
}

func NewCategoryRepository(db database.DBExecutor, o11y observability.Observability) interfaces.CategoryRepository {
	return &categoryRepository{
		db:   db,
		o11y: o11y,
	}
}

func (r *categoryRepository) List(ctx context.Context, userID vos.UUID) ([]*entities.Category, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

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
		span.AddEvent("error finding categories", observability.Field{Key: "user_id", Value: userID.String()}, observability.Field{Key: "error", Value: err})
		r.o11y.Logger().Error(ctx, "error finding categories", observability.Error(err), observability.String("user_id", userID.String()))
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			span.AddEvent("error closing rows", observability.Field{Key: "user_id", Value: userID.String()}, observability.Field{Key: "error", Value: closeErr})
			r.o11y.Logger().Error(ctx, "error closing rows", observability.Error(closeErr), observability.String("user_id", userID.String()))
		}
	}()

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
			span.AddEvent("error scanning categories", observability.Field{Key: "user_id", Value: userID.String()}, observability.Field{Key: "error", Value: err})
			r.o11y.Logger().Error(ctx, "error scanning categories", observability.Error(err), observability.String("user_id", userID.String()))
			return nil, err
		}
		categories = append(categories, &category)
	}
	return categories, nil
}

func (r *categoryRepository) FindByID(ctx context.Context, userID, id vos.UUID) (*entities.Category, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	ctx, span := r.o11y.Tracer().Start(ctx, "category_repository.find_by_id")
	defer span.End()

	query := `select
					c.id,
					c.user_id,
					c.parent_id,
					c.name,
					c.sequence,
					c.created_at,
					c.updated_at,
					c.deleted_at,
					c2.id,
					c2.user_id,
					c2.name,
					c2.sequence,
					c2.created_at,
					c2.updated_at,
					c2.deleted_at
				from
					categories c
					left join categories c2 on c.id = c2.parent_id AND c2.deleted_at IS NULL
				where
					c.user_id = $1
					and c.deleted_at is null
					and c.id = $2
				order by
					c.sequence;`

	rows, err := r.db.QueryContext(ctx, query, userID.String(), id.String())
	if err != nil {
		span.AddEvent("error finding category", observability.Field{Key: "user_id", Value: userID.String()}, observability.Field{Key: "error", Value: err})
		r.o11y.Logger().Error(ctx, "error finding category", observability.Error(err), observability.String("user_id", userID.String()))
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			span.AddEvent("error closing rows", observability.Field{Key: "user_id", Value: userID.String()}, observability.Field{Key: "error", Value: closeErr})
			r.o11y.Logger().Error(ctx, "error closing rows", observability.Error(closeErr), observability.String("user_id", userID.String()))
		}
	}()

	var category entities.Category
	var subCategory entities.Category
	var subCategories = make(map[vos.UUID][]entities.Category)

	hasRows := false
	for rows.Next() {
		hasRows = true
		err := rows.Scan(
			&category.ID.Value,
			&category.UserID.Value,
			&category.ParentID,
			&category.Name.Value,
			&category.Sequence.Sequence,
			&category.CreatedAt,
			&category.UpdatedAt,
			&category.DeletedAt,
			&subCategory.ID.Value,
			&subCategory.UserID.Value,
			&subCategory.Name.Value,
			&subCategory.Sequence.Sequence,
			&subCategory.CreatedAt,
			&subCategory.UpdatedAt,
			&subCategory.DeletedAt,
		)

		if err != nil {
			span.AddEvent("error scanning category", observability.Field{Key: "user_id", Value: userID.String()}, observability.Field{Key: "error", Value: err})
			r.o11y.Logger().Error(ctx, "error scanning category", observability.Error(err), observability.String("user_id", userID.String()))
			return nil, err
		}

		if subCategory.ID.IsEmpty() {
			continue
		}

		if _, ok := subCategories[category.ID]; !ok {
			subCategories[category.ID] = []entities.Category{subCategory}
			continue
		}
		subCategories[category.ID] = append(subCategories[category.ID], subCategory)
	}

	if !hasRows {
		return nil, nil
	}

	category.AddChildrens(subCategories[category.ID])
	return &category, nil
}

func (r *categoryRepository) Save(ctx context.Context, category *entities.Category) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

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
		r.o11y.Logger().Error(ctx, "error preparing insert category", observability.Error(err), observability.String("user_id", category.UserID.String()))
		return err
	}
	defer stmt.Close()

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
		r.o11y.Logger().Error(ctx, "error inserting category", observability.Error(err), observability.String("user_id", category.UserID.String()))
		return err
	}
	return nil
}

func (r *categoryRepository) Update(ctx context.Context, category *entities.Category) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

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
		r.o11y.Logger().Error(ctx, "error preparing update category", observability.Error(err), observability.String("user_id", category.UserID.String()))
		return err
	}
	defer stmt.Close()

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
		r.o11y.Logger().Error(ctx, "error updating category", observability.Error(err), observability.String("user_id", category.UserID.String()))
		return err
	}

	return nil
}

func (r *categoryRepository) CheckCycleExists(ctx context.Context, userID, categoryID, parentID vos.UUID) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

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
