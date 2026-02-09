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
