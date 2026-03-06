package vos

import (
	"fmt"

	domain "github.com/jailtonjunior94/financial/internal/card/domain"
)

const (
	CardTypeCredit = "credit"
	CardTypeDebit  = "debit"
)

type CardType struct {
	Value string
}

func NewCardType(t string) (CardType, error) {
	if t != CardTypeCredit && t != CardTypeDebit {
		return CardType{}, fmt.Errorf("invalid card type: %w", domain.ErrInvalidCardType)
	}
	return CardType{Value: t}, nil
}

func (ct CardType) IsCredit() bool {
	return ct.Value == CardTypeCredit
}

func (ct CardType) IsDebit() bool {
	return ct.Value == CardTypeDebit
}

func (ct CardType) String() string {
	return ct.Value
}
