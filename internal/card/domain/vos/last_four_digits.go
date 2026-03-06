package vos

import (
	"fmt"
	"regexp"

	domain "github.com/jailtonjunior94/financial/internal/card/domain"
)

var lastFourDigitsRegex = regexp.MustCompile(`^[0-9]{4}$`)

type LastFourDigits struct {
	Value string
}

func NewLastFourDigits(d string) (LastFourDigits, error) {
	if !lastFourDigitsRegex.MatchString(d) {
		return LastFourDigits{}, fmt.Errorf("invalid last four digits: %w", domain.ErrInvalidLastFourDigits)
	}
	return LastFourDigits{Value: d}, nil
}

func (l LastFourDigits) String() string {
	return l.Value
}
