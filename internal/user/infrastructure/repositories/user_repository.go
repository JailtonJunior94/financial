package repositories

import (
	"context"
	"database/sql"

	"github.com/jailtonjunior94/financial/internal/user/domain/entities"
	"github.com/jailtonjunior94/financial/internal/user/domain/interfaces"
	"github.com/jailtonjunior94/financial/pkg/o11y"
)

type userRepository struct {
	db   *sql.DB
	o11y o11y.Observability
}

func NewUserRepository(db *sql.DB, o11y o11y.Observability) interfaces.UserRepository {
	return &userRepository{
		db:   db,
		o11y: o11y,
	}
}

func (r *userRepository) Insert(ctx context.Context, user *entities.User) (*entities.User, error) {
	ctx, span := r.o11y.Start(ctx, "user_repository.insert")
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
				(?, ?, ?, ?, ?, ?, ?)`

	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		span.AddAttributes(
			ctx, o11y.Error, "error prepare query insert user",
			o11y.Attributes{Key: "email", Value: user.Email},
			o11y.Attributes{Key: "error", Value: err},
		)
		return nil, err
	}

	_, err = stmt.ExecContext(
		ctx,
		user.ID.String(),
		user.Name.String(),
		user.Email.String(),
		user.Password,
		user.CreatedAt,
		user.UpdatedAt.Time,
		user.DeletedAt.Time,
	)
	if err != nil {
		span.AddAttributes(
			ctx, o11y.Error, "error insert user",
			o11y.Attributes{Key: "email", Value: user.Email},
			o11y.Attributes{Key: "error", Value: err},
		)
		return nil, err
	}
	return user, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	ctx, span := r.o11y.Start(ctx, "user_repository.find_by_email")
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
				email = ?
				and deleted_at is null;`

	var user entities.User
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID.Value,
		&user.Name.Value,
		&user.Email.Value,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt.Time,
		&user.DeletedAt.Time,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			span.AddAttributes(ctx, o11y.Ok, "user not found", o11y.Attributes{Key: "e-mail", Value: email})
			return nil, nil
		}
		span.AddAttributes(ctx, o11y.Error, "error finding user", o11y.Attributes{Key: "e-mail", Value: email})
		return nil, err
	}
	return &user, nil
}
