package repositories

import (
	"context"
	"database/sql"
	"time"

	"github.com/jailtonjunior94/financial/internal/user/domain/entities"
	"github.com/jailtonjunior94/financial/internal/user/domain/interfaces"
	customErrors "github.com/jailtonjunior94/financial/pkg/custom_errors"
	"github.com/jailtonjunior94/financial/pkg/database"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
)

type userRepository struct {
	db   database.DBExecutor
	o11y observability.Observability
}

func NewUserRepository(db database.DBExecutor, o11y observability.Observability) interfaces.UserRepository {
	return &userRepository{
		db:   db,
		o11y: o11y,
	}
}

func (r *userRepository) Insert(ctx context.Context, user *entities.User) (*entities.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	ctx, span := r.o11y.Tracer().Start(ctx, "user_repository.insert")
	defer span.End()

	// Verificar se email já existe
	existing, err := r.FindByEmail(ctx, user.Email.String())
	if err != nil {
		return nil, err
	}
	if existing != nil {
		span.AddEvent("email already exists", observability.Field{Key: "email", Value: user.Email})
		r.o11y.Logger().Error(ctx, "email already exists", observability.Error(customErrors.ErrEmailAlreadyExists))
		return nil, customErrors.ErrEmailAlreadyExists
	}

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
		span.AddEvent(
			"error preparing insert user query",
			observability.Field{Key: "email", Value: user.Email},
			observability.Field{Key: "error", Value: err},
		)
		r.o11y.Logger().Error(ctx, "error preparing insert user query", observability.Error(err), observability.String("email", user.Email.String()))
		return nil, err
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
		span.AddEvent(
			"error inserting user",
			observability.Field{Key: "email", Value: user.Email},
			observability.Field{Key: "error", Value: err},
		)
		r.o11y.Logger().Error(ctx, "error inserting user", observability.Error(err), observability.String("email", user.Email.String()))
		return nil, err
	}
	return user, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	ctx, span := r.o11y.Tracer().Start(ctx, "user_repository.find_by_email")
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
		span.AddEvent("error finding user", observability.Field{Key: "e-mail", Value: email}, observability.Field{Key: "error", Value: err})
		r.o11y.Logger().Error(ctx, "error finding user", observability.Error(err), observability.String("e-mail", email))
		return nil, err
	}
	return &user, nil
}
