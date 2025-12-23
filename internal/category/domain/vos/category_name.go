package vos

import (
	"fmt"
	"strings"

	customErrors "github.com/jailtonjunior94/financial/pkg/custom_errors"
)

type CategoryName struct {
	Value *string
	Valid bool
}

func NewCategoryName(name string) (CategoryName, error) {
	trimmed := strings.TrimSpace(name)
	if len(trimmed) == 0 {
		return CategoryName{}, fmt.Errorf("invalid category name: %w", customErrors.ErrNameIsRequired)
	}
	if len(trimmed) > 255 {
		return CategoryName{}, fmt.Errorf("invalid category name: %w", customErrors.ErrTooLong)
	}
	return CategoryName{Value: &trimmed, Valid: true}, nil
}

func (v CategoryName) String() string {
	if v.Value != nil {
		return *v.Value
	}
	return ""
}
