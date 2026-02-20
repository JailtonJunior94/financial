package httperrors

import (
	"errors"
	"net/http"
	"testing"

	customerrors "github.com/jailtonjunior94/financial/pkg/custom_errors"
)

func TestErrorMapper_MapError_ValidationErrors(t *testing.T) {
	// Sentinel errors to test the extra-mappings registration mechanism.
	// This mirrors how each domain module registers its own errors at startup.
	errDomainValidation := errors.New("domain validation error")
	errDomainNotFound := errors.New("domain not found error")
	errDomainConflict := errors.New("domain conflict error")
	errServiceUnavailable := errors.New("service unavailable error")

	mapper := NewErrorMapper(map[error]ErrorMapping{
		errDomainValidation:   {Status: http.StatusBadRequest, Message: "Validation failed"},
		errDomainNotFound:     {Status: http.StatusNotFound, Message: "Not found"},
		errDomainConflict:     {Status: http.StatusConflict, Message: "Conflict"},
		errServiceUnavailable: {Status: http.StatusServiceUnavailable, Message: "Service unavailable"},
	})

	tests := []struct {
		name           string
		err            error
		expectedStatus int
		description    string
	}{
		// Extra domain-specific mappings registered at construction
		{
			name:           "registered domain validation error returns 400",
			err:            errDomainValidation,
			expectedStatus: http.StatusBadRequest,
			description:    "Domain validation errors registered via extra should return 400 Bad Request",
		},
		{
			name:           "registered domain not-found error returns 404",
			err:            errDomainNotFound,
			expectedStatus: http.StatusNotFound,
			description:    "Domain not-found errors registered via extra should return 404 Not Found",
		},
		{
			name:           "registered domain conflict error returns 409",
			err:            errDomainConflict,
			expectedStatus: http.StatusConflict,
			description:    "Domain conflict errors registered via extra should return 409 Conflict",
		},
		{
			name:           "registered service unavailable error returns 503",
			err:            errServiceUnavailable,
			expectedStatus: http.StatusServiceUnavailable,
			description:    "Service unavailable errors registered via extra should return 503",
		},

		// Base mappings from pkg/custom_errors (always present without extra)
		{
			name:           "customerrors.ErrBudgetInvalidTotal returns 400",
			err:            customerrors.ErrBudgetInvalidTotal,
			expectedStatus: http.StatusBadRequest,
			description:    "Custom error budget validation should return 400 Bad Request",
		},
		{
			name:           "customerrors.ErrUserNotFound returns 404",
			err:            customerrors.ErrUserNotFound,
			expectedStatus: http.StatusNotFound,
			description:    "User not found should return 404 Not Found",
		},
		{
			name:           "customerrors.ErrEmailAlreadyExists returns 409",
			err:            customerrors.ErrEmailAlreadyExists,
			expectedStatus: http.StatusConflict,
			description:    "Email already exists should return 409 Conflict",
		},
		{
			name:           "customerrors.ErrUnauthorized returns 401",
			err:            customerrors.ErrUnauthorized,
			expectedStatus: http.StatusUnauthorized,
			description:    "Unauthorized should return 401 Unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapping := mapper.MapError(tt.err)

			if mapping.Status != tt.expectedStatus {
				t.Errorf("%s: expected status %d, got %d. Error: %v",
					tt.description,
					tt.expectedStatus,
					mapping.Status,
					tt.err,
				)
			}
		})
	}
}

func TestErrorMapper_MapError_UnmappedErrors(t *testing.T) {
	mapper := NewErrorMapper()

	tests := []struct {
		name           string
		err            error
		expectedStatus int
		description    string
	}{
		{
			name:           "Unmapped error returns 500",
			err:            errors.New("some unknown error"),
			expectedStatus: http.StatusInternalServerError,
			description:    "Unknown errors should return 500 Internal Server Error",
		},
		{
			name:           "Nil error returns 500",
			err:            nil,
			expectedStatus: http.StatusInternalServerError,
			description:    "Nil errors should return 500 Internal Server Error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapping := mapper.MapError(tt.err)

			if mapping.Status != tt.expectedStatus {
				t.Errorf("%s: expected status %d, got %d",
					tt.description,
					tt.expectedStatus,
					mapping.Status,
				)
			}
		})
	}
}

func TestErrorMapper_ExtraMappingsOverrideBase(t *testing.T) {
	// Verify that extra mappings can override base mappings.
	mapper := NewErrorMapper(map[error]ErrorMapping{
		customerrors.ErrUserNotFound: {Status: http.StatusBadRequest, Message: "Custom override"},
	})

	mapping := mapper.MapError(customerrors.ErrUserNotFound)
	if mapping.Status != http.StatusBadRequest {
		t.Errorf("extra mapping should override base: expected 400, got %d", mapping.Status)
	}
	if mapping.Message != "Custom override" {
		t.Errorf("extra mapping message mismatch: expected 'Custom override', got %q", mapping.Message)
	}
}
