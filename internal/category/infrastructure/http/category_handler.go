package http

import (
	"encoding/json"
	"net/http"

	"github.com/jailtonjunior94/financial/internal/category/application/dtos"
	"github.com/jailtonjunior94/financial/internal/category/application/usecase"
	"github.com/jailtonjunior94/financial/pkg/api/httperrors"
	"github.com/jailtonjunior94/financial/pkg/api/middlewares"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/responses"

	"github.com/go-chi/chi/v5"
)

type CategoryHandler struct {
	o11y                  observability.Observability
	errorHandler          httperrors.ErrorHandler
	findCategoryUseCase   usecase.FindCategoryUseCase
	createCategoryUseCase usecase.CreateCategoryUseCase
	findCategoryByUseCase usecase.FindCategoryByUseCase
	updateCategoryUseCase usecase.UpdateCategoryUseCase
	removeCategoryUseCase usecase.RemoveCategoryUseCase
}

func NewCategoryHandler(
	o11y observability.Observability,
	errorHandler httperrors.ErrorHandler,
	findCategoryUseCase usecase.FindCategoryUseCase,
	createCategoryUseCase usecase.CreateCategoryUseCase,
	findCategoryByUseCase usecase.FindCategoryByUseCase,
	updateCategoryUseCase usecase.UpdateCategoryUseCase,
	removeCategoryUseCase usecase.RemoveCategoryUseCase,
) *CategoryHandler {
	return &CategoryHandler{
		o11y:                  o11y,
		errorHandler:          errorHandler,
		findCategoryUseCase:   findCategoryUseCase,
		createCategoryUseCase: createCategoryUseCase,
		updateCategoryUseCase: updateCategoryUseCase,
		findCategoryByUseCase: findCategoryByUseCase,
		removeCategoryUseCase: removeCategoryUseCase,
	}
}

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

	output, err := h.createCategoryUseCase.Execute(ctx, user.ID, input)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	responses.JSON(w, http.StatusCreated, output)
}

func (h *CategoryHandler) Find(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "category_handler.find")
	defer span.End()

	user, err := middlewares.GetUserFromContext(ctx)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	output, err := h.findCategoryUseCase.Execute(ctx, user.ID)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	responses.JSON(w, http.StatusOK, output)
}

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

	output, err := h.updateCategoryUseCase.Execute(ctx, user.ID, chi.URLParam(r, "id"), input)
	if err != nil {
		return
	}

	responses.JSON(w, http.StatusOK, output)
}

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
