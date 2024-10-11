package http

import (
	"encoding/json"
	"net/http"

	"github.com/jailtonjunior94/financial/internal/category/application/dtos"
	"github.com/jailtonjunior94/financial/internal/category/application/usecase"
	"github.com/jailtonjunior94/financial/pkg/api/middlewares"

	"github.com/JailtonJunior94/devkit-go/pkg/o11y"
	"github.com/JailtonJunior94/devkit-go/pkg/responses"
)

type CategoryHandler struct {
	o11y                  o11y.Observability
	createCategoryUseCase usecase.CreateCategoryUseCase
	findCategoryUseCase   usecase.FindCategoryUseCase
}

func NewCategoryHandler(
	o11y o11y.Observability,
	createCategoryUseCase usecase.CreateCategoryUseCase,
	findCategoryUseCase usecase.FindCategoryUseCase,
) *CategoryHandler {
	return &CategoryHandler{
		o11y:                  o11y,
		findCategoryUseCase:   findCategoryUseCase,
		createCategoryUseCase: createCategoryUseCase,
	}
}

func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) error {
	ctx, span := h.o11y.Start(r.Context(), "category_handler.create")
	defer span.End()

	user := middlewares.GetUserFromContext(ctx)

	var input *dtos.CreateCategoryInput
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		span.RecordError(err)
		responses.Error(w, http.StatusUnprocessableEntity, "unprocessable entity")
		return nil
	}

	output, err := h.createCategoryUseCase.Execute(ctx, user.ID, input)
	if err != nil {
		span.RecordError(err)
		responses.Error(w, http.StatusBadRequest, "error creating category")
		return nil
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
		responses.Error(w, http.StatusBadRequest, "error finding categories")
		return nil
	}
	responses.JSON(w, http.StatusCreated, output)
	return nil
}
