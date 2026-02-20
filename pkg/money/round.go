// Package money provê utilitários de arredondamento bancário (half-even / ABNT NBR 5891)
// e construtores de Value Objects monetários e percentuais com esse padrão.
//
// O padrão half-even difere do math.Round (half-away-from-zero) no caso de exatos .5:
//
//	math.Round(2.5) = 3 (half-away)
//	BankersRound(2.5) = 2 (half-even: arredonda para o par mais próximo)
//	math.Round(3.5) = 4 (half-away)
//	BankersRound(3.5) = 4 (half-even: 4 já é par)
//
// Para entradas com ≤ 2 casas decimais (padrão BRL), os resultados são idênticos ao
// math.Round, mas a implementação é compatível com o padrão BACEN.
package money

import (
	"fmt"
	"math"
	"strconv"

	"github.com/JailtonJunior94/devkit-go/pkg/vos"
)

// BankersRound implementa IEEE 754 round-half-to-even (arredondamento bancário).
// Empates (fração exatamente 0.5) são arredondados para o inteiro par mais próximo.
// Usa epsilon para compensar imprecisão de representação float64.
func BankersRound(x float64) int64 {
	floor := math.Floor(x)
	diff := x - floor

	const epsilon = 1e-9

	switch {
	case diff < 0.5-epsilon:
		return int64(floor)
	case diff > 0.5+epsilon:
		return int64(floor) + 1
	default:
		// Empate: arredonda para o inteiro par mais próximo
		floorInt := int64(floor)
		if floorInt%2 == 0 {
			return floorInt
		}
		return floorInt + 1
	}
}

// NewMoney converte uma string decimal em um Money VO usando arredondamento half-even.
// É o construtor padrão para valores monetários provenientes de entrada do usuário.
func NewMoney(s string, currency vos.Currency) (vos.Money, error) {
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return vos.Money{}, fmt.Errorf("invalid monetary amount %q: %w", s, err)
	}
	cents := BankersRound(val * 100)
	return vos.NewMoney(cents, currency)
}

// NewMoneyBRL converte uma string decimal em um Money VO em BRL usando arredondamento half-even.
func NewMoneyBRL(s string) (vos.Money, error) {
	return NewMoney(s, vos.CurrencyBRL)
}

// NewPercentageFromString converte uma string decimal em um Percentage VO usando arredondamento half-even.
// O VO armazena internamente com escala 1000 (25.5% = 25500).
func NewPercentageFromString(s string) (vos.Percentage, error) {
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return vos.Percentage{}, fmt.Errorf("invalid percentage %q: %w", s, err)
	}
	raw := BankersRound(val * 1000)
	return vos.NewPercentage(raw)
}
