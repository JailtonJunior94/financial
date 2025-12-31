package factories

import (
	"fmt"

	"github.com/jailtonjunior94/financial/internal/card/domain/entities"

	sharedVos "github.com/JailtonJunior94/devkit-go/pkg/vos"
	"github.com/jailtonjunior94/financial/internal/card/domain/vos"
)

func CreateCard(userID, name string, dueDay int) (*entities.Card, error) {
	id, err := sharedVos.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("error generating card id: %v", err)
	}

	user, err := sharedVos.NewUUIDFromString(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user_id: %s", userID)
	}

	cardName, err := vos.NewCardName(name)
	if err != nil {
		return nil, err
	}

	cardDueDay, err := vos.NewDueDay(dueDay)
	if err != nil {
		return nil, err
	}

	card, err := entities.NewCard(user, cardName, cardDueDay)
	if err != nil {
		return nil, fmt.Errorf("error creating card: %w", err)
	}

	card.ID = id
	return card, nil
}
