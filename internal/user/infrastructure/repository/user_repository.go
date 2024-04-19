package repository

import (
	"context"
	"database/sql"

	"github.com/jailtonjunior94/financial/internal/user/domain/entities"
	"github.com/jailtonjunior94/financial/internal/user/domain/interfaces"
	"github.com/jailtonjunior94/financial/pkg/observability"
)

type userRepository struct {
	db            *sql.DB
	observability observability.Observability
}

func NewUserRepository(
	db *sql.DB,
	observability observability.Observability,
) interfaces.UserRepository {
	return &userRepository{
		db:            db,
		observability: observability,
	}
}

func (r *userRepository) Insert(ctx context.Context, user *entities.User) (*entities.User, error) {
	ctx, span := r.observability.Tracer().Start(ctx, "user_repository.Create")
	defer span.End()

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
		return nil, err
	}

	_, err = stmt.ExecContext(
		ctx,
		user.ID.Value,
		user.Name.String(),
		user.Email.String(),
		user.Password,
		user.CreatedAt,
		user.UpdatedAt,
		user.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	ctx, span := r.observability.Tracer().Start(ctx, "user_repository.FindByEmail")
	defer span.End()

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
		return nil, err
	}
	return &user, nil
}
