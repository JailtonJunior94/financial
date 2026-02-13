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

// Register godoc
//
//	@Summary		Registrar transação
//	@Description	Cria uma nova transação mensal (ou adiciona um item a transação existente do mês).
//	@Description
//	@Description	**Enums:**
//	@Description	- `direction`: `INCOME` | `EXPENSE`
//	@Description	- `type`: `PIX` | `BOLETO` | `TRANSFER` | `CREDIT_CARD`
//	@Description
//	@Description	**Formato de campos:**
//	@Description	- `reference_month`: `YYYY-MM` (ex: `2025-01`)
//	@Description	- `amount`: string decimal (ex: `"1234.56"`)
//	@Tags			transactions
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dtos.RegisterTransactionInput	true	"Dados da transação"
//	@Success		201		{object}	dtos.MonthlyTransactionOutput	"Transação registrada"
//	@Failure		400		{object}	httperrors.ProblemDetail		"Dados inválidos"
//	@Failure		401		{object}	httperrors.ProblemDetail		"Não autenticado"
//	@Failure		404		{object}	httperrors.ProblemDetail		"Categoria não encontrada"
//	@Failure		500		{object}	httperrors.ProblemDetail		"Erro interno"
//	@Router			/api/v1/transactions [post]
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

	if validationErrs := input.Validate(); validationErrs.HasErrors() {
		h.errorHandler.HandleError(w, r, validationErrs)
		return
	}

	output, err := h.registerTransactionUseCase.Execute(ctx, user.ID, input)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	responses.JSON(w, http.StatusCreated, output)
}

// List godoc
//
//	@Summary		Listar transações mensais
//	@Description	Retorna a lista paginada de transações mensais consolidadas do usuário autenticado.
//	@Description	Cada item representa um mês com seus totais e itens individuais.
//	@Tags			transactions
//	@Produce		json
//	@Security		BearerAuth
//	@Param			limit	query		integer	false	"Itens por página (default: 20, max: 100)"	minimum(1)	maximum(100)	default(20)
//	@Param			cursor	query		string	false	"Cursor de paginação"
//	@Success		200		{object}	dtos.MonthlyTransactionPaginatedOutput	"Lista paginada de transações mensais"
//	@Failure		400		{object}	httperrors.ProblemDetail									"Parâmetro inválido"
//	@Failure		401		{object}	httperrors.ProblemDetail									"Não autenticado"
//	@Failure		500		{object}	httperrors.ProblemDetail									"Erro interno"
//	@Router			/api/v1/transactions [get]
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

// Get godoc
//
//	@Summary		Buscar transação mensal por ID
//	@Description	Retorna os detalhes de uma transação mensal específica, incluindo todos os itens e totais.
//	@Tags			transactions
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string							true	"ID da transação mensal"	format(uuid)
//	@Success		200	{object}	dtos.MonthlyTransactionOutput	"Dados da transação mensal"
//	@Failure		400	{object}	httperrors.ProblemDetail		"ID inválido"
//	@Failure		401	{object}	httperrors.ProblemDetail		"Não autenticado"
//	@Failure		404	{object}	httperrors.ProblemDetail		"Transação não encontrada"
//	@Failure		500	{object}	httperrors.ProblemDetail		"Erro interno"
//	@Router			/api/v1/transactions/{id} [get]
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

// UpdateItem godoc
//
//	@Summary		Atualizar item de transação
//	@Description	Atualiza os dados de um item individual dentro de uma transação mensal.
//	@Description	O `transactionId` é o ID do consolidado mensal; o `itemId` é o item específico.
//	@Description
//	@Description	**Enums:**
//	@Description	- `direction`: `INCOME` | `EXPENSE`
//	@Description	- `type`: `PIX` | `BOLETO` | `TRANSFER` | `CREDIT_CARD`
//	@Tags			transactions
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			transactionId	path		string							true	"ID da transação mensal"	format(uuid)
//	@Param			itemId			path		string							true	"ID do item de transação"	format(uuid)
//	@Param			request			body		dtos.UpdateTransactionItemInput	true	"Dados atualizados do item"
//	@Success		200				{object}	dtos.TransactionItemOutput		"Item atualizado"
//	@Failure		400				{object}	httperrors.ProblemDetail		"Dados inválidos"
//	@Failure		401				{object}	httperrors.ProblemDetail		"Não autenticado"
//	@Failure		404				{object}	httperrors.ProblemDetail		"Item não encontrado"
//	@Failure		500				{object}	httperrors.ProblemDetail		"Erro interno"
//	@Router			/api/v1/transactions/{transactionId}/items/{itemId} [put]
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

	if validationErrs := input.Validate(); validationErrs.HasErrors() {
		h.errorHandler.HandleError(w, r, validationErrs)
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

// DeleteItem godoc
//
//	@Summary		Remover item de transação
//	@Description	Remove um item individual de uma transação mensal. Se for o último item, a transação mensal também é removida.
//	@Tags			transactions
//	@Produce		json
//	@Security		BearerAuth
//	@Param			transactionId	path	string	true	"ID da transação mensal"	format(uuid)
//	@Param			itemId			path	string	true	"ID do item de transação"	format(uuid)
//	@Success		204	"Item removido com sucesso"
//	@Failure		400	{object}	httperrors.ProblemDetail	"IDs inválidos"
//	@Failure		401	{object}	httperrors.ProblemDetail	"Não autenticado"
//	@Failure		404	{object}	httperrors.ProblemDetail	"Item não encontrado"
//	@Failure		500	{object}	httperrors.ProblemDetail	"Erro interno"
//	@Router			/api/v1/transactions/{transactionId}/items/{itemId} [delete]
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
