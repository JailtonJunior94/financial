package rest

import (
	"encoding/json"
	"net/http"

	"github.com/jailtonjunior94/financial/internal/category/usecase"
	"github.com/jailtonjunior94/financial/internal/shared/http/middlewares"
	"github.com/jailtonjunior94/financial/pkg/auth"
	"github.com/jailtonjunior94/financial/pkg/responses"
)

type CategoryHandler struct {
	createCategoryUseCase usecase.CreateCategoryUseCase
}

func NewCategoryHandler(createCategoryUseCase usecase.CreateCategoryUseCase) *CategoryHandler {
	return &CategoryHandler{createCategoryUseCase: createCategoryUseCase}
}

func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middlewares.UserCtxKey).(*auth.User)

	var input *usecase.CreateCategoryInput
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		responses.Error(w, http.StatusUnprocessableEntity, "unprocessable entity")
		return
	}

	output, err := h.createCategoryUseCase.Execute(user.ID, input)
	if err != nil {
		responses.Error(w, http.StatusBadRequest, "error creating category")
		return
	}
	responses.JSON(w, http.StatusCreated, output)
}
