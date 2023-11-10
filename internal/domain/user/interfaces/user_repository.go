package interfaces

import "github.com/jailtonjunior94/financial/internal/domain/user/entity"

type UserRepository interface {
	Create(u *entity.User) (*entity.User, error)
}
