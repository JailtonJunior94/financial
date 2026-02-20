package validation

import (
	"regexp"
	"slices"
	"strings"
	"time"
)

var (
	// UUIDRegex valida UUIDs em qualquer formato (v1, v4, v7, ULID, etc)
	UUIDRegex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

	// DateRegex valida datas no formato YYYY-MM-DD
	DateRegex = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

	// MonthRegex valida meses no formato YYYY-MM
	MonthRegex = regexp.MustCompile(`^\d{4}-\d{2}$`)

	// MoneyRegex valida valores monetários (até 2 casas decimais, máximo 15 dígitos inteiros)
	MoneyRegex = regexp.MustCompile(`^\d{1,15}(\.\d{1,2})?$`)

	// PercentageRegex valida valores percentuais (até 3 casas decimais, alinhado com NUMERIC(6,3))
	PercentageRegex = regexp.MustCompile(`^\d+(\.\d{1,3})?$`)
)

// IsRequired verifica se o valor não está vazio.
func IsRequired(value string) bool {
	return strings.TrimSpace(value) != ""
}

// IsUUID verifica se é um UUID válido.
func IsUUID(value string) bool {
	return UUIDRegex.MatchString(value)
}

// IsDate verifica se é uma data válida no formato YYYY-MM-DD.
func IsDate(value string) bool {
	if !DateRegex.MatchString(value) {
		return false
	}
	_, err := time.Parse("2006-01-02", value)
	return err == nil
}

// IsMonth verifica se é um mês válido no formato YYYY-MM.
func IsMonth(value string) bool {
	if !MonthRegex.MatchString(value) {
		return false
	}
	_, err := time.Parse("2006-01", value)
	return err == nil
}

// IsMoney verifica se é um valor monetário válido.
func IsMoney(value string) bool {
	if value == "" {
		return false
	}
	return MoneyRegex.MatchString(value)
}

// IsPercentage verifica se é um valor percentual válido (até 3 casas decimais).
func IsPercentage(value string) bool {
	if value == "" {
		return false
	}
	return PercentageRegex.MatchString(value)
}

// IsPositiveInt verifica se é um inteiro positivo.
func IsPositiveInt(value int) bool {
	return value > 0
}

// IsInRange verifica se o valor está dentro do intervalo.
func IsInRange(value, min, max int) bool {
	return value >= min && value <= max
}

// IsMaxLength verifica se o comprimento não excede o máximo.
func IsMaxLength(value string, max int) bool {
	return len(value) <= max
}

// IsMinLength verifica se o comprimento é pelo menos o mínimo.
func IsMinLength(value string, min int) bool {
	return len(value) >= min
}

// IsOneOf verifica se o valor está na lista de valores permitidos.
func IsOneOf(value string, allowed []string) bool {
	return slices.Contains(allowed, value)
}
