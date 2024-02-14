package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNameIsRequired     = errors.New("name is required")
	ErrEmailIsRequired    = errors.New("email is required")
	ErrPasswordIsRequired = errors.New("password is required")
)

type User struct {
	ID        string
	Name      string
	Email     string
	Password  string
	CreatedAt time.Time
	UpdatedAt *time.Time
	Active    bool
}

func NewUser(name, email string) (*User, error) {
	user := &User{
		ID:        uuid.New().String(),
		Name:      name,
		Email:     email,
		CreatedAt: time.Now(),
		Active:    true,
	}

	if err := user.IsValid(); err != nil {
		return nil, err
	}
	return user, nil
}

func (u *User) IsValid() error {
	if u.Name == "" {
		return ErrNameIsRequired
	}
	if u.Email == "" {
		return ErrEmailIsRequired
	}
	return nil
}

func (u *User) SetPassword(password string) error {
	if password == "" {
		return ErrPasswordIsRequired
	}
	u.Password = password
	return nil
}
