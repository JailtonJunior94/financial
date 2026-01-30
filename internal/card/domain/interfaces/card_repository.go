package interfaces

import (
	"context"

	"github.com/jailtonjunior94/financial/internal/card/domain/entities"
	"github.com/jailtonjunior94/financial/pkg/pagination"

	"github.com/JailtonJunior94/devkit-go/pkg/vos"
)

// ListCardsParams representa os parâmetros para paginação de cards.
type ListCardsParams struct {
	UserID vos.UUID
	Limit  int
	Cursor pagination.Cursor
}

type CardRepository interface {
	List(ctx context.Context, userID vos.UUID) ([]*entities.Card, error)
	ListPaginated(ctx context.Context, params ListCardsParams) ([]*entities.Card, error)
	FindByID(ctx context.Context, userID, id vos.UUID) (*entities.Card, error)
	Save(ctx context.Context, card *entities.Card) error
	Update(ctx context.Context, card *entities.Card) error
}
