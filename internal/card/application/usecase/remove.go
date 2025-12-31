package usecase

import (
	"context"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
	"github.com/jailtonjunior94/financial/internal/card/domain/interfaces"
	customErrors "github.com/jailtonjunior94/financial/pkg/custom_errors"
)

type (
	RemoveCardUseCase interface {
		Execute(ctx context.Context, userID, id string) error
	}

	removeCardUseCase struct {
		o11y       observability.Observability
		repository interfaces.CardRepository
	}
)

func NewRemoveCardUseCase(
	o11y observability.Observability,
	repository interfaces.CardRepository,
) RemoveCardUseCase {
	return &removeCardUseCase{
		o11y:       o11y,
		repository: repository,
	}
}

func (u *removeCardUseCase) Execute(ctx context.Context, userID, id string) error {
	ctx, span := u.o11y.Tracer().Start(ctx, "remove_card_usecase.execute")
	defer span.End()

	user, err := vos.NewUUIDFromString(userID)
	if err != nil {
		span.AddEvent(
			"error parsing user id",
			observability.Field{Key: "user_id", Value: userID},
			observability.Field{Key: "error", Value: err},
		)

		return err
	}

	cardID, err := vos.NewUUIDFromString(id)
	if err != nil {
		span.AddEvent(
			"error parsing card id",
			observability.Field{Key: "card_id", Value: id},
			observability.Field{Key: "error", Value: err},
		)

		return err
	}

	card, err := u.repository.FindByID(ctx, user, cardID)
	if err != nil {
		span.AddEvent(
			"error finding card by id",
			observability.Field{Key: "user_id", Value: userID},
			observability.Field{Key: "card_id", Value: id},
			observability.Field{Key: "error", Value: err},
		)
		u.o11y.Logger().Error(ctx, "error finding card by id",
			observability.Error(err),
			observability.String("user_id", userID),
			observability.String("card_id", id))
		return err
	}

	if card == nil {
		span.AddEvent(
			"card not found",
			observability.Field{Key: "user_id", Value: userID},
			observability.Field{Key: "card_id", Value: id},
		)
		u.o11y.Logger().Error(ctx, "card not found",
			observability.Error(customErrors.ErrCardNotFound),
			observability.String("user_id", userID),
			observability.String("card_id", id))
		return customErrors.ErrCardNotFound
	}

	if err := u.repository.Update(ctx, card.Delete()); err != nil {
		span.AddEvent(
			"error deleting card in repository",
			observability.Field{Key: "user_id", Value: userID},
			observability.Field{Key: "card_id", Value: id},
			observability.Field{Key: "error", Value: err},
		)
		u.o11y.Logger().Error(ctx, "error deleting card in repository",
			observability.Error(err),
			observability.String("user_id", userID),
			observability.String("card_id", id))
		return err
	}

	return nil
}
