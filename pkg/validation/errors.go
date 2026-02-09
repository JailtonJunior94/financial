package validation

import "fmt"

// ValidationError representa um erro de validação de campo.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Error implementa a interface error.
func (v ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", v.Field, v.Message)
}

// NewValidationError cria um novo erro de validação.
func NewValidationError(field, message string) ValidationError {
	return ValidationError{
		Field:   field,
		Message: message,
	}
}

// ValidationErrors é uma lista de erros de validação.
type ValidationErrors []ValidationError

// Error implementa a interface error.
func (v ValidationErrors) Error() string {
	if len(v) == 0 {
		return ""
	}
	if len(v) == 1 {
		return v[0].Error()
	}
	return fmt.Sprintf("validation failed with %d errors", len(v))
}

// Add adiciona um erro à lista.
func (v *ValidationErrors) Add(field, message string) {
	*v = append(*v, NewValidationError(field, message))
}

// AddError adiciona um ValidationError à lista.
func (v *ValidationErrors) AddError(err ValidationError) {
	*v = append(*v, err)
}

// HasErrors retorna true se há erros.
func (v ValidationErrors) HasErrors() bool {
	return len(v) > 0
}

// ToError retorna nil se não há erros, ou ValidationErrors se houver.
func (v ValidationErrors) ToError() error {
	if !v.HasErrors() {
		return nil
	}
	return v
}
