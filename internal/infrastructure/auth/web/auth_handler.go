package web

import (
	"net/http"

	"github.com/jailtonjunior94/financial/internal/usecase/auth"
	"github.com/jailtonjunior94/financial/pkg/responses"
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
		responses.Error(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	responses.JSON(w, http.StatusOK, output)
}
