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
	stmt, err := r.Db.Prepare("insert into users (id, name, email, password, created_at, updated_at, active) values (?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return nil, err
	}

	_, err = stmt.Exec(u.ID, u.Name, u.Email, u.Password, u.CreatedAt, u.UpdatedAt, u.Active)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *userRepository) FindByEmail(email string) (*entity.User, error) {
	row := r.Db.QueryRow("select * from users where email = ? and active = true", email)
	var user entity.User
	if err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt, &user.Active); err != nil {
		return nil, err
	}
	return &user, nil
}
