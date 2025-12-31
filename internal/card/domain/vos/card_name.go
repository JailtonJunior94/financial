package vos

import (
	"fmt"
	"strings"

	customErrors "github.com/jailtonjunior94/financial/pkg/custom_errors"
)

type CardName struct {
	Value *string
	Valid bool
}

func NewCardName(name string) (CardName, error) {
	trimmed := strings.TrimSpace(name)
	if len(trimmed) == 0 {
		return CardName{}, fmt.Errorf("invalid card name: %w", customErrors.ErrNameIsRequired)
	}
	if len(trimmed) > 255 {
		return CardName{}, fmt.Errorf("invalid card name: %w", customErrors.ErrTooLong)
	}
	return CardName{Value: &trimmed, Valid: true}, nil
}

func (v CardName) String() string {
	if v.Value != nil {
		return *v.Value
	}
	return ""
}
