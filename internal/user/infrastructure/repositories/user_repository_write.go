package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/jailtonjunior94/financial/internal/user/domain/entities"
	"github.com/jailtonjunior94/financial/pkg/pagination"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
)

func (r *userRepository) FindAll(ctx context.Context, limit int, cursor string) ([]*entities.User, *string, error) {
	start := time.Now()
	ctx, span := r.o11y.Tracer().Start(ctx, "user_repository.find_all")
	defer span.End()
	r.o11y.Logger().Debug(ctx, "query_started",
		observability.String("operation", "find_all"),
		observability.String("layer", "repository"),
		observability.String("entity", "user"),
	)
	query, args, err := r.buildFindAllQuery(limit, cursor)
	if err != nil {
		span.RecordError(err)
		r.o11y.Logger().Error(ctx, "query_failed",
			observability.String("operation", "find_all"),
			observability.String("layer", "repository"),
			observability.String("entity", "user"),
			observability.Error(err),
		)
		r.fm.RecordRepositoryFailure(ctx, "find_all", "user", "infra", time.Since(start))
		return nil, nil, fmt.Errorf("decoding cursor: %w", err)
	}
	users, err := r.scanUsers(ctx, query, args)
	if err != nil {
		span.RecordError(err)
		r.o11y.Logger().Error(ctx, "query_failed",
			observability.String("operation", "find_all"),
			observability.String("layer", "repository"),
			observability.String("entity", "user"),
			observability.Error(err),
		)
		r.fm.RecordRepositoryFailure(ctx, "find_all", "user", "infra", time.Since(start))
		return nil, nil, err
	}
	hasNext := len(users) > limit
	if hasNext {
		users = users[:limit]
	}
	nextCursor, err := r.buildNextCursor(hasNext, users)
	if err != nil {
		span.RecordError(err)
		r.o11y.Logger().Error(ctx, "query_failed",
			observability.String("operation", "find_all"),
			observability.String("layer", "repository"),
			observability.String("entity", "user"),
			observability.Error(err),
		)
		r.fm.RecordRepositoryFailure(ctx, "find_all", "user", "infra", time.Since(start))
		return nil, nil, fmt.Errorf("encoding cursor: %w", err)
	}
	r.o11y.Logger().Debug(ctx, "query_completed",
		observability.String("operation", "find_all"),
		observability.String("layer", "repository"),
		observability.String("entity", "user"),
	)
	r.fm.RecordRepositoryQuery(ctx, "find_all", "user", time.Since(start))
	return users, nextCursor, nil
}

func (r *userRepository) buildFindAllQuery(limit int, cursor string) (string, []interface{}, error) {
	if cursor != "" {
		decoded, err := pagination.DecodeCursor(cursor)
		if err != nil {
			return "", nil, err
		}
		cursorName, hasName := decoded.GetString("name")
		cursorID, hasID := decoded.GetString("id")
		if hasName && hasID && cursorName != "" && cursorID != "" {
			query := `select id, name, email, password, created_at, updated_at, deleted_at
					from users
					where deleted_at is null and (name > $1 or (name = $1 and id > $2))
					order by name asc, id asc
					limit $3`
			return query, []interface{}{cursorName, cursorID, limit + 1}, nil
		}
	}
	query := `select id, name, email, password, created_at, updated_at, deleted_at
			from users
			where deleted_at is null
			order by name asc, id asc
			limit $1`
	return query, []interface{}{limit + 1}, nil
}

func (r *userRepository) scanUsers(ctx context.Context, query string, args []interface{}) ([]*entities.User, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying users: %w", err)
	}
	defer func() { _ = rows.Close() }()
	users := make([]*entities.User, 0)
	for rows.Next() {
		var user entities.User
		err := rows.Scan(
			&user.ID.Value,
			&user.Name.Value,
			&user.Email.Value,
			&user.Password,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning user row: %w", err)
		}
		users = append(users, &user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating user rows: %w", err)
	}
	return users, nil
}

func (r *userRepository) buildNextCursor(hasNext bool, users []*entities.User) (*string, error) {
	if !hasNext || len(users) == 0 {
		return nil, nil
	}
	last := users[len(users)-1]
	newCursor := pagination.Cursor{
		Fields: map[string]interface{}{
			"name": last.Name.String(),
			"id":   last.ID.String(),
		},
	}
	encoded, err := pagination.EncodeCursor(newCursor)
	if err != nil {
		return nil, err
	}
	return &encoded, nil
}

func (r *userRepository) Update(ctx context.Context, user *entities.User) (*entities.User, error) {
	start := time.Now()
	ctx, span := r.o11y.Tracer().Start(ctx, "user_repository.update")
	defer span.End()
	r.o11y.Logger().Debug(ctx, "query_started",
		observability.String("operation", "update"),
		observability.String("layer", "repository"),
		observability.String("entity", "user"),
		observability.String("user_id", user.ID.String()),
	)
	query := `update users
			set
				name = $1,
				email = $2,
				password = $3,
				updated_at = $4
			where
				id = $5
				and deleted_at is null`
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		span.RecordError(err)
		r.o11y.Logger().Error(ctx, "query_failed",
			observability.String("operation", "update"),
			observability.String("layer", "repository"),
			observability.String("entity", "user"),
			observability.String("user_id", user.ID.String()),
			observability.Error(err),
		)
		r.fm.RecordRepositoryFailure(ctx, "update", "user", "infra", time.Since(start))
		return nil, fmt.Errorf("preparing update user statement: %w", err)
	}
	defer func() { _ = stmt.Close() }()
	result, err := stmt.ExecContext(
		ctx,
		user.Name.String(),
		user.Email.String(),
		user.Password,
		user.UpdatedAt.Ptr(),
		user.ID.String(),
	)
	if err != nil {
		span.RecordError(err)
		r.o11y.Logger().Error(ctx, "query_failed",
			observability.String("operation", "update"),
			observability.String("layer", "repository"),
			observability.String("entity", "user"),
			observability.String("user_id", user.ID.String()),
			observability.Error(err),
		)
		r.fm.RecordRepositoryFailure(ctx, "update", "user", "infra", time.Since(start))
		return nil, fmt.Errorf("executing update user statement: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		span.RecordError(err)
		r.o11y.Logger().Error(ctx, "query_failed",
			observability.String("operation", "update"),
			observability.String("layer", "repository"),
			observability.String("entity", "user"),
			observability.String("user_id", user.ID.String()),
			observability.Error(err),
		)
		r.fm.RecordRepositoryFailure(ctx, "update", "user", "infra", time.Since(start))
		return nil, fmt.Errorf("checking rows affected: %w", err)
	}
	if rowsAffected == 0 {
		r.fm.RecordRepositoryQuery(ctx, "update", "user", time.Since(start))
		return nil, nil
	}
	r.o11y.Logger().Debug(ctx, "query_completed",
		observability.String("operation", "update"),
		observability.String("layer", "repository"),
		observability.String("entity", "user"),
		observability.String("user_id", user.ID.String()),
	)
	r.fm.RecordRepositoryQuery(ctx, "update", "user", time.Since(start))
	return user, nil
}

func (r *userRepository) SoftDelete(ctx context.Context, id string) error {
	start := time.Now()
	ctx, span := r.o11y.Tracer().Start(ctx, "user_repository.soft_delete")
	defer span.End()
	r.o11y.Logger().Debug(ctx, "query_started",
		observability.String("operation", "soft_delete"),
		observability.String("layer", "repository"),
		observability.String("entity", "user"),
		observability.String("user_id", id),
	)
	query := `update users
			set
				deleted_at = now(),
				updated_at = now()
			where
				id = $1
				and deleted_at is null`
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		span.RecordError(err)
		r.o11y.Logger().Error(ctx, "query_failed",
			observability.String("operation", "soft_delete"),
			observability.String("layer", "repository"),
			observability.String("entity", "user"),
			observability.String("user_id", id),
			observability.Error(err),
		)
		r.fm.RecordRepositoryFailure(ctx, "soft_delete", "user", "infra", time.Since(start))
		return fmt.Errorf("preparing soft delete user statement: %w", err)
	}
	defer func() { _ = stmt.Close() }()
	result, err := stmt.ExecContext(ctx, id)
	if err != nil {
		span.RecordError(err)
		r.o11y.Logger().Error(ctx, "query_failed",
			observability.String("operation", "soft_delete"),
			observability.String("layer", "repository"),
			observability.String("entity", "user"),
			observability.String("user_id", id),
			observability.Error(err),
		)
		r.fm.RecordRepositoryFailure(ctx, "soft_delete", "user", "infra", time.Since(start))
		return fmt.Errorf("executing soft delete user: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		span.RecordError(err)
		r.o11y.Logger().Error(ctx, "query_failed",
			observability.String("operation", "soft_delete"),
			observability.String("layer", "repository"),
			observability.String("entity", "user"),
			observability.String("user_id", id),
			observability.Error(err),
		)
		r.fm.RecordRepositoryFailure(ctx, "soft_delete", "user", "infra", time.Since(start))
		return fmt.Errorf("checking rows affected: %w", err)
	}
	if rowsAffected == 0 {
		r.fm.RecordRepositoryQuery(ctx, "soft_delete", "user", time.Since(start))
		return nil
	}
	r.o11y.Logger().Debug(ctx, "query_completed",
		observability.String("operation", "soft_delete"),
		observability.String("layer", "repository"),
		observability.String("entity", "user"),
		observability.String("user_id", id),
	)
	r.fm.RecordRepositoryQuery(ctx, "soft_delete", "user", time.Since(start))
	return nil
}
