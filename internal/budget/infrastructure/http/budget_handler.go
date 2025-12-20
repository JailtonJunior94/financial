package http

import (
	"encoding/json"
	"net/http"

	"github.com/jailtonjunior94/financial/internal/budget/domain/dtos"
	"github.com/jailtonjunior94/financial/internal/budget/usecase"
	"github.com/jailtonjunior94/financial/pkg/api/middlewares"

	"github.com/JailtonJunior94/devkit-go/pkg/o11y"
	"github.com/JailtonJunior94/devkit-go/pkg/responses"
)

type CategoryHandler struct {
	o11y                o11y.Telemetry
	createBudgetUseCase usecase.CreateBudgetUseCase
}

func NewBudgetHandler(
	o11y o11y.Telemetry,
	createBudgetUseCase usecase.CreateBudgetUseCase,
) *CategoryHandler {
	return &CategoryHandler{
		o11y:                o11y,
		createBudgetUseCase: createBudgetUseCase,
	}
}

func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "budget_handler.create")
	defer span.End()

	user, err := middlewares.GetUserFromContext(ctx)
	if err != nil {
		responses.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var input *dtos.BugetInput
	if err = json.NewDecoder(r.Body).Decode(&input); err != nil {
		responses.Error(w, http.StatusUnprocessableEntity, "unprocessable entity")
		return
	}

	output, err := h.createBudgetUseCase.Execute(ctx, user.ID, input)
	if err != nil {
		responses.Error(w, http.StatusBadRequest, "error creating budget")
		return
	}
	responses.JSON(w, http.StatusCreated, output)
}
