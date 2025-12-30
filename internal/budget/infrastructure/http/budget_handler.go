package http

import (
	"encoding/json"
	"net/http"

	"github.com/jailtonjunior94/financial/internal/budget/domain/dtos"
	"github.com/jailtonjunior94/financial/internal/budget/usecase"
	"github.com/jailtonjunior94/financial/pkg/api/middlewares"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/responses"
)

type BudgetHandler struct {
	o11y                observability.Observability
	createBudgetUseCase usecase.CreateBudgetUseCase
}

func NewBudgetHandler(
	o11y observability.Observability,
	createBudgetUseCase usecase.CreateBudgetUseCase,
) *BudgetHandler {
	return &BudgetHandler{
		o11y:                o11y,
		createBudgetUseCase: createBudgetUseCase,
	}
}

func (h *BudgetHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "budget_handler.create")
	defer span.End()

	user, err := middlewares.GetUserFromContext(ctx)
	if err != nil {
		return
	}

	var input *dtos.BugetInput
	if err = json.NewDecoder(r.Body).Decode(&input); err != nil {
		return
	}

	output, err := h.createBudgetUseCase.Execute(ctx, user.ID, input)
	if err != nil {
		return
	}

	responses.JSON(w, http.StatusCreated, output)
}
