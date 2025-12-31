package vos

import (
	"errors"
	"fmt"
	"regexp"
	"time"
)

var (
	ErrInvalidReferenceMonth = errors.New("invalid reference month format, expected YYYY-MM")
)

var referenceMonthRegex = regexp.MustCompile(`^\d{4}-(0[1-9]|1[0-2])$`)

// ReferenceMonth representa um mês de referência no formato YYYY-MM
type ReferenceMonth struct {
	value string
	year  int
	month time.Month
}

// NewReferenceMonth cria um novo ReferenceMonth validado
func NewReferenceMonth(yearMonth string) (ReferenceMonth, error) {
	if !referenceMonthRegex.MatchString(yearMonth) {
		return ReferenceMonth{}, ErrInvalidReferenceMonth
	}

	// Parse to validate it's a real date
	t, err := time.Parse("2006-01", yearMonth)
	if err != nil {
		return ReferenceMonth{}, fmt.Errorf("%w: %v", ErrInvalidReferenceMonth, err)
	}

	return ReferenceMonth{
		value: yearMonth,
		year:  t.Year(),
		month: t.Month(),
	}, nil
}

// NewReferenceMonthFromDate cria um ReferenceMonth a partir de um time.Time
func NewReferenceMonthFromDate(date time.Time) ReferenceMonth {
	return ReferenceMonth{
		value: date.Format("2006-01"),
		year:  date.Year(),
		month: date.Month(),
	}
}

// String retorna o valor no formato YYYY-MM
func (r ReferenceMonth) String() string {
	return r.value
}

// Year retorna o ano
func (r ReferenceMonth) Year() int {
	return r.year
}

// Month retorna o mês
func (r ReferenceMonth) Month() time.Month {
	return r.month
}

// Equals verifica se dois ReferenceMonth são iguais
func (r ReferenceMonth) Equals(other ReferenceMonth) bool {
	return r.value == other.value
}

// ToTime retorna o primeiro dia do mês como time.Time
func (r ReferenceMonth) ToTime() time.Time {
	return time.Date(r.year, r.month, 1, 0, 0, 0, 0, time.UTC)
}
