package web

import (
	"encoding/json"
	"net/http"

	usecase "github.com/jailtonjunior94/financial/internal/usecase/user"
	"github.com/jailtonjunior94/financial/pkg/responses"
)

type UserHandler struct {
	useCase usecase.CreateUserUseCase
}

func NewUserHandler(useCase usecase.CreateUserUseCase) *UserHandler {
	return &UserHandler{useCase: useCase}
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input usecase.CreateUserInput
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		responses.Error(w, http.StatusUnprocessableEntity, "Unprocessable Entity")
		return
	}

	output, err := h.useCase.Execute(&input)
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	responses.JSON(w, http.StatusCreated, output)
}
