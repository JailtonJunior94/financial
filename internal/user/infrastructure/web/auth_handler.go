package web

import (
	"net/http"

	"github.com/jailtonjunior94/financial/internal/user/usecase"
	"github.com/jailtonjunior94/financial/pkg/responses"
)

type AuthHandler struct {
	tokenUseCase usecase.TokenUseCase
}

func NewAuthHandler(tokenUseCase usecase.TokenUseCase) *AuthHandler {
	return &AuthHandler{tokenUseCase: tokenUseCase}
}

func (h *AuthHandler) Token(w http.ResponseWriter, r *http.Request) {
	input := &usecase.AuthInput{
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	}

	output, err := h.tokenUseCase.Execute(r.Context(), input)
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	responses.JSON(w, http.StatusOK, output)
}
