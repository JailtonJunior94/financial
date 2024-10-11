package http

import (
	"net/http"

	"github.com/jailtonjunior94/financial/internal/user/application/dtos"
	"github.com/jailtonjunior94/financial/internal/user/application/usecase"

	"github.com/JailtonJunior94/devkit-go/pkg/o11y"
	"github.com/JailtonJunior94/devkit-go/pkg/responses"
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

func (h *AuthHandler) Token(w http.ResponseWriter, r *http.Request) error {
	ctx, span := h.o11y.Start(r.Context(), "auth_handler.token")
	defer span.End()

	input := &dtos.AuthInput{
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	}

	output, err := h.tokenUseCase.Execute(ctx, input)
	if err != nil {
		span.RecordError(err)
		return err
	}

	responses.JSON(w, http.StatusOK, output)
	return nil
}
