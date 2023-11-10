package web

import (
	"encoding/json"
	"net/http"

	usecase "github.com/jailtonjunior94/financial/internal/usecase/user"
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	output, err := h.useCase.Execute(&input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(output)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
