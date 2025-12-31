package interfaces

import (
	"context"

	"github.com/jailtonjunior94/financial/internal/card/domain/entities"

	"github.com/JailtonJunior94/devkit-go/pkg/vos"
)

type CardRepository interface {
	List(ctx context.Context, userID vos.UUID) ([]*entities.Card, error)
	FindByID(ctx context.Context, userID, id vos.UUID) (*entities.Card, error)
	Save(ctx context.Context, card *entities.Card) error
	Update(ctx context.Context, card *entities.Card) error
}
