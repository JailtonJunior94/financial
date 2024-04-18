package rest

import (
	"net/http"

	"github.com/jailtonjunior94/financial/internal/user/usecase"
	"github.com/jailtonjunior94/financial/pkg/observability"
	"github.com/jailtonjunior94/financial/pkg/responses"
)

type AuthHandler struct {
	tokenUseCase  usecase.TokenUseCase
	observability observability.Observability
}

func NewAuthHandler(observability observability.Observability, tokenUseCase usecase.TokenUseCase) *AuthHandler {
	return &AuthHandler{observability: observability, tokenUseCase: tokenUseCase}
}

func (h *AuthHandler) Token(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.observability.Tracer().Start(r.Context(), "auth_handler.token")
	defer span.End()

	input := &usecase.AuthInput{
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	}

	output, err := h.tokenUseCase.Execute(ctx, input)
	if err != nil {
		responses.Error(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	responses.JSON(w, http.StatusOK, output)
}
