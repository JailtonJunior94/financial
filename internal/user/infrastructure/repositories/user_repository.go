package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jailtonjunior94/financial/internal/user/domain/entities"
	"github.com/jailtonjunior94/financial/internal/user/domain/interfaces"
	customerrors "github.com/jailtonjunior94/financial/pkg/custom_errors"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"
	"github.com/jailtonjunior94/financial/pkg/pagination"

	"github.com/JailtonJunior94/devkit-go/pkg/database"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
)

type userRepository struct {
	db   database.DBTX
	o11y observability.Observability
	fm   *metrics.FinancialMetrics
}

func NewUserRepository(db database.DBTX, o11y observability.Observability, fm *metrics.FinancialMetrics) interfaces.UserRepository {
	return &userRepository{
		db:   db,
		o11y: o11y,
		fm:   fm,
	}
}

func (r *userRepository) Insert(ctx context.Context, user *entities.User) (*entities.User, error) {
	start := time.Now()
	ctx, span := r.o11y.Tracer().Start(ctx, "user_repository.insert")
	defer span.End()

	span.AddEvent("inserting user", observability.String("id", user.ID.String()))

	query := `insert into
				users (
					id,
					name,
					email,
					password,
					created_at,
					updated_at,
					deleted_at
				)
				values
				($1, $2, $3, $4, $5, $6, $7)`

	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		span.RecordError(err)
		r.fm.RecordRepositoryFailure(ctx, "insert", "user", "infra", time.Since(start))
		return nil, fmt.Errorf("preparing insert user statement: %w", err)
	}
	defer func() { _ = stmt.Close() }()

	_, err = stmt.ExecContext(
		ctx,
		user.ID.String(),
		user.Name.String(),
		user.Email.String(),
		user.Password,
		user.CreatedAt,
		user.UpdatedAt.Ptr(),
		user.DeletedAt.Ptr(),
	)
	if err != nil {
		span.RecordError(err)
		r.fm.RecordRepositoryFailure(ctx, "insert", "user", "infra", time.Since(start))
		return nil, fmt.Errorf("executing insert user statement: %w", err)
	}

	r.fm.RecordRepositoryQuery(ctx, "insert", "user", time.Since(start))
	return user, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	start := time.Now()
	ctx, span := r.o11y.Tracer().Start(ctx, "user_repository.find_by_email")
	defer span.End()

	span.AddEvent("finding user by email", observability.String("email", email))

	query := `select
				id,
				name,
				email,
				password,
				created_at,
				updated_at,
				deleted_at
			from
				users
			where
				email = $1
				and deleted_at is null;`

	var user entities.User
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID.Value,
		&user.Name.Value,
		&user.Email.Value,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.fm.RecordRepositoryQuery(ctx, "find_by_email", "user", time.Since(start))
			return nil, nil
		}
		span.RecordError(err)
		r.fm.RecordRepositoryFailure(ctx, "find_by_email", "user", "infra", time.Since(start))
		return nil, fmt.Errorf("querying user by email: %w", err)
	}

	r.fm.RecordRepositoryQuery(ctx, "find_by_email", "user", time.Since(start))
	return &user, nil
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*entities.User, error) {
	start := time.Now()
	ctx, span := r.o11y.Tracer().Start(ctx, "user_repository.find_by_id")
	defer span.End()

	span.AddEvent("finding user by id", observability.String("id", id))

	query := `select
				id,
				name,
				email,
				password,
				created_at,
				updated_at,
				deleted_at
			from
				users
			where
				id = $1
				and deleted_at is null;`

	var user entities.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID.Value,
		&user.Name.Value,
		&user.Email.Value,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.fm.RecordRepositoryQuery(ctx, "find_by_id", "user", time.Since(start))
			return nil, nil
		}
		span.RecordError(err)
		r.fm.RecordRepositoryFailure(ctx, "find_by_id", "user", "infra", time.Since(start))
		return nil, fmt.Errorf("querying user by id: %w", err)
	}

	r.fm.RecordRepositoryQuery(ctx, "find_by_id", "user", time.Since(start))
	return &user, nil
}

func (r *userRepository) FindAll(ctx context.Context, limit int, cursor string) ([]*entities.User, *string, error) {
	start := time.Now()
	ctx, span := r.o11y.Tracer().Start(ctx, "user_repository.find_all")
	defer span.End()

	var (
		query string
		args  []interface{}
	)

	if cursor != "" {
		decoded, err := pagination.DecodeCursor(cursor)
		if err != nil {
			span.RecordError(err)
			r.fm.RecordRepositoryFailure(ctx, "find_all", "user", "infra", time.Since(start))
			return nil, nil, fmt.Errorf("decoding cursor: %w", err)
		}

		cursorName, hasName := decoded.GetString("name")
		cursorID, hasID := decoded.GetString("id")

		if hasName && hasID && cursorName != "" && cursorID != "" {
			query = `select
						id,
						name,
						email,
						password,
						created_at,
						updated_at,
						deleted_at
					from
						users
					where
						deleted_at is null
						and (name > $1 or (name = $1 and id > $2))
					order by
						name asc, id asc
					limit $3`
			args = []interface{}{cursorName, cursorID, limit + 1}
		} else {
			cursor = ""
		}
	}

	if cursor == "" {
		query = `select
					id,
					name,
					email,
					password,
					created_at,
					updated_at,
					deleted_at
				from
					users
				where
					deleted_at is null
				order by
					name asc, id asc
				limit $1`
		args = []interface{}{limit + 1}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		span.RecordError(err)
		r.fm.RecordRepositoryFailure(ctx, "find_all", "user", "infra", time.Since(start))
		return nil, nil, fmt.Errorf("querying users: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			span.RecordError(closeErr)
		}
	}()

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
			span.RecordError(err)
			r.fm.RecordRepositoryFailure(ctx, "find_all", "user", "infra", time.Since(start))
			return nil, nil, fmt.Errorf("scanning user row: %w", err)
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		span.RecordError(err)
		r.fm.RecordRepositoryFailure(ctx, "find_all", "user", "infra", time.Since(start))
		return nil, nil, fmt.Errorf("iterating user rows: %w", err)
	}

	hasNext := len(users) > limit
	if hasNext {
		users = users[:limit]
	}

	var nextCursor *string
	if hasNext && len(users) > 0 {
		last := users[len(users)-1]
		newCursor := pagination.Cursor{
			Fields: map[string]interface{}{
				"name": last.Name.String(),
				"id":   last.ID.String(),
			},
		}
		encoded, err := pagination.EncodeCursor(newCursor)
		if err != nil {
			span.RecordError(err)
			r.fm.RecordRepositoryFailure(ctx, "find_all", "user", "infra", time.Since(start))
			return nil, nil, fmt.Errorf("encoding cursor: %w", err)
		}
		nextCursor = &encoded
	}

	r.fm.RecordRepositoryQuery(ctx, "find_all", "user", time.Since(start))
	return users, nextCursor, nil
}

func (r *userRepository) Update(ctx context.Context, user *entities.User) (*entities.User, error) {
	start := time.Now()
	ctx, span := r.o11y.Tracer().Start(ctx, "user_repository.update")
	defer span.End()

	span.AddEvent("updating user", observability.String("id", user.ID.String()))

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
		r.fm.RecordRepositoryFailure(ctx, "update", "user", "infra", time.Since(start))
		return nil, fmt.Errorf("executing update user statement: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		span.RecordError(err)
		r.fm.RecordRepositoryFailure(ctx, "update", "user", "infra", time.Since(start))
		return nil, fmt.Errorf("checking rows affected: %w", err)
	}

	if rowsAffected == 0 {
		r.fm.RecordRepositoryQuery(ctx, "update", "user", time.Since(start))
		return nil, customerrors.ErrUserNotFound
	}

	r.fm.RecordRepositoryQuery(ctx, "update", "user", time.Since(start))
	return user, nil
}

func (r *userRepository) SoftDelete(ctx context.Context, id string) error {
	start := time.Now()
	ctx, span := r.o11y.Tracer().Start(ctx, "user_repository.soft_delete")
	defer span.End()

	span.AddEvent("soft deleting user", observability.String("id", id))

	query := `update users
			set
				deleted_at = now(),
				updated_at = now()
			where
				id = $1
				and deleted_at is null`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		span.RecordError(err)
		r.fm.RecordRepositoryFailure(ctx, "soft_delete", "user", "infra", time.Since(start))
		return fmt.Errorf("executing soft delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		span.RecordError(err)
		r.fm.RecordRepositoryFailure(ctx, "soft_delete", "user", "infra", time.Since(start))
		return fmt.Errorf("checking rows affected: %w", err)
	}

	if rowsAffected == 0 {
		r.fm.RecordRepositoryQuery(ctx, "soft_delete", "user", time.Since(start))
		return customerrors.ErrUserNotFound
	}

	r.fm.RecordRepositoryQuery(ctx, "soft_delete", "user", time.Since(start))
	return nil
}
