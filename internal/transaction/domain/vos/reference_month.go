package vos

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrInvalidReferenceMonth = errors.New("invalid reference month")
)

// ReferenceMonth representa um mês de referência (YYYY-MM).
type ReferenceMonth struct {
	year  int
	month time.Month
}

// NewReferenceMonth cria um novo ReferenceMonth validado.
func NewReferenceMonth(year int, month int) (ReferenceMonth, error) {
	if year < 1900 || year > 2100 {
		return ReferenceMonth{}, fmt.Errorf("%w: invalid year %d", ErrInvalidReferenceMonth, year)
	}
	if month < 1 || month > 12 {
		return ReferenceMonth{}, fmt.Errorf("%w: invalid month %d", ErrInvalidReferenceMonth, month)
	}
	return ReferenceMonth{
		year:  year,
		month: time.Month(month),
	}, nil
}

// NewReferenceMonthFromDate cria um ReferenceMonth a partir de uma data.
func NewReferenceMonthFromDate(date time.Time) ReferenceMonth {
	return ReferenceMonth{
		year:  date.Year(),
		month: date.Month(),
	}
}

// NewReferenceMonthFromString cria um ReferenceMonth a partir de uma string (YYYY-MM).
func NewReferenceMonthFromString(value string) (ReferenceMonth, error) {
	date, err := time.Parse("2006-01", value)
	if err != nil {
		return ReferenceMonth{}, fmt.Errorf("%w: %s", ErrInvalidReferenceMonth, value)
	}
	return NewReferenceMonthFromDate(date), nil
}

// Year retorna o ano.
func (r ReferenceMonth) Year() int {
	return r.year
}

// Month retorna o mês.
func (r ReferenceMonth) Month() time.Month {
	return r.month
}

// String retorna a representação em string (YYYY-MM).
func (r ReferenceMonth) String() string {
	return fmt.Sprintf("%04d-%02d", r.year, int(r.month))
}

// FirstDay retorna o primeiro dia do mês.
func (r ReferenceMonth) FirstDay() time.Time {
	return time.Date(r.year, r.month, 1, 0, 0, 0, 0, time.UTC)
}

// LastDay retorna o último dia do mês.
func (r ReferenceMonth) LastDay() time.Time {
	return r.FirstDay().AddDate(0, 1, 0).AddDate(0, 0, -1)
}

// Equal verifica se dois ReferenceMonth são iguais.
func (r ReferenceMonth) Equal(other ReferenceMonth) bool {
	return r.year == other.year && r.month == other.month
}

// AddMonths adiciona N meses e retorna um novo ReferenceMonth.
func (r ReferenceMonth) AddMonths(months int) ReferenceMonth {
	firstDay := r.FirstDay().AddDate(0, months, 0)
	return NewReferenceMonthFromDate(firstDay)
}
