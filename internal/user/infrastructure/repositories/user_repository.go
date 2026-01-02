package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jailtonjunior94/financial/internal/user/domain/entities"
	"github.com/jailtonjunior94/financial/internal/user/domain/interfaces"

	"github.com/JailtonJunior94/devkit-go/pkg/database"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
)

type userRepository struct {
	db   database.DBTX
	o11y observability.Observability
}

func NewUserRepository(db database.DBTX, o11y observability.Observability) interfaces.UserRepository {
	return &userRepository{
		db:   db,
		o11y: o11y,
	}
}

func (r *userRepository) Insert(ctx context.Context, user *entities.User) (*entities.User, error) {
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
		return nil, fmt.Errorf("preparing insert user statement: %w", err)
	}
	defer stmt.Close()

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
		return nil, fmt.Errorf("executing insert user statement: %w", err)
	}

	return user, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
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
		if err == sql.ErrNoRows {
			return nil, nil
		}
		span.RecordError(err)
		return nil, fmt.Errorf("querying user by email: %w", err)
	}

	return &user, nil
}
