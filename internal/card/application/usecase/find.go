package usecase

import (
	"context"
	"time"

	"github.com/jailtonjunior94/financial/internal/card/application/dtos"
	"github.com/jailtonjunior94/financial/internal/card/domain/entities"
	"github.com/jailtonjunior94/financial/internal/card/domain/interfaces"

	"github.com/JailtonJunior94/devkit-go/pkg/linq"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
)

type (
	FindCardUseCase interface {
		Execute(ctx context.Context, userID string) ([]*dtos.CardOutput, error)
	}

	findCardUseCase struct {
		o11y       observability.Observability
		repository interfaces.CardRepository
	}
)

func NewFindCardUseCase(
	o11y observability.Observability,
	repository interfaces.CardRepository,
) FindCardUseCase {
	return &findCardUseCase{
		o11y:       o11y,
		repository: repository,
	}
}

func (u *findCardUseCase) Execute(ctx context.Context, userID string) ([]*dtos.CardOutput, error) {
	ctx, span := u.o11y.Tracer().Start(ctx, "find_card_usecase.execute")
	defer span.End()

	user, err := vos.NewUUIDFromString(userID)
	if err != nil {
		span.AddEvent(
			"error parsing user id",
			observability.Field{Key: "user_id", Value: userID},
			observability.Field{Key: "error", Value: err},
		)

		return nil, err
	}

	cards, err := u.repository.List(ctx, user)
	if err != nil {
		span.AddEvent(
			"error listing cards from repository",
			observability.Field{Key: "user_id", Value: userID},
			observability.Field{Key: "error", Value: err},
		)

		return nil, err
	}

	cardsOutput := linq.Map(cards, func(card *entities.Card) *dtos.CardOutput {
		output := &dtos.CardOutput{
			ID:                card.ID.String(),
			Name:              card.Name.String(),
			DueDay:            card.DueDay.Int(),
			ClosingOffsetDays: card.ClosingOffsetDays.Int(),
			CreatedAt:         card.CreatedAt.ValueOr(time.Time{}),
		}
		if !card.UpdatedAt.ValueOr(time.Time{}).IsZero() {
			output.UpdatedAt = card.UpdatedAt.ValueOr(time.Time{})
		}
		return output
	})

	return cardsOutput, nil
}
