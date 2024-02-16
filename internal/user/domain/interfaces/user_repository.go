package interfaces

import (
	"context"

	"github.com/jailtonjunior94/financial/internal/user/domain/entity"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) (*entity.User, error)
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
}
