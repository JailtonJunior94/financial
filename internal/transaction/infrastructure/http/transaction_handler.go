package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/responses"
	"github.com/go-chi/chi/v5"

	"github.com/jailtonjunior94/financial/internal/transaction/application/dtos"
	"github.com/jailtonjunior94/financial/internal/transaction/application/usecase"
	"github.com/jailtonjunior94/financial/pkg/api/httperrors"
	"github.com/jailtonjunior94/financial/pkg/api/middlewares"
	"github.com/jailtonjunior94/financial/pkg/pagination"
)

type TransactionHandler struct {
	o11y                         observability.Observability
	errorHandler                 httperrors.ErrorHandler
	registerTransactionUseCase   usecase.RegisterTransactionUseCase
	updateTransactionItemUseCase usecase.UpdateTransactionItemUseCase
	deleteTransactionItemUseCase usecase.DeleteTransactionItemUseCase
	listMonthlyPaginatedUseCase  usecase.ListMonthlyPaginatedUseCase
	getMonthlyUseCase            usecase.GetMonthlyUseCase
}

func NewTransactionHandler(
	o11y observability.Observability,
	errorHandler httperrors.ErrorHandler,
	registerTransactionUseCase usecase.RegisterTransactionUseCase,
	updateTransactionItemUseCase usecase.UpdateTransactionItemUseCase,
	deleteTransactionItemUseCase usecase.DeleteTransactionItemUseCase,
	listMonthlyPaginatedUseCase usecase.ListMonthlyPaginatedUseCase,
	getMonthlyUseCase usecase.GetMonthlyUseCase,
) *TransactionHandler {
	return &TransactionHandler{
		o11y:                         o11y,
		errorHandler:                 errorHandler,
		registerTransactionUseCase:   registerTransactionUseCase,
		updateTransactionItemUseCase: updateTransactionItemUseCase,
		deleteTransactionItemUseCase: deleteTransactionItemUseCase,
		listMonthlyPaginatedUseCase:  listMonthlyPaginatedUseCase,
		getMonthlyUseCase:            getMonthlyUseCase,
	}
}

func (h *TransactionHandler) Register(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "transaction_handler.register")
	defer span.End()

	user, err := middlewares.GetUserFromContext(ctx)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	var input *dtos.RegisterTransactionInput
	if err = json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	output, err := h.registerTransactionUseCase.Execute(ctx, user.ID, input)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	responses.JSON(w, http.StatusCreated, output)
}

func (h *TransactionHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "transaction_handler.list")
	defer span.End()

	user, err := middlewares.GetUserFromContext(ctx)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	// Parse cursor parameters (default: limit=20, max=100)
	params, err := pagination.ParseCursorParams(r, 20, 100)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	output, err := h.listMonthlyPaginatedUseCase.Execute(ctx, usecase.ListMonthlyPaginatedInput{
		UserID: user.ID,
		Limit:  params.Limit,
		Cursor: params.Cursor,
	})
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	// Build paginated response
	response := pagination.NewPaginatedResponse(output.MonthlyTransactions, params.Limit, output.NextCursor)
	responses.JSON(w, http.StatusOK, response)
}

func (h *TransactionHandler) Get(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "transaction_handler.get")
	defer span.End()

	user, err := middlewares.GetUserFromContext(ctx)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	monthlyID := chi.URLParam(r, "id")
	if monthlyID == "" {
		h.errorHandler.HandleError(w, r, fmt.Errorf("monthly_id is required"))
		return
	}

	output, err := h.getMonthlyUseCase.Execute(ctx, user.ID, monthlyID)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	responses.JSON(w, http.StatusOK, output)
}

func (h *TransactionHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "transaction_handler.update_item")
	defer span.End()

	user, err := middlewares.GetUserFromContext(ctx)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	// Change 7: Extract both transactionId and itemId from path
	transactionID := chi.URLParam(r, "transactionId")
	itemID := chi.URLParam(r, "itemId")

	if transactionID == "" || itemID == "" {
		h.errorHandler.HandleError(w, r, fmt.Errorf("transactionId and itemId are required"))
		return
	}

	var input *dtos.UpdateTransactionItemInput
	if err = json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	// Note: Current use case uses itemId only. The domain aggregate ensures item belongs to transaction.
	output, err := h.updateTransactionItemUseCase.Execute(ctx, user.ID, itemID, input)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	responses.JSON(w, http.StatusOK, output)
}

func (h *TransactionHandler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "transaction_handler.delete_item")
	defer span.End()

	user, err := middlewares.GetUserFromContext(ctx)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	// Change 7: Extract both transactionId and itemId from path
	transactionID := chi.URLParam(r, "transactionId")
	itemID := chi.URLParam(r, "itemId")

	if transactionID == "" || itemID == "" {
		h.errorHandler.HandleError(w, r, fmt.Errorf("transactionId and itemId are required"))
		return
	}

	// Note: Current use case uses itemId only. The domain aggregate ensures item belongs to transaction.
	_, err = h.deleteTransactionItemUseCase.Execute(ctx, user.ID, itemID)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	// Phase 2 Fix: DELETE should return 204 No Content with empty body
	responses.JSON(w, http.StatusNoContent, nil)
}
