package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jailtonjunior94/financial/internal/user/domain/entities"
	"github.com/jailtonjunior94/financial/internal/user/domain/interfaces"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"

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
	r.o11y.Logger().Debug(ctx, "query_started",
		observability.String("operation", "insert"),
		observability.String("layer", "repository"),
		observability.String("entity", "user"),
		observability.String("user_id", user.ID.String()),
	)
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
		r.o11y.Logger().Error(ctx, "query_failed",
			observability.String("operation", "insert"),
			observability.String("layer", "repository"),
			observability.String("entity", "user"),
			observability.String("user_id", user.ID.String()),
			observability.Error(err),
		)
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
		r.o11y.Logger().Error(ctx, "query_failed",
			observability.String("operation", "insert"),
			observability.String("layer", "repository"),
			observability.String("entity", "user"),
			observability.String("user_id", user.ID.String()),
			observability.Error(err),
		)
		r.fm.RecordRepositoryFailure(ctx, "insert", "user", "infra", time.Since(start))
		return nil, fmt.Errorf("executing insert user statement: %w", err)
	}
	r.o11y.Logger().Debug(ctx, "query_completed",
		observability.String("operation", "insert"),
		observability.String("layer", "repository"),
		observability.String("entity", "user"),
		observability.String("user_id", user.ID.String()),
	)
	r.fm.RecordRepositoryQuery(ctx, "insert", "user", time.Since(start))
	return user, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	start := time.Now()
	ctx, span := r.o11y.Tracer().Start(ctx, "user_repository.find_by_email")
	defer span.End()
	r.o11y.Logger().Debug(ctx, "query_started",
		observability.String("operation", "find_by_email"),
		observability.String("layer", "repository"),
		observability.String("entity", "user"),
	)
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
			r.o11y.Logger().Debug(ctx, "query_completed",
				observability.String("operation", "find_by_email"),
				observability.String("layer", "repository"),
				observability.String("entity", "user"),
			)
			r.fm.RecordRepositoryQuery(ctx, "find_by_email", "user", time.Since(start))
			return nil, nil
		}
		span.RecordError(err)
		r.o11y.Logger().Error(ctx, "query_failed",
			observability.String("operation", "find_by_email"),
			observability.String("layer", "repository"),
			observability.String("entity", "user"),
			observability.Error(err),
		)
		r.fm.RecordRepositoryFailure(ctx, "find_by_email", "user", "infra", time.Since(start))
		return nil, fmt.Errorf("querying user by email: %w", err)
	}
	r.o11y.Logger().Debug(ctx, "query_completed",
		observability.String("operation", "find_by_email"),
		observability.String("layer", "repository"),
		observability.String("entity", "user"),
		observability.String("user_id", user.ID.String()),
	)
	r.fm.RecordRepositoryQuery(ctx, "find_by_email", "user", time.Since(start))
	return &user, nil
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*entities.User, error) {
	start := time.Now()
	ctx, span := r.o11y.Tracer().Start(ctx, "user_repository.find_by_id")
	defer span.End()
	r.o11y.Logger().Debug(ctx, "query_started",
		observability.String("operation", "find_by_id"),
		observability.String("layer", "repository"),
		observability.String("entity", "user"),
		observability.String("user_id", id),
	)
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
			r.o11y.Logger().Debug(ctx, "query_completed",
				observability.String("operation", "find_by_id"),
				observability.String("layer", "repository"),
				observability.String("entity", "user"),
				observability.String("user_id", id),
			)
			r.fm.RecordRepositoryQuery(ctx, "find_by_id", "user", time.Since(start))
			return nil, nil
		}
		span.RecordError(err)
		r.o11y.Logger().Error(ctx, "query_failed",
			observability.String("operation", "find_by_id"),
			observability.String("layer", "repository"),
			observability.String("entity", "user"),
			observability.String("user_id", id),
			observability.Error(err),
		)
		r.fm.RecordRepositoryFailure(ctx, "find_by_id", "user", "infra", time.Since(start))
		return nil, fmt.Errorf("querying user by id: %w", err)
	}
	r.o11y.Logger().Debug(ctx, "query_completed",
		observability.String("operation", "find_by_id"),
		observability.String("layer", "repository"),
		observability.String("entity", "user"),
		observability.String("user_id", id),
	)
	r.fm.RecordRepositoryQuery(ctx, "find_by_id", "user", time.Since(start))
	return &user, nil
}
