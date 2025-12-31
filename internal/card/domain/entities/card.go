package entities

import (
	"time"

	sharedVos "github.com/JailtonJunior94/devkit-go/pkg/vos"
	"github.com/jailtonjunior94/financial/internal/card/domain/vos"
)

type Card struct {
	ID        sharedVos.UUID
	UserID    sharedVos.UUID
	Name      vos.CardName
	DueDay    vos.DueDay
	CreatedAt sharedVos.NullableTime
	UpdatedAt sharedVos.NullableTime
	DeletedAt sharedVos.NullableTime
}

func NewCard(userID sharedVos.UUID, name vos.CardName, dueDay vos.DueDay) (*Card, error) {
	card := &Card{
		Name:      name,
		UserID:    userID,
		DueDay:    dueDay,
		CreatedAt: sharedVos.NewNullableTime(time.Now()),
	}
	return card, nil
}

func (c *Card) Update(name string, dueDay int) error {
	cardName, err := vos.NewCardName(name)
	if err != nil {
		return err
	}

	cardDueDay, err := vos.NewDueDay(dueDay)
	if err != nil {
		return err
	}

	c.Name = cardName
	c.DueDay = cardDueDay
	c.UpdatedAt = sharedVos.NewNullableTime(time.Now())

	return nil
}

func (c *Card) Delete() *Card {
	c.DeletedAt = sharedVos.NewNullableTime(time.Now())
	return c
}
