package vos

import (
	"fmt"

	customErrors "github.com/jailtonjunior94/financial/pkg/custom_errors"
)

type CategorySequence struct {
	Sequence *uint
	Valid    bool
}

func NewCategorySequence(i uint) (CategorySequence, error) {
	if i == 0 {
		return CategorySequence{}, fmt.Errorf("invalid category sequence: %w", customErrors.ErrSequenceIsRequired)
	}
	if i > 1000 {
		return CategorySequence{}, fmt.Errorf("invalid category sequence: %w", customErrors.ErrSequenceTooLarge)
	}
	return CategorySequence{Sequence: &i, Valid: true}, nil
}

func (n CategorySequence) Value() uint {
	if n.Sequence != nil {
		return *n.Sequence
	}
	return 0
}
