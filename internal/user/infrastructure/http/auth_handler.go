package http

import (
	"net/http"

	"github.com/jailtonjunior94/financial/internal/user/application/dtos"
	"github.com/jailtonjunior94/financial/internal/user/application/usecase"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/responses"
)

type AuthHandler struct {
	o11y         observability.Observability
	tokenUseCase usecase.TokenUseCase
}

func NewAuthHandler(
	o11y observability.Observability,
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

	if err := r.ParseForm(); err != nil {
		return
	}

	input := &dtos.AuthInput{
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	}

	output, err := h.tokenUseCase.Execute(ctx, input)
	if err != nil {
		return
	}

	responses.JSON(w, http.StatusOK, output)
	return
}
