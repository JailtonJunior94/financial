package http

import (
	"encoding/json"
	"net/http"

	"github.com/jailtonjunior94/financial/internal/payment_method/application/dtos"
	"github.com/jailtonjunior94/financial/internal/payment_method/application/usecase"
	"github.com/jailtonjunior94/financial/pkg/api/httperrors"
	"github.com/jailtonjunior94/financial/pkg/pagination"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/responses"

	"github.com/go-chi/chi/v5"
)

type PaymentMethodHandler struct {
	o11y                              observability.Observability
	errorHandler                      httperrors.ErrorHandler
	findPaymentMethodUseCase          usecase.FindPaymentMethodUseCase
	findPaymentMethodPaginatedUseCase usecase.FindPaymentMethodPaginatedUseCase
	createPaymentMethodUseCase        usecase.CreatePaymentMethodUseCase
	findPaymentMethodByUseCase        usecase.FindPaymentMethodByUseCase
	findPaymentMethodByCodeUseCase    usecase.FindPaymentMethodByCodeUseCase
	updatePaymentMethodUseCase        usecase.UpdatePaymentMethodUseCase
	removePaymentMethodUseCase        usecase.RemovePaymentMethodUseCase
}

func NewPaymentMethodHandler(
	o11y observability.Observability,
	errorHandler httperrors.ErrorHandler,
	findPaymentMethodUseCase usecase.FindPaymentMethodUseCase,
	findPaymentMethodPaginatedUseCase usecase.FindPaymentMethodPaginatedUseCase,
	createPaymentMethodUseCase usecase.CreatePaymentMethodUseCase,
	findPaymentMethodByUseCase usecase.FindPaymentMethodByUseCase,
	findPaymentMethodByCodeUseCase usecase.FindPaymentMethodByCodeUseCase,
	updatePaymentMethodUseCase usecase.UpdatePaymentMethodUseCase,
	removePaymentMethodUseCase usecase.RemovePaymentMethodUseCase,
) *PaymentMethodHandler {
	return &PaymentMethodHandler{
		o11y:                              o11y,
		errorHandler:                      errorHandler,
		findPaymentMethodUseCase:          findPaymentMethodUseCase,
		findPaymentMethodPaginatedUseCase: findPaymentMethodPaginatedUseCase,
		createPaymentMethodUseCase:        createPaymentMethodUseCase,
		updatePaymentMethodUseCase:        updatePaymentMethodUseCase,
		findPaymentMethodByUseCase:        findPaymentMethodByUseCase,
		findPaymentMethodByCodeUseCase:    findPaymentMethodByCodeUseCase,
		removePaymentMethodUseCase:        removePaymentMethodUseCase,
	}
}

func (h *PaymentMethodHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "payment_method_handler.create")
	defer span.End()

	var input *dtos.PaymentMethodInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	output, err := h.createPaymentMethodUseCase.Execute(ctx, input)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	responses.JSON(w, http.StatusCreated, output)
}

func (h *PaymentMethodHandler) Find(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "payment_method_handler.find")
	defer span.End()

	// Parse cursor parameters (default: limit=20, max=100)
	params, err := pagination.ParseCursorParams(r, 20, 100)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	// Get code query param if provided (Change 5: code filter via query param)
	code := r.URL.Query().Get("code")

	output, err := h.findPaymentMethodPaginatedUseCase.Execute(ctx, usecase.FindPaymentMethodPaginatedInput{
		Limit:  params.Limit,
		Cursor: params.Cursor,
		Code:   code,
	})
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	// Build paginated response
	response := pagination.NewPaginatedResponse(output.PaymentMethods, params.Limit, output.NextCursor)
	responses.JSON(w, http.StatusOK, response)
}

func (h *PaymentMethodHandler) FindBy(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "payment_method_handler.find_by")
	defer span.End()

	output, err := h.findPaymentMethodByUseCase.Execute(ctx, chi.URLParam(r, "id"))
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	responses.JSON(w, http.StatusOK, output)
}

func (h *PaymentMethodHandler) FindByCode(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "payment_method_handler.find_by_code")
	defer span.End()

	output, err := h.findPaymentMethodByCodeUseCase.Execute(ctx, chi.URLParam(r, "code"))
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	responses.JSON(w, http.StatusOK, output)
}

func (h *PaymentMethodHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "payment_method_handler.update")
	defer span.End()

	var input *dtos.PaymentMethodUpdateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	output, err := h.updatePaymentMethodUseCase.Execute(ctx, chi.URLParam(r, "id"), input)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	responses.JSON(w, http.StatusOK, output)
}

func (h *PaymentMethodHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "payment_method_handler.delete")
	defer span.End()

	if err := h.removePaymentMethodUseCase.Execute(ctx, chi.URLParam(r, "id")); err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	responses.JSON(w, http.StatusNoContent, nil)
}
