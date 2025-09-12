package http

import (
	"encoding/json"
	"net/http"

	"github.com/jailtonjunior94/financial/internal/category/application/dtos"
	"github.com/jailtonjunior94/financial/internal/category/application/usecase"
	"github.com/jailtonjunior94/financial/pkg/api/middlewares"

	"github.com/JailtonJunior94/devkit-go/pkg/o11y"
	"github.com/JailtonJunior94/devkit-go/pkg/responses"

	"github.com/go-chi/chi/v5"
)

type CategoryHandler struct {
	o11y                  o11y.Observability
	createCategoryUseCase usecase.CreateCategoryUseCase
	findCategoryUseCase   usecase.FindCategoryUseCase
	findCategoryByUseCase usecase.FindCategoryByUseCase
	updateCategoryUseCase usecase.UpdateCategoryUseCase
}

func NewCategoryHandler(
	o11y o11y.Observability,
	createCategoryUseCase usecase.CreateCategoryUseCase,
	findCategoryUseCase usecase.FindCategoryUseCase,
	findCategoryByUseCase usecase.FindCategoryByUseCase,
	updateCategoryUseCase usecase.UpdateCategoryUseCase,
) *CategoryHandler {
	return &CategoryHandler{
		o11y:                  o11y,
		findCategoryUseCase:   findCategoryUseCase,
		createCategoryUseCase: createCategoryUseCase,
		updateCategoryUseCase: updateCategoryUseCase,
		findCategoryByUseCase: findCategoryByUseCase,
	}
}

func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) error {
	ctx, span := h.o11y.Start(r.Context(), "category_handler.create")
	defer span.End()

	user := middlewares.GetUserFromContext(ctx)

	var input *dtos.CategoryInput
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		span.RecordError(err)
		return err
	}

	output, err := h.createCategoryUseCase.Execute(ctx, user.ID, input)
	if err != nil {
		span.RecordError(err)
		return err
	}

	responses.JSON(w, http.StatusCreated, output)
	return nil
}

func (h *CategoryHandler) Find(w http.ResponseWriter, r *http.Request) error {
	ctx, span := h.o11y.Start(r.Context(), "category_handler.find")
	defer span.End()

	user := middlewares.GetUserFromContext(ctx)
	output, err := h.findCategoryUseCase.Execute(ctx, user.ID)
	if err != nil {
		span.RecordError(err)
		return err
	}

	responses.JSON(w, http.StatusOK, output)
	return nil
}

func (h *CategoryHandler) FindBy(w http.ResponseWriter, r *http.Request) error {
	ctx, span := h.o11y.Start(r.Context(), "category_handler.find")
	defer span.End()

	user := middlewares.GetUserFromContext(ctx)
	output, err := h.findCategoryByUseCase.Execute(ctx, user.ID, chi.URLParam(r, "id"))
	if err != nil {
		span.RecordError(err)
		return err
	}

	responses.JSON(w, http.StatusOK, output)
	return nil
}

func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) error {
	ctx, span := h.o11y.Start(r.Context(), "category_handler.update")
	defer span.End()

	user := middlewares.GetUserFromContext(ctx)

	var input *dtos.CategoryInput
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		span.RecordError(err)
		return err
	}

	output, err := h.updateCategoryUseCase.Execute(ctx, user.ID, chi.URLParam(r, "id"), input)
	if err != nil {
		span.RecordError(err)
		return err
	}

	responses.JSON(w, http.StatusOK, output)
	return nil

}
