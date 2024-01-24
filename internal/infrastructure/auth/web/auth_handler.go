package web

import (
	"encoding/json"
	"net/http"

	"github.com/jailtonjunior94/financial/internal/usecase/auth"
)

type AuthHandler struct {
	useCase auth.TokenUseCase
}

func NewAuthHandler(useCase auth.TokenUseCase) *AuthHandler {
	return &AuthHandler{useCase: useCase}
}

func (h *AuthHandler) Token(w http.ResponseWriter, r *http.Request) {
	input := &auth.AuthInput{
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	}

	output, err := h.useCase.Execute(input)
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
