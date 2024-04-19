package interfaces

import (
	"context"

	"github.com/jailtonjunior94/financial/internal/user/domain/entities"
)

type UserRepository interface {
	Insert(ctx context.Context, user *entities.User) (*entities.User, error)
	FindByEmail(ctx context.Context, email string) (*entities.User, error)
}
