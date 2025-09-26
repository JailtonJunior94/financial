package entities

import (
	"time"

	"github.com/jailtonjunior94/financial/internal/user/domain/vos"
	financialErrors "github.com/jailtonjunior94/financial/pkg/custom_errors"

	"github.com/JailtonJunior94/devkit-go/pkg/entity"
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
		return financialErrors.ErrPasswordIsRequired
	}
	u.Password = password
	return nil
}
