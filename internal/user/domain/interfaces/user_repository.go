package interfaces

import (
	"context"

	"github.com/jailtonjunior94/financial/internal/user/domain/entities"
)

type UserRepository interface {
	Insert(ctx context.Context, user *entities.User) (*entities.User, error)
	FindByEmail(ctx context.Context, email string) (*entities.User, error)
	FindByID(ctx context.Context, id string) (*entities.User, error)
	FindAll(ctx context.Context, limit int, cursor string) ([]*entities.User, *string, error)
	Update(ctx context.Context, user *entities.User) (*entities.User, error)
	SoftDelete(ctx context.Context, id string) error
}
