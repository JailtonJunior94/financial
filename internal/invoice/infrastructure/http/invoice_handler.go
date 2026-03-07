package http

import (
	"net/http"
	"strconv"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/responses"
	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel/trace"

	"github.com/jailtonjunior94/financial/internal/invoice/application/usecase"
	"github.com/jailtonjunior94/financial/pkg/api/httperrors"
	"github.com/jailtonjunior94/financial/pkg/api/middlewares"
	"github.com/jailtonjunior94/financial/pkg/pagination"
)

// InvoiceHandler handles HTTP requests for the invoice resource.
type InvoiceHandler struct {
	o11y         observability.Observability
	errorHandler httperrors.ErrorHandler
	listByCardUC usecase.ListInvoicesByCardPaginatedUseCase
	getByCardUC  usecase.GetInvoiceUseCase
}

// NewInvoiceHandler creates a new InvoiceHandler.
func NewInvoiceHandler(
	o11y observability.Observability,
	errorHandler httperrors.ErrorHandler,
	listByCardUC usecase.ListInvoicesByCardPaginatedUseCase,
	getByCardUC usecase.GetInvoiceUseCase,
) *InvoiceHandler {
	return &InvoiceHandler{
		o11y:         o11y,
		errorHandler: errorHandler,
		listByCardUC: listByCardUC,
		getByCardUC:  getByCardUC,
	}
}

// ListByCard godoc
//
//	@Summary		List invoices for a card
//	@Tags			invoices
//	@Produce		json
//	@Security		BearerAuth
//	@Param			cardId	path		string	true	"Card ID"	format(uuid)
//	@Param			status	query		string	false	"Filter by status (open, closed, paid)"
//	@Param			limit	query		int		false	"Limit (default 20, max 100)"
//	@Param			cursor	query		string	false	"Pagination cursor"
//	@Success		200	{object}	pagination.PaginatedResponse
//	@Failure		401	{object}	httperrors.ProblemDetail
//	@Failure		500	{object}	httperrors.ProblemDetail
//	@Router			/api/v1/cards/{cardId}/invoices [get]
func (h *InvoiceHandler) ListByCard(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "invoice_handler.list_by_card")
	defer span.End()
	correlationID := trace.SpanFromContext(ctx).SpanContext().TraceID().String()
	user, err := middlewares.GetUserFromContext(ctx)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}
	h.o11y.Logger().Info(ctx, "request_received",
		observability.String("operation", "list_by_card"),
		observability.String("layer", "handler"),
		observability.String("entity", "invoice"),
		observability.String("correlation_id", correlationID),
		observability.String("user_id", user.ID),
	)
	cardID := chi.URLParam(r, "cardId")
	status := r.URL.Query().Get("status")
	limit := parseLimit(r.URL.Query().Get("limit"))
	cursor := r.URL.Query().Get("cursor")
	output, err := h.listByCardUC.Execute(ctx, usecase.ListInvoicesByCardPaginatedInput{
		UserID: user.ID,
		CardID: cardID,
		Status: status,
		Limit:  limit,
		Cursor: cursor,
	})
	if err != nil {
		span.RecordError(err)
		h.o11y.Logger().Error(ctx, "request_failed",
			observability.String("operation", "list_by_card"),
			observability.String("layer", "handler"),
			observability.String("entity", "invoice"),
			observability.String("correlation_id", correlationID),
			observability.String("user_id", user.ID),
			observability.Error(err),
		)
		h.errorHandler.HandleError(w, r, err)
		return
	}
	h.o11y.Logger().Info(ctx, "request_completed",
		observability.String("operation", "list_by_card"),
		observability.String("layer", "handler"),
		observability.String("entity", "invoice"),
		observability.String("correlation_id", correlationID),
		observability.String("user_id", user.ID),
	)
	response := pagination.NewPaginatedResponse(output.Invoices, limit, output.NextCursor)
	responses.JSON(w, http.StatusOK, response)
}

// GetByCard godoc
//
//	@Summary		Get invoice details
//	@Tags			invoices
//	@Produce		json
//	@Security		BearerAuth
//	@Param			cardId		path		string	true	"Card ID"		format(uuid)
//	@Param			invoiceId	path		string	true	"Invoice ID"	format(uuid)
//	@Success		200	{object}	dtos.InvoiceOutput
//	@Failure		401	{object}	httperrors.ProblemDetail
//	@Failure		403	{object}	httperrors.ProblemDetail
//	@Failure		404	{object}	httperrors.ProblemDetail
//	@Failure		500	{object}	httperrors.ProblemDetail
//	@Router			/api/v1/cards/{cardId}/invoices/{invoiceId} [get]
func (h *InvoiceHandler) GetByCard(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "invoice_handler.get_by_card")
	defer span.End()
	correlationID := trace.SpanFromContext(ctx).SpanContext().TraceID().String()
	user, err := middlewares.GetUserFromContext(ctx)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}
	h.o11y.Logger().Info(ctx, "request_received",
		observability.String("operation", "get_by_card"),
		observability.String("layer", "handler"),
		observability.String("entity", "invoice"),
		observability.String("correlation_id", correlationID),
		observability.String("user_id", user.ID),
	)
	cardID := chi.URLParam(r, "cardId")
	invoiceID := chi.URLParam(r, "invoiceId")
	output, err := h.getByCardUC.Execute(ctx, user.ID, cardID, invoiceID)
	if err != nil {
		span.RecordError(err)
		h.o11y.Logger().Error(ctx, "request_failed",
			observability.String("operation", "get_by_card"),
			observability.String("layer", "handler"),
			observability.String("entity", "invoice"),
			observability.String("correlation_id", correlationID),
			observability.String("user_id", user.ID),
			observability.Error(err),
		)
		h.errorHandler.HandleError(w, r, err)
		return
	}
	h.o11y.Logger().Info(ctx, "request_completed",
		observability.String("operation", "get_by_card"),
		observability.String("layer", "handler"),
		observability.String("entity", "invoice"),
		observability.String("correlation_id", correlationID),
		observability.String("user_id", user.ID),
	)
	responses.JSON(w, http.StatusOK, output)
}

const (
	maxInvoiceHandlerLimit     = 100
	defaultInvoiceHandlerLimit = 20
)

func parseLimit(raw string) int {
	if raw == "" {
		return defaultInvoiceHandlerLimit
	}
	parsed, err := strconv.Atoi(raw)
	if err != nil || parsed <= 0 {
		return defaultInvoiceHandlerLimit
	}
	if parsed > maxInvoiceHandlerLimit {
		return maxInvoiceHandlerLimit
	}
	return parsed
}
