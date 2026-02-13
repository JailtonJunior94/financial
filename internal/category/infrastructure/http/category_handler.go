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

	"github.com/go-chi/chi/v5"
)

type CategoryHandler struct {
	o11y                         observability.Observability
	errorHandler                 httperrors.ErrorHandler
	findCategoryUseCase          usecase.FindCategoryUseCase
	findCategoryPaginatedUseCase usecase.FindCategoryPaginatedUseCase
	createCategoryUseCase        usecase.CreateCategoryUseCase
	findCategoryByUseCase        usecase.FindCategoryByUseCase
	updateCategoryUseCase        usecase.UpdateCategoryUseCase
	removeCategoryUseCase        usecase.RemoveCategoryUseCase
}

func NewCategoryHandler(
	o11y observability.Observability,
	errorHandler httperrors.ErrorHandler,
	findCategoryUseCase usecase.FindCategoryUseCase,
	findCategoryPaginatedUseCase usecase.FindCategoryPaginatedUseCase,
	createCategoryUseCase usecase.CreateCategoryUseCase,
	findCategoryByUseCase usecase.FindCategoryByUseCase,
	updateCategoryUseCase usecase.UpdateCategoryUseCase,
	removeCategoryUseCase usecase.RemoveCategoryUseCase,
) *CategoryHandler {
	return &CategoryHandler{
		o11y:                         o11y,
		errorHandler:                 errorHandler,
		findCategoryUseCase:          findCategoryUseCase,
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

	user, err := middlewares.GetUserFromContext(ctx)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	var input *dtos.CategoryInput
	if err = json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	if validationErrs := input.Validate(); validationErrs.HasErrors() {
		h.errorHandler.HandleError(w, r, validationErrs)
		return
	}

	output, err := h.createCategoryUseCase.Execute(ctx, user.ID, input)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

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

	output, err := h.findCategoryPaginatedUseCase.Execute(ctx, usecase.FindCategoryPaginatedInput{
		UserID: user.ID,
		Limit:  params.Limit,
		Cursor: params.Cursor,
	})
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	// Build paginated response
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

	user, err := middlewares.GetUserFromContext(ctx)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	output, err := h.findCategoryByUseCase.Execute(ctx, user.ID, chi.URLParam(r, "id"))
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

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

	user, err := middlewares.GetUserFromContext(ctx)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	var input *dtos.CategoryInput
	if err = json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	if validationErrs := input.Validate(); validationErrs.HasErrors() {
		h.errorHandler.HandleError(w, r, validationErrs)
		return
	}

	output, err := h.updateCategoryUseCase.Execute(ctx, user.ID, chi.URLParam(r, "id"), input)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

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

	user, err := middlewares.GetUserFromContext(ctx)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	if err := h.removeCategoryUseCase.Execute(ctx, user.ID, chi.URLParam(r, "id")); err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	responses.JSON(w, http.StatusNoContent, nil)
}
