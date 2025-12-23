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

type BudgetHandler struct {
	o11y                o11y.Telemetry
	createBudgetUseCase usecase.CreateBudgetUseCase
}

func NewBudgetHandler(
	o11y o11y.Telemetry,
	createBudgetUseCase usecase.CreateBudgetUseCase,
) *BudgetHandler {
	return &BudgetHandler{
		o11y:                o11y,
		createBudgetUseCase: createBudgetUseCase,
	}
}

func (h *BudgetHandler) Create(w http.ResponseWriter, r *http.Request) error {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "budget_handler.create")
	defer span.End()

	user, err := middlewares.GetUserFromContext(ctx)
	if err != nil {
		return err
	}

	var input *dtos.BugetInput
	if err = json.NewDecoder(r.Body).Decode(&input); err != nil {
		return err
	}

	output, err := h.createBudgetUseCase.Execute(ctx, user.ID, input)
	if err != nil {
		return err
	}

	responses.JSON(w, http.StatusCreated, output)
	return nil
}
