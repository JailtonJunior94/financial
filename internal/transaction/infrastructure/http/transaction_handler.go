package http

import (
	"encoding/json"
	"net/http"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/responses"
	"github.com/go-chi/chi/v5"

	"github.com/jailtonjunior94/financial/internal/transaction/application/dtos"
	"github.com/jailtonjunior94/financial/internal/transaction/application/usecase"
	"github.com/jailtonjunior94/financial/pkg/api/httperrors"
	"github.com/jailtonjunior94/financial/pkg/api/middlewares"
)

type TransactionHandler struct {
	o11y                         observability.Observability
	errorHandler                 httperrors.ErrorHandler
	registerTransactionUseCase   usecase.RegisterTransactionUseCase
	updateTransactionItemUseCase usecase.UpdateTransactionItemUseCase
	deleteTransactionItemUseCase usecase.DeleteTransactionItemUseCase
}

func NewTransactionHandler(
	o11y observability.Observability,
	errorHandler httperrors.ErrorHandler,
	registerTransactionUseCase usecase.RegisterTransactionUseCase,
	updateTransactionItemUseCase usecase.UpdateTransactionItemUseCase,
	deleteTransactionItemUseCase usecase.DeleteTransactionItemUseCase,
) *TransactionHandler {
	return &TransactionHandler{
		o11y:                         o11y,
		errorHandler:                 errorHandler,
		registerTransactionUseCase:   registerTransactionUseCase,
		updateTransactionItemUseCase: updateTransactionItemUseCase,
		deleteTransactionItemUseCase: deleteTransactionItemUseCase,
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

func (h *TransactionHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "transaction_handler.update_item")
	defer span.End()

	user, err := middlewares.GetUserFromContext(ctx)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	var input *dtos.UpdateTransactionItemInput
	if err = json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	output, err := h.updateTransactionItemUseCase.Execute(ctx, user.ID, chi.URLParam(r, "id"), input)
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

	output, err := h.deleteTransactionItemUseCase.Execute(ctx, user.ID, chi.URLParam(r, "id"))
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	responses.JSON(w, http.StatusOK, output)
}
