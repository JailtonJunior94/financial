package http

import (
	"encoding/json"
	"net/http"

	"github.com/jailtonjunior94/financial/internal/user/application/dtos"
	"github.com/jailtonjunior94/financial/internal/user/application/usecase"
	"github.com/jailtonjunior94/financial/pkg/api/httperrors"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/responses"
)

type UserHandler struct {
	o11y              observability.Observability
	errorHandler      httperrors.ErrorHandler
	createUserUseCase usecase.CreateUserUseCase
}

func NewUserHandler(
	o11y observability.Observability,
	errorHandler httperrors.ErrorHandler,
	createUserUseCase usecase.CreateUserUseCase,
) *UserHandler {
	return &UserHandler{
		o11y:              o11y,
		errorHandler:      errorHandler,
		createUserUseCase: createUserUseCase,
	}
}

// Create godoc
//
//	@Summary		Criar usuário
//	@Description	Cria um novo usuário na plataforma. Email deve ser único.
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dtos.CreateUserInput		true	"Dados do usuário"
//	@Success		201		{object}	dtos.CreateUserOutput		"Usuário criado com sucesso"
//	@Failure		400		{object}	httperrors.ProblemDetail	"Dados inválidos ou mal-formados"
//	@Failure		409		{object}	httperrors.ProblemDetail	"Email já cadastrado"
//	@Failure		500		{object}	httperrors.ProblemDetail	"Erro interno"
//	@Router			/api/v1/users [post]
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "user_handler.create")
	defer span.End()

	var input *dtos.CreateUserInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	output, err := h.createUserUseCase.Execute(ctx, input)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	responses.JSON(w, http.StatusCreated, output)
}
