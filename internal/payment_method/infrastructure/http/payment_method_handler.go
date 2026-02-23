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
	"go.opentelemetry.io/otel/trace"

	"github.com/go-chi/chi/v5"
)

type PaymentMethodHandler struct {
	o11y                              observability.Observability
	errorHandler                      httperrors.ErrorHandler
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
		findPaymentMethodPaginatedUseCase: findPaymentMethodPaginatedUseCase,
		createPaymentMethodUseCase:        createPaymentMethodUseCase,
		updatePaymentMethodUseCase:        updatePaymentMethodUseCase,
		findPaymentMethodByUseCase:        findPaymentMethodByUseCase,
		findPaymentMethodByCodeUseCase:    findPaymentMethodByCodeUseCase,
		removePaymentMethodUseCase:        removePaymentMethodUseCase,
	}
}

// Create godoc
//
//	@Summary		Criar método de pagamento
//	@Description	Cria um novo método de pagamento. Este endpoint é público (não requer autenticação).
//	@Description	O campo `code` deve ser único e imutável após criação (ex: `PIX`, `BOLETO`, `CREDIT_CARD`).
//	@Tags			payment-methods
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dtos.PaymentMethodInput		true	"Dados do método de pagamento"
//	@Success		201		{object}	dtos.PaymentMethodOutput	"Método de pagamento criado"
//	@Failure		400		{object}	httperrors.ProblemDetail	"Dados inválidos"
//	@Failure		409		{object}	httperrors.ProblemDetail	"Código já existente"
//	@Failure		500		{object}	httperrors.ProblemDetail	"Erro interno"
//	@Router			/api/v1/payment-methods [post]
func (h *PaymentMethodHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "payment_method_handler.create")
	defer span.End()

	correlationID := trace.SpanFromContext(ctx).SpanContext().TraceID().String()

	h.o11y.Logger().Info(ctx, "request_received",
		observability.String("operation", "CreatePaymentMethod"),
		observability.String("layer", "handler"),
		observability.String("entity", "payment_method"),
		observability.String("correlation_id", correlationID),
	)

	var input *dtos.PaymentMethodInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.o11y.Logger().Error(ctx, "validation_failed",
			observability.String("operation", "CreatePaymentMethod"),
			observability.String("layer", "handler"),
			observability.String("entity", "payment_method"),
			observability.String("correlation_id", correlationID),
			observability.String("error_type", "validation"),
			observability.String("error_code", "DECODE_BODY_FAILED"),
			observability.Error(err),
		)
		h.errorHandler.HandleError(w, r, err)
		return
	}

	output, err := h.createPaymentMethodUseCase.Execute(ctx, input)
	if err != nil {
		h.o11y.Logger().Error(ctx, "request_failed",
			observability.String("operation", "CreatePaymentMethod"),
			observability.String("layer", "handler"),
			observability.String("entity", "payment_method"),
			observability.String("correlation_id", correlationID),
			observability.String("error_type", "business"),
			observability.String("error_code", "CREATE_PAYMENT_METHOD_FAILED"),
			observability.Error(err),
		)
		h.errorHandler.HandleError(w, r, err)
		return
	}

	h.o11y.Logger().Info(ctx, "request_completed",
		observability.String("operation", "CreatePaymentMethod"),
		observability.String("layer", "handler"),
		observability.String("entity", "payment_method"),
		observability.String("correlation_id", correlationID),
		observability.String("payment_method_id", output.ID),
	)

	responses.JSON(w, http.StatusCreated, output)
}

// Find godoc
//
//	@Summary		Listar métodos de pagamento
//	@Description	Retorna lista paginada de métodos de pagamento. Endpoint público.
//	@Description	Filtragem opcional por `code` via query param (ex: `?code=PIX`).
//	@Tags			payment-methods
//	@Produce		json
//	@Param			limit	query		integer	false	"Itens por página (default: 20, max: 100)"	minimum(1)	maximum(100)	default(20)
//	@Param			cursor	query		string	false	"Cursor de paginação"
//	@Param			code	query		string	false	"Filtrar pelo código do método"				example(PIX)
//	@Success		200		{object}	dtos.PaymentMethodPaginatedOutput	"Lista paginada"
//	@Failure		400		{object}	httperrors.ProblemDetail							"Parâmetro inválido"
//	@Failure		500		{object}	httperrors.ProblemDetail							"Erro interno"
//	@Router			/api/v1/payment-methods [get]
func (h *PaymentMethodHandler) Find(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "payment_method_handler.find")
	defer span.End()

	correlationID := trace.SpanFromContext(ctx).SpanContext().TraceID().String()

	h.o11y.Logger().Info(ctx, "request_received",
		observability.String("operation", "FindPaymentMethods"),
		observability.String("layer", "handler"),
		observability.String("entity", "payment_method"),
		observability.String("correlation_id", correlationID),
	)

	params, err := pagination.ParseCursorParams(r, 20, 100)
	if err != nil {
		h.o11y.Logger().Error(ctx, "request_failed",
			observability.String("operation", "FindPaymentMethods"),
			observability.String("layer", "handler"),
			observability.String("entity", "payment_method"),
			observability.String("correlation_id", correlationID),
			observability.String("error_type", "validation"),
			observability.String("error_code", "PAGINATION_PARAMS_INVALID"),
			observability.Error(err),
		)
		h.errorHandler.HandleError(w, r, err)
		return
	}

	code := r.URL.Query().Get("code")

	output, err := h.findPaymentMethodPaginatedUseCase.Execute(ctx, usecase.FindPaymentMethodPaginatedInput{
		Limit:  params.Limit,
		Cursor: params.Cursor,
		Code:   code,
	})
	if err != nil {
		h.o11y.Logger().Error(ctx, "request_failed",
			observability.String("operation", "FindPaymentMethods"),
			observability.String("layer", "handler"),
			observability.String("entity", "payment_method"),
			observability.String("correlation_id", correlationID),
			observability.String("error_type", "business"),
			observability.String("error_code", "FIND_PAYMENT_METHODS_FAILED"),
			observability.Error(err),
		)
		h.errorHandler.HandleError(w, r, err)
		return
	}

	h.o11y.Logger().Info(ctx, "request_completed",
		observability.String("operation", "FindPaymentMethods"),
		observability.String("layer", "handler"),
		observability.String("entity", "payment_method"),
		observability.String("correlation_id", correlationID),
	)

	response := pagination.NewPaginatedResponse(output.PaymentMethods, params.Limit, output.NextCursor)
	responses.JSON(w, http.StatusOK, response)
}

// FindBy godoc
//
//	@Summary		Buscar método de pagamento por ID
//	@Description	Retorna os detalhes de um método de pagamento pelo UUID. Endpoint público.
//	@Tags			payment-methods
//	@Produce		json
//	@Param			id	path		string						true	"ID do método de pagamento"	format(uuid)
//	@Success		200	{object}	dtos.PaymentMethodOutput	"Dados do método de pagamento"
//	@Failure		404	{object}	httperrors.ProblemDetail	"Não encontrado"
//	@Failure		500	{object}	httperrors.ProblemDetail	"Erro interno"
//	@Router			/api/v1/payment-methods/{id} [get]
func (h *PaymentMethodHandler) FindBy(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "payment_method_handler.find_by")
	defer span.End()

	correlationID := trace.SpanFromContext(ctx).SpanContext().TraceID().String()
	pmID := chi.URLParam(r, "id")

	h.o11y.Logger().Info(ctx, "request_received",
		observability.String("operation", "FindPaymentMethodBy"),
		observability.String("layer", "handler"),
		observability.String("entity", "payment_method"),
		observability.String("correlation_id", correlationID),
		observability.String("payment_method_id", pmID),
	)

	output, err := h.findPaymentMethodByUseCase.Execute(ctx, pmID)
	if err != nil {
		h.o11y.Logger().Error(ctx, "request_failed",
			observability.String("operation", "FindPaymentMethodBy"),
			observability.String("layer", "handler"),
			observability.String("entity", "payment_method"),
			observability.String("correlation_id", correlationID),
			observability.String("payment_method_id", pmID),
			observability.String("error_type", "business"),
			observability.String("error_code", "FIND_PAYMENT_METHOD_FAILED"),
			observability.Error(err),
		)
		h.errorHandler.HandleError(w, r, err)
		return
	}

	h.o11y.Logger().Info(ctx, "request_completed",
		observability.String("operation", "FindPaymentMethodBy"),
		observability.String("layer", "handler"),
		observability.String("entity", "payment_method"),
		observability.String("correlation_id", correlationID),
		observability.String("payment_method_id", pmID),
	)

	responses.JSON(w, http.StatusOK, output)
}

func (h *PaymentMethodHandler) FindByCode(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "payment_method_handler.find_by_code")
	defer span.End()

	correlationID := trace.SpanFromContext(ctx).SpanContext().TraceID().String()
	code := chi.URLParam(r, "code")

	h.o11y.Logger().Info(ctx, "request_received",
		observability.String("operation", "FindPaymentMethodByCode"),
		observability.String("layer", "handler"),
		observability.String("entity", "payment_method"),
		observability.String("correlation_id", correlationID),
		observability.String("code", code),
	)

	output, err := h.findPaymentMethodByCodeUseCase.Execute(ctx, code)
	if err != nil {
		h.o11y.Logger().Error(ctx, "request_failed",
			observability.String("operation", "FindPaymentMethodByCode"),
			observability.String("layer", "handler"),
			observability.String("entity", "payment_method"),
			observability.String("correlation_id", correlationID),
			observability.String("code", code),
			observability.String("error_type", "business"),
			observability.String("error_code", "FIND_PAYMENT_METHOD_BY_CODE_FAILED"),
			observability.Error(err),
		)
		h.errorHandler.HandleError(w, r, err)
		return
	}

	h.o11y.Logger().Info(ctx, "request_completed",
		observability.String("operation", "FindPaymentMethodByCode"),
		observability.String("layer", "handler"),
		observability.String("entity", "payment_method"),
		observability.String("correlation_id", correlationID),
		observability.String("code", code),
	)

	responses.JSON(w, http.StatusOK, output)
}

// Update godoc
//
//	@Summary		Atualizar método de pagamento
//	@Description	Atualiza nome e descrição de um método de pagamento. O campo `code` não pode ser alterado. Endpoint público.
//	@Tags			payment-methods
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string							true	"ID do método de pagamento"	format(uuid)
//	@Param			request	body		dtos.PaymentMethodUpdateInput	true	"Dados atualizados"
//	@Success		200		{object}	dtos.PaymentMethodOutput		"Método atualizado"
//	@Failure		400		{object}	httperrors.ProblemDetail		"Dados inválidos"
//	@Failure		404		{object}	httperrors.ProblemDetail		"Não encontrado"
//	@Failure		500		{object}	httperrors.ProblemDetail		"Erro interno"
//	@Router			/api/v1/payment-methods/{id} [put]
func (h *PaymentMethodHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "payment_method_handler.update")
	defer span.End()

	correlationID := trace.SpanFromContext(ctx).SpanContext().TraceID().String()
	pmID := chi.URLParam(r, "id")

	h.o11y.Logger().Info(ctx, "request_received",
		observability.String("operation", "UpdatePaymentMethod"),
		observability.String("layer", "handler"),
		observability.String("entity", "payment_method"),
		observability.String("correlation_id", correlationID),
		observability.String("payment_method_id", pmID),
	)

	var input *dtos.PaymentMethodUpdateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.o11y.Logger().Error(ctx, "validation_failed",
			observability.String("operation", "UpdatePaymentMethod"),
			observability.String("layer", "handler"),
			observability.String("entity", "payment_method"),
			observability.String("correlation_id", correlationID),
			observability.String("error_type", "validation"),
			observability.String("error_code", "DECODE_BODY_FAILED"),
			observability.Error(err),
		)
		h.errorHandler.HandleError(w, r, err)
		return
	}

	output, err := h.updatePaymentMethodUseCase.Execute(ctx, pmID, input)
	if err != nil {
		h.o11y.Logger().Error(ctx, "request_failed",
			observability.String("operation", "UpdatePaymentMethod"),
			observability.String("layer", "handler"),
			observability.String("entity", "payment_method"),
			observability.String("correlation_id", correlationID),
			observability.String("payment_method_id", pmID),
			observability.String("error_type", "business"),
			observability.String("error_code", "UPDATE_PAYMENT_METHOD_FAILED"),
			observability.Error(err),
		)
		h.errorHandler.HandleError(w, r, err)
		return
	}

	h.o11y.Logger().Info(ctx, "request_completed",
		observability.String("operation", "UpdatePaymentMethod"),
		observability.String("layer", "handler"),
		observability.String("entity", "payment_method"),
		observability.String("correlation_id", correlationID),
		observability.String("payment_method_id", pmID),
	)

	responses.JSON(w, http.StatusOK, output)
}

// Delete godoc
//
//	@Summary		Remover método de pagamento
//	@Description	Remove um método de pagamento. Endpoint público.
//	@Tags			payment-methods
//	@Produce		json
//	@Param			id	path	string	true	"ID do método de pagamento"	format(uuid)
//	@Success		204	"Removido com sucesso"
//	@Failure		404	{object}	httperrors.ProblemDetail	"Não encontrado"
//	@Failure		500	{object}	httperrors.ProblemDetail	"Erro interno"
//	@Router			/api/v1/payment-methods/{id} [delete]
func (h *PaymentMethodHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "payment_method_handler.delete")
	defer span.End()

	correlationID := trace.SpanFromContext(ctx).SpanContext().TraceID().String()
	pmID := chi.URLParam(r, "id")

	h.o11y.Logger().Info(ctx, "request_received",
		observability.String("operation", "DeletePaymentMethod"),
		observability.String("layer", "handler"),
		observability.String("entity", "payment_method"),
		observability.String("correlation_id", correlationID),
		observability.String("payment_method_id", pmID),
	)

	if err := h.removePaymentMethodUseCase.Execute(ctx, pmID); err != nil {
		h.o11y.Logger().Error(ctx, "request_failed",
			observability.String("operation", "DeletePaymentMethod"),
			observability.String("layer", "handler"),
			observability.String("entity", "payment_method"),
			observability.String("correlation_id", correlationID),
			observability.String("payment_method_id", pmID),
			observability.String("error_type", "business"),
			observability.String("error_code", "DELETE_PAYMENT_METHOD_FAILED"),
			observability.Error(err),
		)
		h.errorHandler.HandleError(w, r, err)
		return
	}

	h.o11y.Logger().Info(ctx, "request_completed",
		observability.String("operation", "DeletePaymentMethod"),
		observability.String("layer", "handler"),
		observability.String("entity", "payment_method"),
		observability.String("correlation_id", correlationID),
		observability.String("payment_method_id", pmID),
	)

	responses.JSON(w, http.StatusNoContent, nil)
}
