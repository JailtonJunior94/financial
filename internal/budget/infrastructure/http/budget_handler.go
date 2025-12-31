package http

import (
	"encoding/json"
	"net/http"

	"github.com/jailtonjunior94/financial/internal/budget/application/dtos"
	"github.com/jailtonjunior94/financial/internal/budget/application/usecase"
	"github.com/jailtonjunior94/financial/pkg/api/httperrors"
	"github.com/jailtonjunior94/financial/pkg/api/middlewares"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/responses"
)

type BudgetHandler struct {
	o11y                observability.Observability
	errorHandler        httperrors.ErrorHandler
	createBudgetUseCase usecase.CreateBudgetUseCase
}

func NewBudgetHandler(
	o11y observability.Observability,
	errorHandler httperrors.ErrorHandler,
	createBudgetUseCase usecase.CreateBudgetUseCase,
) *BudgetHandler {
	return &BudgetHandler{
		o11y:                o11y,
		errorHandler:        errorHandler,
		createBudgetUseCase: createBudgetUseCase,
	}
}

func (h *BudgetHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "budget_handler.create")
	defer span.End()

	user, err := middlewares.GetUserFromContext(ctx)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	var input *dtos.BudgetCreateInput
	if err = json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	output, err := h.createBudgetUseCase.Execute(ctx, user.ID, input)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	responses.JSON(w, http.StatusCreated, output)
}
