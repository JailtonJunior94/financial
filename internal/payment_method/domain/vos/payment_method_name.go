package vos

import (
	"fmt"
	"strings"

	customErrors "github.com/jailtonjunior94/financial/pkg/custom_errors"
)

type PaymentMethodName struct {
	Value *string
	Valid bool
}

func NewPaymentMethodName(name string) (PaymentMethodName, error) {
	trimmed := strings.TrimSpace(name)
	if len(trimmed) == 0 {
		return PaymentMethodName{}, fmt.Errorf("invalid payment method name: %w", customErrors.ErrNameIsRequired)
	}
	if len(trimmed) > 255 {
		return PaymentMethodName{}, fmt.Errorf("invalid payment method name: %w", customErrors.ErrTooLong)
	}
	return PaymentMethodName{Value: &trimmed, Valid: true}, nil
}

func (v PaymentMethodName) String() string {
	if v.Value != nil {
		return *v.Value
	}
	return ""
}
