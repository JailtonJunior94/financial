package http

import (
	"net/http"
	"time"

	"github.com/jailtonjunior94/financial/internal/user/application/dtos"
	"github.com/jailtonjunior94/financial/internal/user/application/usecase"

	"github.com/JailtonJunior94/devkit-go/pkg/o11y"
	"github.com/JailtonJunior94/devkit-go/pkg/responses"

	"go.opentelemetry.io/otel/metric"
)

type AuthHandler struct {
	metrics      authMetrics
	o11y         o11y.Observability
	tokenUseCase usecase.TokenUseCase
}

type authMetrics struct {
	requestCounter  metric.Int64Counter
	requestDuration metric.Float64Histogram
}

func NewAuthHandler(
	o11y o11y.Observability,
	tokenUseCase usecase.TokenUseCase,
) *AuthHandler {
	meter := o11y.MeterProvider().Meter("financial")

	counter, err := meter.Int64Counter("auth.request.counter", metric.WithDescription("HTTP Requests Counter (Authentication)"))
	if err != nil {
		return nil
	}

	duration, err := meter.Float64Histogram("auth.request.duration", metric.WithDescription("HTTP Request Duration Histogram (Authentication)"))
	if err != nil {
		return nil
	}

	return &AuthHandler{
		o11y:         o11y,
		tokenUseCase: tokenUseCase,
		metrics: authMetrics{
			requestCounter:  counter,
			requestDuration: duration,
		},
	}
}

func (h *AuthHandler) Token(w http.ResponseWriter, r *http.Request) error {
	ctx, span := h.o11y.Start(r.Context(), "auth_handler.token")
	defer span.End()

	start := time.Now()
	h.metrics.requestCounter.Add(ctx, 1)

	input := &dtos.AuthInput{
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	}

	output, err := h.tokenUseCase.Execute(ctx, input)
	if err != nil {
		span.RecordError(err)
		return err
	}

	responses.JSON(w, http.StatusOK, output)
	h.metrics.requestDuration.Record(ctx, float64(time.Since(start).Nanoseconds()))
	span.AddAttributes(ctx, o11y.Ok, "authentication successful",
		o11y.Attributes{Key: "email", Value: input.Email},
		o11y.Attributes{Key: "duration", Value: time.Since(start).String()},
	)

	return nil
}
