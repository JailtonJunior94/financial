package web

import (
	"encoding/json"
	"net/http"

	"github.com/jailtonjunior94/financial/internal/infrastructure/web/middlewares"
	usecase "github.com/jailtonjunior94/financial/internal/usecase/category"
	"github.com/jailtonjunior94/financial/pkg/authentication"
	"github.com/jailtonjunior94/financial/pkg/responses"
)

type CategoryHandler struct {
	useCase usecase.CreateCategoryUseCase
}

func NewCategoryHandler(useCase usecase.CreateCategoryUseCase) *CategoryHandler {
	return &CategoryHandler{useCase: useCase}
}

func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middlewares.UserCtxKey).(*authentication.User)

	var input usecase.CreateCategoryInput
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		responses.Error(w, http.StatusUnprocessableEntity, "unprocessable entity")
		return
	}

	output, err := h.useCase.Execute(user.ID, &input)
	if err != nil {
		responses.Error(w, http.StatusBadRequest, "error creating category")
		return
	}
	responses.JSON(w, http.StatusCreated, output)
}
