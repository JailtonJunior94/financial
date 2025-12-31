package usecase

import (
	"context"
	"time"

	"github.com/jailtonjunior94/financial/internal/card/application/dtos"
	"github.com/jailtonjunior94/financial/internal/card/domain/factories"
	"github.com/jailtonjunior94/financial/internal/card/domain/interfaces"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
)

type (
	CreateCardUseCase interface {
		Execute(ctx context.Context, userID string, input *dtos.CardInput) (*dtos.CardOutput, error)
	}

	createCardUseCase struct {
		o11y       observability.Observability
		repository interfaces.CardRepository
	}
)

func NewCreateCardUseCase(
	o11y observability.Observability,
	repository interfaces.CardRepository,
) CreateCardUseCase {
	return &createCardUseCase{
		o11y:       o11y,
		repository: repository,
	}
}

func (u *createCardUseCase) Execute(ctx context.Context, userID string, input *dtos.CardInput) (*dtos.CardOutput, error) {
	ctx, span := u.o11y.Tracer().Start(ctx, "create_card_usecase.execute")
	defer span.End()

	card, err := factories.CreateCard(userID, input.Name, input.DueDay)
	if err != nil {
		span.AddEvent(
			"error creating card entity",
			observability.Field{Key: "user_id", Value: userID},
			observability.Field{Key: "error", Value: err},
		)

		return nil, err
	}

	if err := u.repository.Save(ctx, card); err != nil {
		span.AddEvent(
			"error saving card to repository",
			observability.Field{Key: "user_id", Value: userID},
			observability.Field{Key: "error", Value: err},
		)

		return nil, err
	}

	return &dtos.CardOutput{
		ID:        card.ID.String(),
		Name:      card.Name.String(),
		DueDay:    card.DueDay.Int(),
		CreatedAt: card.CreatedAt.ValueOr(time.Time{}),
	}, nil
}
