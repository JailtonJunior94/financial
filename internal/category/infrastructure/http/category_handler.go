package http

import (
	"encoding/json"
	"net/http"

	"github.com/jailtonjunior94/financial/internal/category/application/dtos"
	"github.com/jailtonjunior94/financial/internal/category/application/usecase"
	"github.com/jailtonjunior94/financial/pkg/api/httperrors"
	"github.com/jailtonjunior94/financial/pkg/api/middlewares"
	"github.com/jailtonjunior94/financial/pkg/pagination"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/responses"
	"go.opentelemetry.io/otel/trace"

	"github.com/go-chi/chi/v5"
)

type CategoryHandler struct {
	o11y                         observability.Observability
	errorHandler                 httperrors.ErrorHandler
	findCategoryPaginatedUseCase usecase.FindCategoryPaginatedUseCase
	createCategoryUseCase        usecase.CreateCategoryUseCase
	findCategoryByUseCase        usecase.FindCategoryByUseCase
	updateCategoryUseCase        usecase.UpdateCategoryUseCase
	removeCategoryUseCase        usecase.RemoveCategoryUseCase
}

func NewCategoryHandler(
	o11y observability.Observability,
	errorHandler httperrors.ErrorHandler,
	findCategoryPaginatedUseCase usecase.FindCategoryPaginatedUseCase,
	createCategoryUseCase usecase.CreateCategoryUseCase,
	findCategoryByUseCase usecase.FindCategoryByUseCase,
	updateCategoryUseCase usecase.UpdateCategoryUseCase,
	removeCategoryUseCase usecase.RemoveCategoryUseCase,
) *CategoryHandler {
	return &CategoryHandler{
		o11y:                         o11y,
		errorHandler:                 errorHandler,
		findCategoryPaginatedUseCase: findCategoryPaginatedUseCase,
		createCategoryUseCase:        createCategoryUseCase,
		updateCategoryUseCase:        updateCategoryUseCase,
		findCategoryByUseCase:        findCategoryByUseCase,
		removeCategoryUseCase:        removeCategoryUseCase,
	}
}

// Create godoc
//
//	@Summary		Criar categoria
//	@Description	Cria uma nova categoria para o usuário autenticado.
//	@Description	Categorias podem ser hierárquicas: informe `parent_id` para criar uma subcategoria.
//	@Description	- `parent_id`: UUID da categoria pai (opcional)
//	@Description	- `sequence`: ordem de exibição (mínimo 1)
//	@Tags			categories
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dtos.CategoryInput			true	"Dados da categoria"
//	@Success		201		{object}	dtos.CategoryOutput			"Categoria criada com sucesso"
//	@Failure		400		{object}	httperrors.ProblemDetail	"Dados inválidos"
//	@Failure		401		{object}	httperrors.ProblemDetail	"Não autenticado"
//	@Failure		409		{object}	httperrors.ProblemDetail	"Ciclo detectado na hierarquia"
//	@Failure		500		{object}	httperrors.ProblemDetail	"Erro interno"
//	@Router			/api/v1/categories [post]
func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "category_handler.create")
	defer span.End()

	correlationID := trace.SpanFromContext(ctx).SpanContext().TraceID().String()

	user, err := middlewares.GetUserFromContext(ctx)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	h.o11y.Logger().Info(ctx, "request_received",
		observability.String("operation", "CreateCategory"),
		observability.String("layer", "handler"),
		observability.String("entity", "category"),
		observability.String("correlation_id", correlationID),
		observability.String("user_id", user.ID),
	)

	var input *dtos.CategoryInput
	if err = json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.o11y.Logger().Error(ctx, "validation_failed",
			observability.String("operation", "CreateCategory"),
			observability.String("layer", "handler"),
			observability.String("entity", "category"),
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
			observability.String("operation", "CreateCategory"),
			observability.String("layer", "handler"),
			observability.String("entity", "category"),
			observability.String("correlation_id", correlationID),
			observability.String("user_id", user.ID),
			observability.String("error_type", "validation"),
			observability.String("error_code", "INPUT_VALIDATION_FAILED"),
		)
		h.errorHandler.HandleError(w, r, validationErrs)
		return
	}

	output, err := h.createCategoryUseCase.Execute(ctx, user.ID, input)
	if err != nil {
		h.o11y.Logger().Error(ctx, "request_failed",
			observability.String("operation", "CreateCategory"),
			observability.String("layer", "handler"),
			observability.String("entity", "category"),
			observability.String("correlation_id", correlationID),
			observability.String("user_id", user.ID),
			observability.String("error_type", "business"),
			observability.String("error_code", "CREATE_CATEGORY_FAILED"),
			observability.Error(err),
		)
		h.errorHandler.HandleError(w, r, err)
		return
	}

	h.o11y.Logger().Info(ctx, "request_completed",
		observability.String("operation", "CreateCategory"),
		observability.String("layer", "handler"),
		observability.String("entity", "category"),
		observability.String("correlation_id", correlationID),
		observability.String("user_id", user.ID),
		observability.String("category_id", output.ID),
	)

	responses.JSON(w, http.StatusCreated, output)
}

// Find godoc
//
//	@Summary		Listar categorias
//	@Description	Retorna a lista paginada de categorias do usuário autenticado (cursor-based pagination).
//	@Tags			categories
//	@Produce		json
//	@Security		BearerAuth
//	@Param			limit	query		integer	false	"Itens por página (default: 20, max: 100)"	minimum(1)	maximum(100)	default(20)
//	@Param			cursor	query		string	false	"Cursor de paginação"
//	@Success		200		{object}	dtos.CategoryPaginatedOutput	"Lista paginada de categorias"
//	@Failure		400		{object}	httperrors.ProblemDetail						"Parâmetro inválido"
//	@Failure		401		{object}	httperrors.ProblemDetail						"Não autenticado"
//	@Failure		500		{object}	httperrors.ProblemDetail						"Erro interno"
//	@Router			/api/v1/categories [get]
func (h *CategoryHandler) Find(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "category_handler.find")
	defer span.End()

	correlationID := trace.SpanFromContext(ctx).SpanContext().TraceID().String()

	user, err := middlewares.GetUserFromContext(ctx)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	h.o11y.Logger().Info(ctx, "request_received",
		observability.String("operation", "FindCategories"),
		observability.String("layer", "handler"),
		observability.String("entity", "category"),
		observability.String("correlation_id", correlationID),
		observability.String("user_id", user.ID),
	)

	params, err := pagination.ParseCursorParams(r, 20, 100)
	if err != nil {
		h.o11y.Logger().Error(ctx, "request_failed",
			observability.String("operation", "FindCategories"),
			observability.String("layer", "handler"),
			observability.String("entity", "category"),
			observability.String("correlation_id", correlationID),
			observability.String("user_id", user.ID),
			observability.String("error_type", "validation"),
			observability.String("error_code", "PAGINATION_PARAMS_INVALID"),
			observability.Error(err),
		)
		h.errorHandler.HandleError(w, r, err)
		return
	}

	output, err := h.findCategoryPaginatedUseCase.Execute(ctx, usecase.FindCategoryPaginatedInput{
		UserID: user.ID,
		Limit:  params.Limit,
		Cursor: params.Cursor,
	})
	if err != nil {
		h.o11y.Logger().Error(ctx, "request_failed",
			observability.String("operation", "FindCategories"),
			observability.String("layer", "handler"),
			observability.String("entity", "category"),
			observability.String("correlation_id", correlationID),
			observability.String("user_id", user.ID),
			observability.String("error_type", "business"),
			observability.String("error_code", "FIND_CATEGORIES_FAILED"),
			observability.Error(err),
		)
		h.errorHandler.HandleError(w, r, err)
		return
	}

	h.o11y.Logger().Info(ctx, "request_completed",
		observability.String("operation", "FindCategories"),
		observability.String("layer", "handler"),
		observability.String("entity", "category"),
		observability.String("correlation_id", correlationID),
		observability.String("user_id", user.ID),
	)

	response := pagination.NewPaginatedResponse(output.Categories, params.Limit, output.NextCursor)
	responses.JSON(w, http.StatusOK, response)
}

// FindBy godoc
//
//	@Summary		Buscar categoria por ID
//	@Description	Retorna os detalhes de uma categoria específica, incluindo subcategorias (`children`).
//	@Tags			categories
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string					true	"ID da categoria"	format(uuid)
//	@Success		200	{object}	dtos.CategoryOutput		"Dados da categoria com filhos"
//	@Failure		401	{object}	httperrors.ProblemDetail	"Não autenticado"
//	@Failure		404	{object}	httperrors.ProblemDetail	"Categoria não encontrada"
//	@Failure		500	{object}	httperrors.ProblemDetail	"Erro interno"
//	@Router			/api/v1/categories/{id} [get]
func (h *CategoryHandler) FindBy(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "category_handler.find_by")
	defer span.End()

	correlationID := trace.SpanFromContext(ctx).SpanContext().TraceID().String()

	user, err := middlewares.GetUserFromContext(ctx)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	categoryID := chi.URLParam(r, "id")

	h.o11y.Logger().Info(ctx, "request_received",
		observability.String("operation", "FindCategoryBy"),
		observability.String("layer", "handler"),
		observability.String("entity", "category"),
		observability.String("correlation_id", correlationID),
		observability.String("user_id", user.ID),
		observability.String("category_id", categoryID),
	)

	output, err := h.findCategoryByUseCase.Execute(ctx, user.ID, categoryID)
	if err != nil {
		h.o11y.Logger().Error(ctx, "request_failed",
			observability.String("operation", "FindCategoryBy"),
			observability.String("layer", "handler"),
			observability.String("entity", "category"),
			observability.String("correlation_id", correlationID),
			observability.String("user_id", user.ID),
			observability.String("category_id", categoryID),
			observability.String("error_type", "business"),
			observability.String("error_code", "FIND_CATEGORY_FAILED"),
			observability.Error(err),
		)
		h.errorHandler.HandleError(w, r, err)
		return
	}

	h.o11y.Logger().Info(ctx, "request_completed",
		observability.String("operation", "FindCategoryBy"),
		observability.String("layer", "handler"),
		observability.String("entity", "category"),
		observability.String("correlation_id", correlationID),
		observability.String("user_id", user.ID),
		observability.String("category_id", categoryID),
	)

	responses.JSON(w, http.StatusOK, output)
}

// Update godoc
//
//	@Summary		Atualizar categoria
//	@Description	Atualiza os dados de uma categoria existente do usuário autenticado.
//	@Tags			categories
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string					true	"ID da categoria"	format(uuid)
//	@Param			request	body		dtos.CategoryInput		true	"Dados atualizados da categoria"
//	@Success		200		{object}	dtos.CategoryOutput		"Categoria atualizada"
//	@Failure		400		{object}	httperrors.ProblemDetail	"Dados inválidos"
//	@Failure		401		{object}	httperrors.ProblemDetail	"Não autenticado"
//	@Failure		404		{object}	httperrors.ProblemDetail	"Categoria não encontrada"
//	@Failure		409		{object}	httperrors.ProblemDetail	"Ciclo detectado na hierarquia"
//	@Failure		500		{object}	httperrors.ProblemDetail	"Erro interno"
//	@Router			/api/v1/categories/{id} [put]
func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "category_handler.update")
	defer span.End()

	correlationID := trace.SpanFromContext(ctx).SpanContext().TraceID().String()

	user, err := middlewares.GetUserFromContext(ctx)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	categoryID := chi.URLParam(r, "id")

	h.o11y.Logger().Info(ctx, "request_received",
		observability.String("operation", "UpdateCategory"),
		observability.String("layer", "handler"),
		observability.String("entity", "category"),
		observability.String("correlation_id", correlationID),
		observability.String("user_id", user.ID),
		observability.String("category_id", categoryID),
	)

	var input *dtos.CategoryInput
	if err = json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.o11y.Logger().Error(ctx, "validation_failed",
			observability.String("operation", "UpdateCategory"),
			observability.String("layer", "handler"),
			observability.String("entity", "category"),
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
			observability.String("operation", "UpdateCategory"),
			observability.String("layer", "handler"),
			observability.String("entity", "category"),
			observability.String("correlation_id", correlationID),
			observability.String("user_id", user.ID),
			observability.String("error_type", "validation"),
			observability.String("error_code", "INPUT_VALIDATION_FAILED"),
		)
		h.errorHandler.HandleError(w, r, validationErrs)
		return
	}

	output, err := h.updateCategoryUseCase.Execute(ctx, user.ID, categoryID, input)
	if err != nil {
		h.o11y.Logger().Error(ctx, "request_failed",
			observability.String("operation", "UpdateCategory"),
			observability.String("layer", "handler"),
			observability.String("entity", "category"),
			observability.String("correlation_id", correlationID),
			observability.String("user_id", user.ID),
			observability.String("category_id", categoryID),
			observability.String("error_type", "business"),
			observability.String("error_code", "UPDATE_CATEGORY_FAILED"),
			observability.Error(err),
		)
		h.errorHandler.HandleError(w, r, err)
		return
	}

	h.o11y.Logger().Info(ctx, "request_completed",
		observability.String("operation", "UpdateCategory"),
		observability.String("layer", "handler"),
		observability.String("entity", "category"),
		observability.String("correlation_id", correlationID),
		observability.String("user_id", user.ID),
		observability.String("category_id", categoryID),
	)

	responses.JSON(w, http.StatusOK, output)
}

// Delete godoc
//
//	@Summary		Remover categoria
//	@Description	Remove uma categoria do usuário autenticado. Esta operação é irreversível.
//	@Tags			categories
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	string	true	"ID da categoria"	format(uuid)
//	@Success		204	"Categoria removida com sucesso"
//	@Failure		401	{object}	httperrors.ProblemDetail	"Não autenticado"
//	@Failure		404	{object}	httperrors.ProblemDetail	"Categoria não encontrada"
//	@Failure		500	{object}	httperrors.ProblemDetail	"Erro interno"
//	@Router			/api/v1/categories/{id} [delete]
func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "category_handler.delete")
	defer span.End()

	correlationID := trace.SpanFromContext(ctx).SpanContext().TraceID().String()

	user, err := middlewares.GetUserFromContext(ctx)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	categoryID := chi.URLParam(r, "id")

	h.o11y.Logger().Info(ctx, "request_received",
		observability.String("operation", "DeleteCategory"),
		observability.String("layer", "handler"),
		observability.String("entity", "category"),
		observability.String("correlation_id", correlationID),
		observability.String("user_id", user.ID),
		observability.String("category_id", categoryID),
	)

	if err := h.removeCategoryUseCase.Execute(ctx, user.ID, categoryID); err != nil {
		h.o11y.Logger().Error(ctx, "request_failed",
			observability.String("operation", "DeleteCategory"),
			observability.String("layer", "handler"),
			observability.String("entity", "category"),
			observability.String("correlation_id", correlationID),
			observability.String("user_id", user.ID),
			observability.String("category_id", categoryID),
			observability.String("error_type", "business"),
			observability.String("error_code", "DELETE_CATEGORY_FAILED"),
			observability.Error(err),
		)
		h.errorHandler.HandleError(w, r, err)
		return
	}

	h.o11y.Logger().Info(ctx, "request_completed",
		observability.String("operation", "DeleteCategory"),
		observability.String("layer", "handler"),
		observability.String("entity", "category"),
		observability.String("correlation_id", correlationID),
		observability.String("user_id", user.ID),
		observability.String("category_id", categoryID),
	)

	responses.JSON(w, http.StatusNoContent, nil)
}
