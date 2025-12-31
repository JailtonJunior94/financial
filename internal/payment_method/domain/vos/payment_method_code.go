package vos

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var (
	ErrCodeIsRequired = errors.New("code is required")
	ErrCodeInvalid    = errors.New("code must contain only uppercase letters, numbers and underscores")
)

var codeRegex = regexp.MustCompile(`^[A-Z0-9_]+$`)

type PaymentMethodCode struct {
	Value *string
	Valid bool
}

func NewPaymentMethodCode(code string) (PaymentMethodCode, error) {
	trimmed := strings.TrimSpace(code)
	normalized := strings.ToUpper(trimmed)

	if len(normalized) == 0 {
		return PaymentMethodCode{}, fmt.Errorf("invalid payment method code: %w", ErrCodeIsRequired)
	}

	if !codeRegex.MatchString(normalized) {
		return PaymentMethodCode{}, fmt.Errorf("invalid payment method code: %w", ErrCodeInvalid)
	}

	if len(normalized) > 50 {
		return PaymentMethodCode{}, errors.New("code cannot be more than 50 characters")
	}

	return PaymentMethodCode{Value: &normalized, Valid: true}, nil
}

func (v PaymentMethodCode) String() string {
	if v.Value != nil {
		return *v.Value
	}
	return ""
}
