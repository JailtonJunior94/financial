package entities

import (
	"errors"
	"time"

	"github.com/jailtonjunior94/financial/internal/user/domain/vos"

	"github.com/JailtonJunior94/devkit-go/pkg/entity"
)

var (
	ErrPasswordIsRequired = errors.New("password is required")
)

type User struct {
	entity.Base
	Name     vos.UserName
	Email    vos.Email
	Password string
}

func NewUser(name vos.UserName, email vos.Email) (*User, error) {
	user := &User{
		Name:  name,
		Email: email,
		Base: entity.Base{
			CreatedAt: time.Now(),
		},
	}
	return user, nil
}

func (u *User) SetPassword(password string) error {
	if password == "" {
		return ErrPasswordIsRequired
	}
	u.Password = password
	return nil
}
