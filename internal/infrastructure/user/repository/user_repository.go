package repository

import (
	"database/sql"

	"github.com/jailtonjunior94/financial/internal/domain/user/entity"
	"github.com/jailtonjunior94/financial/internal/domain/user/interfaces"
)

type userRepository struct {
	Db *sql.DB
}

func NewUserRepository(db *sql.DB) interfaces.UserRepository {
	return &userRepository{Db: db}
}

func (r *userRepository) Create(u *entity.User) (*entity.User, error) {
	stmt, err := r.Db.Prepare("INSERT INTO users (id, name, email, password, created_at, updated_at, active) VALUES (?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return nil, err
	}

	_, err = stmt.Exec(u.ID, u.Name, u.Email, u.Password, u.CreatedAt, u.UpdatedAt, u.Active)
	if err != nil {
		return nil, err
	}
	return u, nil
}
