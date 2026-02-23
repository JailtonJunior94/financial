package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jailtonjunior94/financial/internal/budget/application/dtos"
	"github.com/jailtonjunior94/financial/internal/budget/application/usecase"
	"github.com/jailtonjunior94/financial/pkg/api/httperrors"
	"github.com/jailtonjunior94/financial/pkg/api/middlewares"
	"github.com/jailtonjunior94/financial/pkg/pagination"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/responses"
	"go.opentelemetry.io/otel/trace"
)

type BudgetHandler struct {
	o11y                        observability.Observability
	errorHandler                httperrors.ErrorHandler
	createBudgetUseCase         usecase.CreateBudgetUseCase
	findBudgetUseCase           usecase.FindBudgetUseCase
	updateBudgetUseCase         usecase.UpdateBudgetUseCase
	deleteBudgetUseCase         usecase.DeleteBudgetUseCase
	listBudgetsPaginatedUseCase usecase.ListBudgetsPaginatedUseCase
}

func NewBudgetHandler(
	o11y observability.Observability,
	errorHandler httperrors.ErrorHandler,
	createBudgetUseCase usecase.CreateBudgetUseCase,
	findBudgetUseCase usecase.FindBudgetUseCase,
	updateBudgetUseCase usecase.UpdateBudgetUseCase,
	deleteBudgetUseCase usecase.DeleteBudgetUseCase,
	listBudgetsPaginatedUseCase usecase.ListBudgetsPaginatedUseCase,
) *BudgetHandler {
	return &BudgetHandler{
		o11y:                        o11y,
		errorHandler:                errorHandler,
		createBudgetUseCase:         createBudgetUseCase,
		findBudgetUseCase:           findBudgetUseCase,
		updateBudgetUseCase:         updateBudgetUseCase,
		deleteBudgetUseCase:         deleteBudgetUseCase,
		listBudgetsPaginatedUseCase: listBudgetsPaginatedUseCase,
	}
}

// Create godoc
//
//	@Summary		Criar orçamento mensal
//	@Description	Cria um novo orçamento para o mês de referência do usuário autenticado.
//	@Description	Apenas um orçamento por mês (`reference_month`) é permitido por usuário.
//	@Description
//	@Description	**Campos:**
//	@Description	- `reference_month`: `YYYY-MM` (ex: `2025-01`)
//	@Description	- `total_amount`: valor total planejado (ex: `"5000.00"`)
//	@Description	- `currency`: `BRL` | `USD` | `EUR` (opcional, default: `BRL`)
//	@Description	- `items`: ao menos um item com `category_id` e `percentage_goal` (ex: `"25.50"`)
//	@Tags			budgets
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dtos.BudgetCreateInput		true	"Dados do orçamento"
//	@Success		201		{object}	dtos.BudgetOutput			"Orçamento criado"
//	@Failure		400		{object}	httperrors.ProblemDetail	"Dados inválidos"
//	@Failure		401		{object}	httperrors.ProblemDetail	"Não autenticado"
//	@Failure		409		{object}	httperrors.ProblemDetail	"Orçamento já existe para este mês"
//	@Failure		500		{object}	httperrors.ProblemDetail	"Erro interno"
//	@Router			/api/v1/budgets [post]
func (h *BudgetHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "budget_handler.create")
	defer span.End()

	correlationID := trace.SpanFromContext(ctx).SpanContext().TraceID().String()

	user, err := middlewares.GetUserFromContext(ctx)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	h.o11y.Logger().Info(ctx, "request_received",
		observability.String("operation", "CreateBudget"),
		observability.String("layer", "handler"),
		observability.String("entity", "budget"),
		observability.String("correlation_id", correlationID),
		observability.String("user_id", user.ID),
	)

	var input *dtos.BudgetCreateInput
	if err = json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.o11y.Logger().Error(ctx, "validation_failed",
			observability.String("operation", "CreateBudget"),
			observability.String("layer", "handler"),
			observability.String("entity", "budget"),
			observability.String("correlation_id", correlationID),
			observability.String("user_id", user.ID),
			observability.String("error_type", "validation"),
			observability.String("error_code", "DECODE_BODY_FAILED"),
			observability.Error(err),
		)
		h.errorHandler.HandleError(w, r, err)
		return
	}

	if validationErrs := input.Validate(); validationErrs.HasErrors() {
		h.o11y.Logger().Warn(ctx, "validation_failed",
			observability.String("operation", "CreateBudget"),
			observability.String("layer", "handler"),
			observability.String("entity", "budget"),
			observability.String("correlation_id", correlationID),
			observability.String("user_id", user.ID),
			observability.String("error_type", "validation"),
			observability.String("error_code", "INPUT_VALIDATION_FAILED"),
		)
		h.errorHandler.HandleError(w, r, validationErrs)
		return
	}

	output, err := h.createBudgetUseCase.Execute(ctx, user.ID, input)
	if err != nil {
		h.o11y.Logger().Error(ctx, "request_failed",
			observability.String("operation", "CreateBudget"),
			observability.String("layer", "handler"),
			observability.String("entity", "budget"),
			observability.String("correlation_id", correlationID),
			observability.String("user_id", user.ID),
			observability.String("error_type", "business"),
			observability.String("error_code", "CREATE_BUDGET_FAILED"),
			observability.Error(err),
		)
		h.errorHandler.HandleError(w, r, err)
		return
	}

	h.o11y.Logger().Info(ctx, "request_completed",
		observability.String("operation", "CreateBudget"),
		observability.String("layer", "handler"),
		observability.String("entity", "budget"),
		observability.String("correlation_id", correlationID),
		observability.String("user_id", user.ID),
		observability.String("budget_id", output.ID),
	)

	responses.JSON(w, http.StatusCreated, output)
}

// List godoc
//
//	@Summary		Listar orçamentos mensais
//	@Description	Retorna a lista paginada de orçamentos do usuário autenticado (cursor-based pagination).
//	@Tags			budgets
//	@Produce		json
//	@Security		BearerAuth
//	@Param			limit	query		integer	false	"Itens por página (default: 20, max: 100)"	minimum(1)	maximum(100)	default(20)
//	@Param			cursor	query		string	false	"Cursor de paginação"
//	@Success		200		{object}	dtos.BudgetPaginatedOutput	"Lista paginada de orçamentos"
//	@Failure		400		{object}	httperrors.ProblemDetail						"Parâmetro inválido"
//	@Failure		401		{object}	httperrors.ProblemDetail						"Não autenticado"
//	@Failure		500		{object}	httperrors.ProblemDetail						"Erro interno"
//	@Router			/api/v1/budgets [get]
func (h *BudgetHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "budget_handler.list")
	defer span.End()

	correlationID := trace.SpanFromContext(ctx).SpanContext().TraceID().String()

	user, err := middlewares.GetUserFromContext(ctx)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	h.o11y.Logger().Info(ctx, "request_received",
		observability.String("operation", "ListBudgets"),
		observability.String("layer", "handler"),
		observability.String("entity", "budget"),
		observability.String("correlation_id", correlationID),
		observability.String("user_id", user.ID),
	)

	params, err := pagination.ParseCursorParams(r, 20, 100)
	if err != nil {
		h.o11y.Logger().Error(ctx, "request_failed",
			observability.String("operation", "ListBudgets"),
			observability.String("layer", "handler"),
			observability.String("entity", "budget"),
			observability.String("correlation_id", correlationID),
			observability.String("user_id", user.ID),
			observability.String("error_type", "validation"),
			observability.String("error_code", "PAGINATION_PARAMS_INVALID"),
			observability.Error(err),
		)
		h.errorHandler.HandleError(w, r, err)
		return
	}

	output, err := h.listBudgetsPaginatedUseCase.Execute(ctx, usecase.ListBudgetsPaginatedInput{
		UserID: user.ID,
		Limit:  params.Limit,
		Cursor: params.Cursor,
	})
	if err != nil {
		h.o11y.Logger().Error(ctx, "request_failed",
			observability.String("operation", "ListBudgets"),
			observability.String("layer", "handler"),
			observability.String("entity", "budget"),
			observability.String("correlation_id", correlationID),
			observability.String("user_id", user.ID),
			observability.String("error_type", "business"),
			observability.String("error_code", "LIST_BUDGETS_FAILED"),
			observability.Error(err),
		)
		h.errorHandler.HandleError(w, r, err)
		return
	}

	h.o11y.Logger().Info(ctx, "request_completed",
		observability.String("operation", "ListBudgets"),
		observability.String("layer", "handler"),
		observability.String("entity", "budget"),
		observability.String("correlation_id", correlationID),
		observability.String("user_id", user.ID),
	)

	response := pagination.NewPaginatedResponse(output.Budgets, params.Limit, output.NextCursor)
	responses.JSON(w, http.StatusOK, response)
}

// Find godoc
//
//	@Summary		Buscar orçamento por ID
//	@Description	Retorna os detalhes de um orçamento mensal específico, incluindo itens com valores planejados, gastos e percentuais.
//	@Tags			budgets
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string					true	"ID do orçamento"	format(uuid)
//	@Success		200	{object}	dtos.BudgetOutput		"Dados do orçamento"
//	@Failure		400	{object}	httperrors.ProblemDetail	"ID inválido"
//	@Failure		401	{object}	httperrors.ProblemDetail	"Não autenticado"
//	@Failure		404	{object}	httperrors.ProblemDetail	"Orçamento não encontrado"
//	@Failure		500	{object}	httperrors.ProblemDetail	"Erro interno"
//	@Router			/api/v1/budgets/{id} [get]
func (h *BudgetHandler) Find(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "budget_handler.find")
	defer span.End()

	correlationID := trace.SpanFromContext(ctx).SpanContext().TraceID().String()

	user, err := middlewares.GetUserFromContext(ctx)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	budgetID := r.PathValue("id")
	if budgetID == "" {
		h.o11y.Logger().Warn(ctx, "validation_failed",
			observability.String("operation", "FindBudget"),
			observability.String("layer", "handler"),
			observability.String("entity", "budget"),
			observability.String("correlation_id", correlationID),
			observability.String("user_id", user.ID),
			observability.String("error_type", "validation"),
			observability.String("error_code", "BUDGET_ID_REQUIRED"),
		)
		h.errorHandler.HandleError(w, r, fmt.Errorf("budget_id is required"))
		return
	}

	h.o11y.Logger().Info(ctx, "request_received",
		observability.String("operation", "FindBudget"),
		observability.String("layer", "handler"),
		observability.String("entity", "budget"),
		observability.String("correlation_id", correlationID),
		observability.String("user_id", user.ID),
		observability.String("budget_id", budgetID),
	)

	output, err := h.findBudgetUseCase.Execute(ctx, user.ID, budgetID)
	if err != nil {
		h.o11y.Logger().Error(ctx, "request_failed",
			observability.String("operation", "FindBudget"),
			observability.String("layer", "handler"),
			observability.String("entity", "budget"),
			observability.String("correlation_id", correlationID),
			observability.String("user_id", user.ID),
			observability.String("budget_id", budgetID),
			observability.String("error_type", "business"),
			observability.String("error_code", "FIND_BUDGET_FAILED"),
			observability.Error(err),
		)
		h.errorHandler.HandleError(w, r, err)
		return
	}

	h.o11y.Logger().Info(ctx, "request_completed",
		observability.String("operation", "FindBudget"),
		observability.String("layer", "handler"),
		observability.String("entity", "budget"),
		observability.String("correlation_id", correlationID),
		observability.String("user_id", user.ID),
		observability.String("budget_id", budgetID),
	)

	responses.JSON(w, http.StatusOK, output)
}

// Update godoc
//
//	@Summary		Atualizar orçamento mensal
//	@Description	Atualiza o valor total e os itens de um orçamento existente.
//	@Description	Os itens existentes são substituídos pelos novos itens enviados.
//	@Tags			budgets
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string					true	"ID do orçamento"	format(uuid)
//	@Param			request	body		dtos.BudgetUpdateInput	true	"Dados atualizados"
//	@Success		200		{object}	dtos.BudgetOutput		"Orçamento atualizado"
//	@Failure		400		{object}	httperrors.ProblemDetail	"Dados inválidos"
//	@Failure		401		{object}	httperrors.ProblemDetail	"Não autenticado"
//	@Failure		404		{object}	httperrors.ProblemDetail	"Orçamento não encontrado"
//	@Failure		500		{object}	httperrors.ProblemDetail	"Erro interno"
//	@Router			/api/v1/budgets/{id} [put]
func (h *BudgetHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "budget_handler.update")
	defer span.End()

	correlationID := trace.SpanFromContext(ctx).SpanContext().TraceID().String()

	user, err := middlewares.GetUserFromContext(ctx)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	budgetID := r.PathValue("id")
	if budgetID == "" {
		h.errorHandler.HandleError(w, r, fmt.Errorf("budget_id is required"))
		return
	}

	h.o11y.Logger().Info(ctx, "request_received",
		observability.String("operation", "UpdateBudget"),
		observability.String("layer", "handler"),
		observability.String("entity", "budget"),
		observability.String("correlation_id", correlationID),
		observability.String("user_id", user.ID),
		observability.String("budget_id", budgetID),
	)

	var input *dtos.BudgetUpdateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.o11y.Logger().Error(ctx, "validation_failed",
			observability.String("operation", "UpdateBudget"),
			observability.String("layer", "handler"),
			observability.String("entity", "budget"),
			observability.String("correlation_id", correlationID),
			observability.String("user_id", user.ID),
			observability.String("error_type", "validation"),
			observability.String("error_code", "DECODE_BODY_FAILED"),
			observability.Error(err),
		)
		h.errorHandler.HandleError(w, r, err)
		return
	}

	if validationErrs := input.Validate(); validationErrs.HasErrors() {
		h.o11y.Logger().Warn(ctx, "validation_failed",
			observability.String("operation", "UpdateBudget"),
			observability.String("layer", "handler"),
			observability.String("entity", "budget"),
			observability.String("correlation_id", correlationID),
			observability.String("user_id", user.ID),
			observability.String("error_type", "validation"),
			observability.String("error_code", "INPUT_VALIDATION_FAILED"),
		)
		h.errorHandler.HandleError(w, r, validationErrs)
		return
	}

	output, err := h.updateBudgetUseCase.Execute(ctx, user.ID, budgetID, input)
	if err != nil {
		h.o11y.Logger().Error(ctx, "request_failed",
			observability.String("operation", "UpdateBudget"),
			observability.String("layer", "handler"),
			observability.String("entity", "budget"),
			observability.String("correlation_id", correlationID),
			observability.String("user_id", user.ID),
			observability.String("budget_id", budgetID),
			observability.String("error_type", "business"),
			observability.String("error_code", "UPDATE_BUDGET_FAILED"),
			observability.Error(err),
		)
		h.errorHandler.HandleError(w, r, err)
		return
	}

	h.o11y.Logger().Info(ctx, "request_completed",
		observability.String("operation", "UpdateBudget"),
		observability.String("layer", "handler"),
		observability.String("entity", "budget"),
		observability.String("correlation_id", correlationID),
		observability.String("user_id", user.ID),
		observability.String("budget_id", budgetID),
	)

	responses.JSON(w, http.StatusOK, output)
}

// Delete godoc
//
//	@Summary		Remover orçamento mensal
//	@Description	Remove um orçamento mensal e todos os seus itens. Esta operação é irreversível.
//	@Tags			budgets
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	string	true	"ID do orçamento"	format(uuid)
//	@Success		204	"Orçamento removido com sucesso"
//	@Failure		400	{object}	httperrors.ProblemDetail	"ID inválido"
//	@Failure		401	{object}	httperrors.ProblemDetail	"Não autenticado"
//	@Failure		404	{object}	httperrors.ProblemDetail	"Orçamento não encontrado"
//	@Failure		500	{object}	httperrors.ProblemDetail	"Erro interno"
//	@Router			/api/v1/budgets/{id} [delete]
func (h *BudgetHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "budget_handler.delete")
	defer span.End()

	correlationID := trace.SpanFromContext(ctx).SpanContext().TraceID().String()

	user, err := middlewares.GetUserFromContext(ctx)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	budgetID := r.PathValue("id")
	if budgetID == "" {
		h.errorHandler.HandleError(w, r, fmt.Errorf("budget_id is required"))
		return
	}

	h.o11y.Logger().Info(ctx, "request_received",
		observability.String("operation", "DeleteBudget"),
		observability.String("layer", "handler"),
		observability.String("entity", "budget"),
		observability.String("correlation_id", correlationID),
		observability.String("user_id", user.ID),
		observability.String("budget_id", budgetID),
	)

	err = h.deleteBudgetUseCase.Execute(ctx, user.ID, budgetID)
	if err != nil {
		h.o11y.Logger().Error(ctx, "request_failed",
			observability.String("operation", "DeleteBudget"),
			observability.String("layer", "handler"),
			observability.String("entity", "budget"),
			observability.String("correlation_id", correlationID),
			observability.String("user_id", user.ID),
			observability.String("budget_id", budgetID),
			observability.String("error_type", "business"),
			observability.String("error_code", "DELETE_BUDGET_FAILED"),
			observability.Error(err),
		)
		h.errorHandler.HandleError(w, r, err)
		return
	}

	h.o11y.Logger().Info(ctx, "request_completed",
		observability.String("operation", "DeleteBudget"),
		observability.String("layer", "handler"),
		observability.String("entity", "budget"),
		observability.String("correlation_id", correlationID),
		observability.String("user_id", user.ID),
		observability.String("budget_id", budgetID),
	)

	responses.JSON(w, http.StatusNoContent, nil)
}
