package entities

import (
	"time"

	sharedVos "github.com/JailtonJunior94/devkit-go/pkg/vos"
	"github.com/jailtonjunior94/financial/internal/card/domain/vos"
)

type Card struct {
	ID                sharedVos.UUID
	UserID            sharedVos.UUID
	Name              vos.CardName
	Type              vos.CardType
	Flag              vos.CardFlag
	LastFourDigits    vos.LastFourDigits
	DueDay            vos.DueDay
	ClosingOffsetDays vos.ClosingOffsetDays
	CreatedAt         sharedVos.NullableTime
	UpdatedAt         sharedVos.NullableTime
	DeletedAt         sharedVos.NullableTime
}

func NewCard(
	userID sharedVos.UUID,
	name vos.CardName,
	cardType vos.CardType,
	flag vos.CardFlag,
	lastFourDigits vos.LastFourDigits,
	dueDay vos.DueDay,
	closingOffsetDays vos.ClosingOffsetDays,
) (*Card, error) {
	card := &Card{
		UserID:         userID,
		Name:           name,
		Type:           cardType,
		Flag:           flag,
		LastFourDigits: lastFourDigits,
		CreatedAt:      sharedVos.NewNullableTime(time.Now()),
	}
	if cardType.IsCredit() {
		card.DueDay = dueDay
		card.ClosingOffsetDays = closingOffsetDays
	}
	return card, nil
}

func (c *Card) Update(name, flag, lastFourDigits string, dueDay, closingOffsetDays int) error {
	cardName, err := vos.NewCardName(name)
	if err != nil {
		return err
	}
	cardFlag, err := vos.NewCardFlag(flag)
	if err != nil {
		return err
	}
	digits, err := vos.NewLastFourDigits(lastFourDigits)
	if err != nil {
		return err
	}
	c.Name = cardName
	c.Flag = cardFlag
	c.LastFourDigits = digits
	if c.Type.IsCredit() {
		cardDueDay, err := vos.NewDueDay(dueDay)
		if err != nil {
			return err
		}
		offset, err := vos.NewClosingOffsetDays(closingOffsetDays)
		if err != nil {
			return err
		}
		c.DueDay = cardDueDay
		c.ClosingOffsetDays = offset
	}
	c.UpdatedAt = sharedVos.NewNullableTime(time.Now())
	return nil
}

func (c *Card) Delete() *Card {
	c.DeletedAt = sharedVos.NewNullableTime(time.Now())
	return c
}
