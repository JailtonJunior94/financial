package httperrors

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/trace"
)

// ProblemDetail represents an RFC 7807 Problem Details for HTTP APIs response
// See: https://tools.ietf.org/html/rfc7807
type ProblemDetail struct {
	Type      string                 `json:"type"`
	Title     string                 `json:"title"`
	Status    int                    `json:"status"`
	Detail    string                 `json:"detail"`
	Instance  string                 `json:"instance"`
	Timestamp time.Time              `json:"timestamp"`
	RequestID string                 `json:"request_id,omitempty"`
	TraceID   string                 `json:"trace_id,omitempty"`
	Errors    map[string]any `json:"errors,omitempty"`
}

// NewProblemDetail creates a new ProblemDetail from an HTTP request and error details.
func NewProblemDetail(r *http.Request, status int, title, detail string) *ProblemDetail {
	requestID := getRequestID(r)
	traceID := getTraceID(r)

	return &ProblemDetail{
		Type:      fmt.Sprintf("https://httpstatuses.com/%d", status),
		Title:     title,
		Status:    status,
		Detail:    detail,
		Instance:  r.URL.Path,
		Timestamp: time.Now().UTC(),
		RequestID: requestID,
		TraceID:   traceID,
	}
}

// WriteProblemDetail writes a ProblemDetail as JSON response.
func WriteProblemDetail(w http.ResponseWriter, problem *ProblemDetail) error {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(problem.Status)
	return json.NewEncoder(w).Encode(problem)
}

// getRequestID extracts the request ID from the HTTP request.
func getRequestID(r *http.Request) string {
	if requestID := r.Header.Get("X-Request-ID"); requestID != "" {
		return requestID
	}
	if requestID, ok := r.Context().Value("request_id").(string); ok {
		return requestID
	}
	return ""
}

// getTraceID extracts the trace ID from the request context.
func getTraceID(r *http.Request) string {
	spanContext := trace.SpanContextFromContext(r.Context())
	if spanContext.IsValid() {
		return spanContext.TraceID().String()
	}
	return ""
}

// getStatusText returns the standard HTTP status text for a given code.
func getStatusText(code int) string {
	text := http.StatusText(code)
	if text == "" {
		return "Unknown Error"
	}
	return text
}
