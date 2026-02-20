package httperrors

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	customerrors "github.com/jailtonjunior94/financial/pkg/custom_errors"
	"github.com/jailtonjunior94/financial/pkg/validation"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// ErrorHandler handles HTTP errors by mapping them to appropriate HTTP responses.
type ErrorHandler interface {
	HandleError(w http.ResponseWriter, r *http.Request, err error)
}

type errorHandler struct {
	o11y   observability.Observability
	mapper ErrorMapper
}

// NewErrorHandler creates a new error handler with observability.
// Domain-specific error mappings can be provided so each module registers
// its own errors without the pkg package importing internal domain packages.
func NewErrorHandler(o11y observability.Observability, extra ...map[error]ErrorMapping) ErrorHandler {
	return &errorHandler{
		o11y:   o11y,
		mapper: NewErrorMapper(extra...),
	}
}

// HandleError handles an error by:
// 1. Unwrapping CustomError to get the original error
// 2. Mapping the error to an HTTP status code and message
// 3. Adding attributes to the span for tracing
// 4. Logging the error once with appropriate level
// 5. Constructing and writing a ProblemDetail response.
func (h *errorHandler) HandleError(w http.ResponseWriter, r *http.Request, err error) {
	if err == nil {
		return
	}

	ctx := r.Context()
	span := trace.SpanFromContext(ctx)

	// 1. Unwrap CustomError to get the original error
	originalErr := h.unwrapError(err)

	// 2. Map error to HTTP status + message
	mapping := h.mapper.MapError(originalErr)

	// 3. Add attributes to span for tracing
	span.RecordError(originalErr)
	span.SetStatus(codes.Error, mapping.Message)
	span.SetAttributes(
		attribute.Int("http.status_code", mapping.Status),
		attribute.String("error.type", fmt.Sprintf("%T", originalErr)),
		attribute.String("error.message", originalErr.Error()),
	)

	// 4. Log error once with appropriate level
	h.logError(ctx, r, originalErr, mapping.Status)

	// 5. Construct and write ProblemDetail response
	problem := NewProblemDetail(r, mapping.Status, getStatusText(mapping.Status), mapping.Message)

	// Add extra details from ValidationErrors
	var validationErrs validation.ValidationErrors
	if errors.As(originalErr, &validationErrs) {
		problem.Errors = h.convertValidationErrors(validationErrs)
	} else if customErr, ok := err.(*customerrors.CustomError); ok && customErr.Details != nil {
		// Add extra details from CustomError if present
		problem.Errors = customErr.Details
	}

	// Write ProblemDetail response
	if writeErr := WriteProblemDetail(w, problem); writeErr != nil {
		// Fallback: write simple error response
		http.Error(w, mapping.Message, mapping.Status)
	}
}

// unwrapError unwraps a CustomError to get the original error.
func (h *errorHandler) unwrapError(err error) error {
	var customErr *customerrors.CustomError
	if errors.As(err, &customErr) && customErr.Err != nil {
		return customErr.Err
	}
	return err
}

// logError logs the error with appropriate level based on HTTP status.
func (h *errorHandler) logError(ctx context.Context, r *http.Request, err error, status int) {
	fields := []observability.Field{
		observability.Error(err),
		observability.String("path", r.URL.Path),
		observability.String("method", r.Method),
		observability.Int("status", status),
	}

	if requestID := getRequestID(r); requestID != "" {
		fields = append(fields, observability.String("request_id", requestID))
	}

	if traceID := getTraceID(r); traceID != "" {
		fields = append(fields, observability.String("trace_id", traceID))
	}

	// Log with appropriate level based on status code
	switch {
	case status >= 500:
		// Server error → ERROR
		h.o11y.Logger().Error(ctx, "internal server error", fields...)
	case status == 404:
		// Not found → INFO (not a system error)
		h.o11y.Logger().Info(ctx, "resource not found", fields...)
	case status >= 400:
		// Client error → WARN
		h.o11y.Logger().Warn(ctx, "client error", fields...)
	default:
		// Other → INFO
		h.o11y.Logger().Info(ctx, "request error", fields...)
	}
}

// convertValidationErrors converts validation errors to error details format.
func (h *errorHandler) convertValidationErrors(errs validation.ValidationErrors) map[string]any {
	fields := make([]map[string]string, 0, len(errs))
	for _, err := range errs {
		fields = append(fields, map[string]string{
			"field":   err.Field,
			"message": err.Message,
		})
	}
	return map[string]any{
		"validation": fields,
	}
}
