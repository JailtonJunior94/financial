package http

import (
	"encoding/json"
	"net/http"

	"github.com/jailtonjunior94/financial/internal/card/application/dtos"
	"github.com/jailtonjunior94/financial/internal/card/application/usecase"
	"github.com/jailtonjunior94/financial/pkg/api/httperrors"
	"github.com/jailtonjunior94/financial/pkg/api/middlewares"
	"github.com/jailtonjunior94/financial/pkg/pagination"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/responses"

	"github.com/go-chi/chi/v5"
)

type CardHandler struct {
	o11y                     observability.Observability
	errorHandler             httperrors.ErrorHandler
	findCardUseCase          usecase.FindCardUseCase
	findCardPaginatedUseCase usecase.FindCardPaginatedUseCase
	createCardUseCase        usecase.CreateCardUseCase
	findCardByUseCase        usecase.FindCardByUseCase
	updateCardUseCase        usecase.UpdateCardUseCase
	removeCardUseCase        usecase.RemoveCardUseCase
}

func NewCardHandler(
	o11y observability.Observability,
	errorHandler httperrors.ErrorHandler,
	findCardUseCase usecase.FindCardUseCase,
	findCardPaginatedUseCase usecase.FindCardPaginatedUseCase,
	createCardUseCase usecase.CreateCardUseCase,
	findCardByUseCase usecase.FindCardByUseCase,
	updateCardUseCase usecase.UpdateCardUseCase,
	removeCardUseCase usecase.RemoveCardUseCase,
) *CardHandler {
	return &CardHandler{
		o11y:                     o11y,
		errorHandler:             errorHandler,
		findCardUseCase:          findCardUseCase,
		findCardPaginatedUseCase: findCardPaginatedUseCase,
		createCardUseCase:        createCardUseCase,
		updateCardUseCase:        updateCardUseCase,
		findCardByUseCase:        findCardByUseCase,
		removeCardUseCase:        removeCardUseCase,
	}
}

// Create godoc
//
//	@Summary		Criar cartão
//	@Description	Cria um novo cartão de crédito para o usuário autenticado.
//	@Description	O campo `closing_offset_days` é opcional (default: 7 dias antes do vencimento).
//	@Description	- `due_day`: dia do mês em que a fatura vence (1–31)
//	@Description	- `closing_offset_days`: quantos dias antes do vencimento a fatura fecha (1–31)
//	@Tags			cards
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dtos.CardInput				true	"Dados do cartão"
//	@Success		201		{object}	dtos.CardOutput				"Cartão criado com sucesso"
//	@Failure		400		{object}	httperrors.ProblemDetail	"Dados inválidos"
//	@Failure		401		{object}	httperrors.ProblemDetail	"Não autenticado"
//	@Failure		500		{object}	httperrors.ProblemDetail	"Erro interno"
//	@Router			/api/v1/cards [post]
func (h *CardHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "card_handler.create")
	defer span.End()

	user, err := middlewares.GetUserFromContext(ctx)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	var input *dtos.CardInput
	if err = json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	if validationErrs := input.Validate(); validationErrs.HasErrors() {
		h.errorHandler.HandleError(w, r, validationErrs)
		return
	}

	output, err := h.createCardUseCase.Execute(ctx, user.ID, input)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	responses.JSON(w, http.StatusCreated, output)
}

// Find godoc
//
//	@Summary		Listar cartões
//	@Description	Retorna a lista paginada de cartões do usuário autenticado (cursor-based pagination).
//	@Tags			cards
//	@Produce		json
//	@Security		BearerAuth
//	@Param			limit	query		integer	false	"Itens por página (default: 20, max: 100)"	minimum(1)	maximum(100)	default(20)
//	@Param			cursor	query		string	false	"Cursor de paginação (retornado em pagination.next_cursor)"
//	@Success		200		{object}	dtos.CardPaginatedOutput	"Lista paginada de cartões"
//	@Failure		400		{object}	httperrors.ProblemDetail					"Parâmetro inválido"
//	@Failure		401		{object}	httperrors.ProblemDetail					"Não autenticado"
//	@Failure		500		{object}	httperrors.ProblemDetail					"Erro interno"
//	@Router			/api/v1/cards [get]
func (h *CardHandler) Find(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "card_handler.find")
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

	output, err := h.findCardPaginatedUseCase.Execute(ctx, usecase.FindCardPaginatedInput{
		UserID: user.ID,
		Limit:  params.Limit,
		Cursor: params.Cursor,
	})
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	// Build paginated response
	response := pagination.NewPaginatedResponse(output.Cards, params.Limit, output.NextCursor)
	responses.JSON(w, http.StatusOK, response)
}

// FindBy godoc
//
//	@Summary		Buscar cartão por ID
//	@Description	Retorna os detalhes de um cartão específico do usuário autenticado.
//	@Tags			cards
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string					true	"ID do cartão"	format(uuid)
//	@Success		200	{object}	dtos.CardOutput			"Dados do cartão"
//	@Failure		401	{object}	httperrors.ProblemDetail	"Não autenticado"
//	@Failure		404	{object}	httperrors.ProblemDetail	"Cartão não encontrado"
//	@Failure		500	{object}	httperrors.ProblemDetail	"Erro interno"
//	@Router			/api/v1/cards/{id} [get]
func (h *CardHandler) FindBy(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "card_handler.find_by")
	defer span.End()

	user, err := middlewares.GetUserFromContext(ctx)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	output, err := h.findCardByUseCase.Execute(ctx, user.ID, chi.URLParam(r, "id"))
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	responses.JSON(w, http.StatusOK, output)
}

// Update godoc
//
//	@Summary		Atualizar cartão
//	@Description	Atualiza os dados de um cartão existente do usuário autenticado.
//	@Tags			cards
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string					true	"ID do cartão"	format(uuid)
//	@Param			request	body		dtos.CardInput			true	"Dados atualizados do cartão"
//	@Success		200		{object}	dtos.CardOutput			"Cartão atualizado com sucesso"
//	@Failure		400		{object}	httperrors.ProblemDetail	"Dados inválidos"
//	@Failure		401		{object}	httperrors.ProblemDetail	"Não autenticado"
//	@Failure		404		{object}	httperrors.ProblemDetail	"Cartão não encontrado"
//	@Failure		500		{object}	httperrors.ProblemDetail	"Erro interno"
//	@Router			/api/v1/cards/{id} [put]
func (h *CardHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "card_handler.update")
	defer span.End()

	user, err := middlewares.GetUserFromContext(ctx)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	var input *dtos.CardInput
	if err = json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	if validationErrs := input.Validate(); validationErrs.HasErrors() {
		h.errorHandler.HandleError(w, r, validationErrs)
		return
	}

	output, err := h.updateCardUseCase.Execute(ctx, user.ID, chi.URLParam(r, "id"), input)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	responses.JSON(w, http.StatusOK, output)
}

// Delete godoc
//
//	@Summary		Remover cartão
//	@Description	Remove um cartão do usuário autenticado. Esta operação é irreversível.
//	@Tags			cards
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	string	true	"ID do cartão"	format(uuid)
//	@Success		204	"Cartão removido com sucesso"
//	@Failure		401	{object}	httperrors.ProblemDetail	"Não autenticado"
//	@Failure		404	{object}	httperrors.ProblemDetail	"Cartão não encontrado"
//	@Failure		500	{object}	httperrors.ProblemDetail	"Erro interno"
//	@Router			/api/v1/cards/{id} [delete]
func (h *CardHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "card_handler.delete")
	defer span.End()

	user, err := middlewares.GetUserFromContext(ctx)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	if err := h.removeCardUseCase.Execute(ctx, user.ID, chi.URLParam(r, "id")); err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	responses.JSON(w, http.StatusNoContent, nil)
}
