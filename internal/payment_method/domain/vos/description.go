package vos

import (
	"errors"
	"strings"
)

var ErrDescriptionTooLong = errors.New("description cannot be more than 500 characters")

type Description struct {
	Value *string
	Valid bool
}

func NewDescription(description string) (Description, error) {
	trimmed := strings.TrimSpace(description)

	// Description is optional
	if len(trimmed) == 0 {
		return Description{Valid: false}, nil
	}

	if len(trimmed) > 500 {
		return Description{}, ErrDescriptionTooLong
	}

	return Description{Value: &trimmed, Valid: true}, nil
}

func (v Description) String() string {
	if v.Value != nil {
		return *v.Value
	}
	return ""
}
