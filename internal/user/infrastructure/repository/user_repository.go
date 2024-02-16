package repository

import (
	"context"
	"database/sql"

	"github.com/jailtonjunior94/financial/internal/user/domain/entity"
	"github.com/jailtonjunior94/financial/internal/user/domain/interfaces"
	"github.com/jailtonjunior94/financial/pkg/observability"
)

type userRepository struct {
	db            *sql.DB
	observability observability.Observability
}

func NewUserRepository(db *sql.DB, observability observability.Observability) interfaces.UserRepository {
	return &userRepository{db: db, observability: observability}
}

func (r *userRepository) Create(ctx context.Context, user *entity.User) (*entity.User, error) {
	ctx, span := r.observability.Tracer().Start(ctx, "user_repository.Create")
	defer span.End()

	stmt, err := r.db.PrepareContext(ctx, "insert into users (id, name, email, password, created_at, updated_at, active) values (?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return nil, err
	}

	_, err = stmt.ExecContext(ctx, user.ID, user.Name, user.Email, user.Password, user.CreatedAt, user.UpdatedAt, user.Active)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	ctx, span := r.observability.Tracer().Start(ctx, "user_repository.FindByEmail")
	defer span.End()

	row := r.db.QueryRowContext(ctx, "select * from users where email = ? and active = true", email)
	var user entity.User
	if err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt, &user.Active); err != nil {
		return nil, err
	}
	return &user, nil
}
