package web

import (
	"encoding/json"
	"net/http"

	"github.com/jailtonjunior94/financial/internal/user/usecase"
	"github.com/jailtonjunior94/financial/pkg/responses"
)

type UserHandler struct {
	createUserUseCase usecase.CreateUserUseCase
}

func NewUserHandler(createUserUseCase usecase.CreateUserUseCase) *UserHandler {
	return &UserHandler{createUserUseCase: createUserUseCase}
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input usecase.CreateUserInput
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		responses.Error(w, http.StatusUnprocessableEntity, "Unprocessable Entity")
		return
	}

	output, err := h.createUserUseCase.Execute(r.Context(), &input)
	if err != nil {
		responses.Error(w, http.StatusBadRequest, "error creating user")
		return
	}
	responses.JSON(w, http.StatusCreated, output)
}
