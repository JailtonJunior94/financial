package http

import (
	"encoding/json"
	"net/http"

	"github.com/jailtonjunior94/financial/internal/user/application/dtos"
	"github.com/jailtonjunior94/financial/internal/user/application/usecase"

	"github.com/JailtonJunior94/devkit-go/pkg/o11y"
	"github.com/JailtonJunior94/devkit-go/pkg/responses"
)

type UserHandler struct {
	o11y              o11y.Observability
	createUserUseCase usecase.CreateUserUseCase
}

func NewUserHandler(
	o11y o11y.Observability,
	createUserUseCase usecase.CreateUserUseCase,
) *UserHandler {
	return &UserHandler{
		o11y:              o11y,
		createUserUseCase: createUserUseCase,
	}
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) error {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "user_handler.create")
	defer span.End()

	var input *dtos.CreateUserInput
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		span.RecordError(err)
		responses.Error(w, http.StatusUnprocessableEntity, "Unprocessable Entity")
		return nil
	}

	output, err := h.createUserUseCase.Execute(ctx, input)
	if err != nil {
		span.RecordError(err)
		responses.Error(w, http.StatusBadRequest, "error creating user")
		return nil
	}
	responses.JSON(w, http.StatusCreated, output)
	return nil
}
