package vos

import (
	"errors"
	"fmt"
	"time"
)

var (
	// ErrInvalidReferenceMonth indica formato inválido de mês de referência
	ErrInvalidReferenceMonth = errors.New("invalid reference month format, expected YYYY-MM")
)

// ReferenceMonth representa um mês de referência no formato YYYY-MM.
// Versão consolidada usada por todos os módulos.
type ReferenceMonth struct {
	year  int
	month time.Month
}

// NewReferenceMonth cria um novo ReferenceMonth a partir de uma string YYYY-MM.
func NewReferenceMonth(value string) (ReferenceMonth, error) {
	date, err := time.Parse("2006-01", value)
	if err != nil {
		return ReferenceMonth{}, fmt.Errorf("%w: %s", ErrInvalidReferenceMonth, value)
	}
	return ReferenceMonth{
		year:  date.Year(),
		month: date.Month(),
	}, nil
}

// NewReferenceMonthFromDate cria um ReferenceMonth a partir de um time.Time.
func NewReferenceMonthFromDate(date time.Time) ReferenceMonth {
	return ReferenceMonth{
		year:  date.Year(),
		month: date.Month(),
	}
}

// NewReferenceMonthFromYearMonth cria um ReferenceMonth a partir de year e month.
func NewReferenceMonthFromYearMonth(year, month int) (ReferenceMonth, error) {
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

// String retorna o valor no formato YYYY-MM.
func (r ReferenceMonth) String() string {
	return fmt.Sprintf("%04d-%02d", r.year, int(r.month))
}

// Year retorna o ano.
func (r ReferenceMonth) Year() int {
	return r.year
}

// Month retorna o mês.
func (r ReferenceMonth) Month() time.Month {
	return r.month
}

// Equal verifica se dois ReferenceMonth são iguais.
func (r ReferenceMonth) Equal(other ReferenceMonth) bool {
	return r.year == other.year && r.month == other.month
}

// FirstDay retorna o primeiro dia do mês como time.Time.
func (r ReferenceMonth) FirstDay() time.Time {
	return time.Date(r.year, r.month, 1, 0, 0, 0, 0, time.UTC)
}

// LastDay retorna o último dia do mês.
func (r ReferenceMonth) LastDay() time.Time {
	return r.FirstDay().AddDate(0, 1, 0).Add(-time.Second)
}

// AddMonths adiciona N meses e retorna um novo ReferenceMonth.
func (r ReferenceMonth) AddMonths(months int) ReferenceMonth {
	firstDay := r.FirstDay().AddDate(0, months, 0)
	return NewReferenceMonthFromDate(firstDay)
}

// ToTime é um alias para FirstDay (para compatibilidade).
func (r ReferenceMonth) ToTime() time.Time {
	return r.FirstDay()
}
