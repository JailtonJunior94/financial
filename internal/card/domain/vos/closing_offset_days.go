package vos

import (
	"errors"
	"fmt"
)

const (
	// DefaultClosingOffsetDays é o padrão brasileiro (Nubank, BTG, XP).
	DefaultClosingOffsetDays = 7
	MinClosingOffsetDays     = 1
	MaxClosingOffsetDays     = 31
)

var (
	ErrClosingOffsetDaysRequired = errors.New("closing offset days is required")
	ErrClosingOffsetDaysInvalid  = errors.New("closing offset days must be between 1 and 31")
)

// ClosingOffsetDays representa quantos dias ANTES do vencimento ocorre o fechamento da fatura.
// Exemplo: vencimento dia 10, offset 7 → fechamento dia 3.
type ClosingOffsetDays struct {
	Value int
	Valid bool
}

// NewClosingOffsetDays cria um novo ClosingOffsetDays validado.
func NewClosingOffsetDays(days int) (ClosingOffsetDays, error) {
	if days < MinClosingOffsetDays || days > MaxClosingOffsetDays {
		return ClosingOffsetDays{}, fmt.Errorf("invalid closing offset days: %w", ErrClosingOffsetDaysInvalid)
	}
	return ClosingOffsetDays{Value: days, Valid: true}, nil
}

// NewDefaultClosingOffsetDays cria um ClosingOffsetDays com o valor padrão brasileiro (7 dias).
func NewDefaultClosingOffsetDays() ClosingOffsetDays {
	return ClosingOffsetDays{Value: DefaultClosingOffsetDays, Valid: true}
}

// Int retorna o valor inteiro.
func (c ClosingOffsetDays) Int() int {
	return c.Value
}
