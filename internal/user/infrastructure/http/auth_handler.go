package http

import (
	"net/http"

	"github.com/jailtonjunior94/financial/internal/user/usecase"
	"github.com/jailtonjunior94/financial/pkg/o11y"
	"github.com/jailtonjunior94/financial/pkg/responses"
)

type AuthHandler struct {
	o11y         o11y.Observability
	tokenUseCase usecase.TokenUseCase
}

func NewAuthHandler(
	o11y o11y.Observability,
	tokenUseCase usecase.TokenUseCase,
) *AuthHandler {
	return &AuthHandler{
		o11y:         o11y,
		tokenUseCase: tokenUseCase,
	}
}

func (h *AuthHandler) Token(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "auth_handler.token")
	defer span.End()

	input := &usecase.AuthInput{
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	}

	output, err := h.tokenUseCase.Execute(ctx, input)
	if err != nil {
		span.RecordError(err)
		responses.Error(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	responses.JSON(w, http.StatusOK, output)
}
