package metrics

import (
	"errors"
	"strings"

	customErrors "github.com/jailtonjunior94/financial/pkg/custom_errors"
)

// ClassifyError classifica um erro em categorias para métricas
func ClassifyError(err error) string {
	if err == nil {
		return ""
	}

	// Unwrap CustomError se necessário
	var customErr *customErrors.CustomError
	if errors.As(err, &customErr) {
		if customErr.Err != nil {
			err = customErr.Err
		}
	}

	errMsg := strings.ToLower(err.Error())

	// Erros de validação
	if strings.Contains(errMsg, "invalid") ||
		strings.Contains(errMsg, "must be") ||
		strings.Contains(errMsg, "required") ||
		strings.Contains(errMsg, "validation") {
		return ErrorTypeValidation
	}

	// Erros de not found
	if strings.Contains(errMsg, "not found") ||
		strings.Contains(errMsg, "does not exist") {
		return ErrorTypeNotFound
	}

	// Erros de parsing/conversão
	if strings.Contains(errMsg, "parse") ||
		strings.Contains(errMsg, "convert") ||
		strings.Contains(errMsg, "uuid") {
		return ErrorTypeParsing
	}

	// Erros de repositório/banco
	if strings.Contains(errMsg, "database") ||
		strings.Contains(errMsg, "sql") ||
		strings.Contains(errMsg, "repository") ||
		strings.Contains(errMsg, "connection") {
		return ErrorTypeRepository
	}

	return ErrorTypeUnknown
}
