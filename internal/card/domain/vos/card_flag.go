package vos

import (
	"fmt"
	"slices"

	domain "github.com/jailtonjunior94/financial/internal/card/domain"
)

var validCardFlags = []string{"visa", "mastercard", "elo", "amex", "hipercard"}

type CardFlag struct {
	Value string
}

func NewCardFlag(f string) (CardFlag, error) {
	if !slices.Contains(validCardFlags, f) {
		return CardFlag{}, fmt.Errorf("invalid card flag: %w", domain.ErrInvalidCardFlag)
	}
	return CardFlag{Value: f}, nil
}

func (cf CardFlag) String() string {
	return cf.Value
}
