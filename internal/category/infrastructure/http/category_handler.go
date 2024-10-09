package http

import (
	"encoding/json"
	"net/http"

	"github.com/jailtonjunior94/financial/internal/category/usecase"
	"github.com/jailtonjunior94/financial/pkg/auth"
	"github.com/jailtonjunior94/financial/pkg/http/middlewares"

	"github.com/JailtonJunior94/devkit-go/pkg/o11y"
	"github.com/JailtonJunior94/devkit-go/pkg/responses"
)

type CategoryHandler struct {
	o11y                  o11y.Observability
	createCategoryUseCase usecase.CreateCategoryUseCase
}

func NewCategoryHandler(
	o11y o11y.Observability,
	createCategoryUseCase usecase.CreateCategoryUseCase,
) *CategoryHandler {
	return &CategoryHandler{
		o11y:                  o11y,
		createCategoryUseCase: createCategoryUseCase,
	}
}

func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Start(r.Context(), "category_handler.create")
	defer span.End()

	user := r.Context().Value(middlewares.UserCtxKey).(*auth.User)
	var input *usecase.CreateCategoryInput
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		span.RecordError(err)
		responses.Error(w, http.StatusUnprocessableEntity, "unprocessable entity")
		return
	}

	output, err := h.createCategoryUseCase.Execute(ctx, user.ID, input)
	if err != nil {
		span.RecordError(err)
		responses.Error(w, http.StatusBadRequest, "error creating category")
		return
	}
	responses.JSON(w, http.StatusCreated, output)
}
