package factories

import (
	"fmt"

	"github.com/jailtonjunior94/financial/internal/card/domain/entities"
	"github.com/jailtonjunior94/financial/internal/card/domain/vos"

	sharedVos "github.com/JailtonJunior94/devkit-go/pkg/vos"
)

type CreateCardParams struct {
	UserID            string
	Name              string
	Type              string
	Flag              string
	LastFourDigits    string
	DueDay            int
	ClosingOffsetDays int
}

func CreateCard(params CreateCardParams) (*entities.Card, error) {
	id, err := sharedVos.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("error generating card id: %v", err)
	}

	user, err := sharedVos.NewUUIDFromString(params.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user_id: %s", params.UserID)
	}

	cardName, err := vos.NewCardName(params.Name)
	if err != nil {
		return nil, err
	}

	cardType, err := vos.NewCardType(params.Type)
	if err != nil {
		return nil, err
	}

	cardFlag, err := vos.NewCardFlag(params.Flag)
	if err != nil {
		return nil, err
	}

	digits, err := vos.NewLastFourDigits(params.LastFourDigits)
	if err != nil {
		return nil, err
	}

	var dueDay vos.DueDay
	var closingOffsetDays vos.ClosingOffsetDays

	if cardType.IsCredit() {
		dueDay, err = vos.NewDueDay(params.DueDay)
		if err != nil {
			return nil, err
		}

		if params.ClosingOffsetDays == 0 {
			closingOffsetDays = vos.NewDefaultClosingOffsetDays()
		} else {
			closingOffsetDays, err = vos.NewClosingOffsetDays(params.ClosingOffsetDays)
			if err != nil {
				return nil, err
			}
		}
	}

	card, err := entities.NewCard(user, cardName, cardType, cardFlag, digits, dueDay, closingOffsetDays)
	if err != nil {
		return nil, fmt.Errorf("error creating card: %w", err)
	}

	card.ID = id
	return card, nil
}
