package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/responses"
	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel/trace"

	"github.com/jailtonjunior94/financial/internal/transaction/application/dtos"
	"github.com/jailtonjunior94/financial/internal/transaction/application/usecase"
	"github.com/jailtonjunior94/financial/pkg/api/httperrors"
	"github.com/jailtonjunior94/financial/pkg/api/middlewares"
)

// TransactionHandler handles HTTP requests for the transaction resource.
type TransactionHandler struct {
	o11y         observability.Observability
	errorHandler httperrors.ErrorHandler
	createUC     usecase.CreateTransactionUseCase
	updateUC     usecase.UpdateTransactionUseCase
	reverseUC    usecase.ReverseTransactionUseCase
	listUC       usecase.ListTransactionsUseCase
	getUC        usecase.GetTransactionUseCase
}

// NewTransactionHandler creates a new TransactionHandler.
func NewTransactionHandler(
	o11y observability.Observability,
	errorHandler httperrors.ErrorHandler,
	createUC usecase.CreateTransactionUseCase,
	updateUC usecase.UpdateTransactionUseCase,
	reverseUC usecase.ReverseTransactionUseCase,
	listUC usecase.ListTransactionsUseCase,
	getUC usecase.GetTransactionUseCase,
) *TransactionHandler {
	return &TransactionHandler{
		o11y:         o11y,
		errorHandler: errorHandler,
		createUC:     createUC,
		updateUC:     updateUC,
		reverseUC:    reverseUC,
		listUC:       listUC,
		getUC:        getUC,
	}
}

func (h *TransactionHandler) logInfo(ctx context.Context, event, operation, correlationID, userID string) {
	h.o11y.Logger().Info(ctx, event,
		observability.String("operation", operation),
		observability.String("layer", "handler"),
		observability.String("entity", "transaction"),
		observability.String("correlation_id", correlationID),
		observability.String("user_id", userID),
	)
}

func (h *TransactionHandler) logError(ctx context.Context, operation, correlationID, userID string, err error) {
	h.o11y.Logger().Error(ctx, "request_failed",
		observability.String("operation", operation),
		observability.String("layer", "handler"),
		observability.String("entity", "transaction"),
		observability.String("correlation_id", correlationID),
		observability.String("user_id", userID),
		observability.Error(err),
	)
}

const (
	maxTransactionLimit     = 100
	defaultTransactionLimit = 20
)

// Create godoc
//
//	@Summary		Create a new transaction
//	@Tags			transactions
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dtos.TransactionInput	true	"Transaction input"
//	@Success		201		{array}		dtos.TransactionOutput
//	@Failure		400		{object}	httperrors.ProblemDetail
//	@Failure		401		{object}	httperrors.ProblemDetail
//	@Failure		422		{object}	httperrors.ProblemDetail
//	@Failure		500		{object}	httperrors.ProblemDetail
//	@Router			/api/v1/transactions [post]
func (h *TransactionHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "transaction_handler.create")
	defer span.End()
	correlationID := trace.SpanFromContext(ctx).SpanContext().TraceID().String()
	user, err := middlewares.GetUserFromContext(ctx)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}
	h.logInfo(ctx, "request_received", "create_transaction", correlationID, user.ID)
	var input dtos.TransactionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}
	outputs, err := h.createUC.Execute(ctx, user.ID, &input)
	if err != nil {
		span.RecordError(err)
		h.logError(ctx, "create_transaction", correlationID, user.ID, err)
		h.errorHandler.HandleError(w, r, err)
		return
	}
	h.logInfo(ctx, "request_completed", "create_transaction", correlationID, user.ID)
	responses.JSON(w, http.StatusCreated, outputs)
}

// List godoc
//
//	@Summary		List transactions with pagination
//	@Tags			transactions
//	@Produce		json
//	@Security		BearerAuth
//	@Param			payment_method	query	string	false	"Filter by payment method"
//	@Param			category_id		query	string	false	"Filter by category ID"
//	@Param			start_date		query	string	false	"Start date (YYYY-MM-DD)"
//	@Param			end_date		query	string	false	"End date (YYYY-MM-DD)"
//	@Param			limit			query	int		false	"Page size (default 20, max 100)"
//	@Param			cursor			query	string	false	"Pagination cursor"
//	@Success		200	{object}	dtos.TransactionListOutput
//	@Failure		401	{object}	httperrors.ProblemDetail
//	@Failure		500	{object}	httperrors.ProblemDetail
//	@Router			/api/v1/transactions [get]
func (h *TransactionHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "transaction_handler.list")
	defer span.End()
	correlationID := trace.SpanFromContext(ctx).SpanContext().TraceID().String()
	user, err := middlewares.GetUserFromContext(ctx)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}
	h.logInfo(ctx, "request_received", "list_transactions", correlationID, user.ID)
	params := &dtos.ListParams{
		PaymentMethod: r.URL.Query().Get("payment_method"),
		CategoryID:    r.URL.Query().Get("category_id"),
		StartDate:     r.URL.Query().Get("start_date"),
		EndDate:       r.URL.Query().Get("end_date"),
		Limit:         parseTransactionLimit(r.URL.Query().Get("limit")),
		Cursor:        r.URL.Query().Get("cursor"),
	}
	output, err := h.listUC.Execute(ctx, user.ID, params)
	if err != nil {
		span.RecordError(err)
		h.logError(ctx, "list_transactions", correlationID, user.ID, err)
		h.errorHandler.HandleError(w, r, err)
		return
	}
	h.logInfo(ctx, "request_completed", "list_transactions", correlationID, user.ID)
	responses.JSON(w, http.StatusOK, output)
}

// Get godoc
//
//	@Summary		Get a transaction by ID
//	@Tags			transactions
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Transaction ID"	format(uuid)
//	@Success		200	{object}	dtos.TransactionOutput
//	@Failure		401	{object}	httperrors.ProblemDetail
//	@Failure		403	{object}	httperrors.ProblemDetail
//	@Failure		404	{object}	httperrors.ProblemDetail
//	@Failure		500	{object}	httperrors.ProblemDetail
//	@Router			/api/v1/transactions/{id} [get]
func (h *TransactionHandler) Get(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "transaction_handler.get")
	defer span.End()
	correlationID := trace.SpanFromContext(ctx).SpanContext().TraceID().String()
	user, err := middlewares.GetUserFromContext(ctx)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}
	transactionID := chi.URLParam(r, "id")
	h.logInfo(ctx, "request_received", "get_transaction", correlationID, user.ID)
	output, err := h.getUC.Execute(ctx, user.ID, transactionID)
	if err != nil {
		span.RecordError(err)
		h.logError(ctx, "get_transaction", correlationID, user.ID, err)
		h.errorHandler.HandleError(w, r, err)
		return
	}
	h.logInfo(ctx, "request_completed", "get_transaction", correlationID, user.ID)
	responses.JSON(w, http.StatusOK, output)
}

// Update godoc
//
//	@Summary		Update a transaction
//	@Tags			transactions
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string						true	"Transaction ID"	format(uuid)
//	@Param			request	body		dtos.TransactionUpdateInput	true	"Update input"
//	@Success		200		{object}	dtos.TransactionOutput
//	@Failure		400		{object}	httperrors.ProblemDetail
//	@Failure		401		{object}	httperrors.ProblemDetail
//	@Failure		403		{object}	httperrors.ProblemDetail
//	@Failure		404		{object}	httperrors.ProblemDetail
//	@Failure		422		{object}	httperrors.ProblemDetail
//	@Failure		500		{object}	httperrors.ProblemDetail
//	@Router			/api/v1/transactions/{id} [put]
func (h *TransactionHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "transaction_handler.update")
	defer span.End()
	correlationID := trace.SpanFromContext(ctx).SpanContext().TraceID().String()
	user, err := middlewares.GetUserFromContext(ctx)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}
	transactionID := chi.URLParam(r, "id")
	h.logInfo(ctx, "request_received", "update_transaction", correlationID, user.ID)
	var input dtos.TransactionUpdateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}
	output, err := h.updateUC.Execute(ctx, user.ID, transactionID, &input)
	if err != nil {
		span.RecordError(err)
		h.logError(ctx, "update_transaction", correlationID, user.ID, err)
		h.errorHandler.HandleError(w, r, err)
		return
	}
	h.logInfo(ctx, "request_completed", "update_transaction", correlationID, user.ID)
	responses.JSON(w, http.StatusOK, output)
}

// Reverse godoc
//
//	@Summary		Reverse a transaction (cancel installments in open invoices)
//	@Tags			transactions
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Transaction ID"	format(uuid)
//	@Success		200	{object}	dtos.ReverseOutput
//	@Failure		401	{object}	httperrors.ProblemDetail
//	@Failure		403	{object}	httperrors.ProblemDetail
//	@Failure		404	{object}	httperrors.ProblemDetail
//	@Failure		409	{object}	httperrors.ProblemDetail
//	@Failure		500	{object}	httperrors.ProblemDetail
//	@Router			/api/v1/transactions/{id}/reverse [post]
func (h *TransactionHandler) Reverse(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "transaction_handler.reverse")
	defer span.End()
	correlationID := trace.SpanFromContext(ctx).SpanContext().TraceID().String()
	user, err := middlewares.GetUserFromContext(ctx)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}
	transactionID := chi.URLParam(r, "id")
	h.logInfo(ctx, "request_received", "reverse_transaction", correlationID, user.ID)
	output, err := h.reverseUC.Execute(ctx, user.ID, transactionID)
	if err != nil {
		span.RecordError(err)
		h.logError(ctx, "reverse_transaction", correlationID, user.ID, err)
		h.errorHandler.HandleError(w, r, err)
		return
	}
	h.logInfo(ctx, "request_completed", "reverse_transaction", correlationID, user.ID)
	responses.JSON(w, http.StatusOK, output)
}

func parseTransactionLimit(raw string) int {
	if raw == "" {
		return defaultTransactionLimit
	}
	parsed, err := strconv.Atoi(raw)
	if err != nil || parsed <= 0 {
		return defaultTransactionLimit
	}
	if parsed > maxTransactionLimit {
		return maxTransactionLimit
	}
	return parsed
}
