package repository

import (
	"context"
	"database/sql"

	"github.com/jailtonjunior94/financial/internal/user/domain/entity"
	"github.com/jailtonjunior94/financial/internal/user/domain/interfaces"
)

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) interfaces.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *entity.User) (*entity.User, error) {
	stmt, err := r.db.Prepare("insert into users (id, name, email, password, created_at, updated_at, active) values (?, ?, ?, ?, ?, ?, ?)")
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
	row := r.db.QueryRowContext(ctx, "select * from users where email = ? and active = true", email)
	var user entity.User
	if err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt, &user.Active); err != nil {
		return nil, err
	}
	return &user, nil
}
