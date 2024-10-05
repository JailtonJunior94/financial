package http

import (
	"encoding/json"
	"net/http"

	"github.com/jailtonjunior94/financial/internal/budget/domain/dtos"
	"github.com/jailtonjunior94/financial/internal/budget/usecase"
	"github.com/jailtonjunior94/financial/pkg/auth"
	"github.com/jailtonjunior94/financial/pkg/http/middlewares"
	"github.com/jailtonjunior94/financial/pkg/responses"

	"github.com/JailtonJunior94/devkit-go/pkg/o11y"
)

type CategoryHandler struct {
	o11y                o11y.Observability
	createBudgetUseCase usecase.CreateBudgetUseCase
}

func NewBudgetHandler(
	o11y o11y.Observability,
	createBudgetUseCase usecase.CreateBudgetUseCase,
) *CategoryHandler {
	return &CategoryHandler{
		o11y:                o11y,
		createBudgetUseCase: createBudgetUseCase,
	}
}

func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Start(r.Context(), "budget_handler.create")
	defer span.End()

	user := r.Context().Value(middlewares.UserCtxKey).(*auth.User)
	var input *dtos.BugetInput
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		span.RecordError(err)
		responses.Error(w, http.StatusUnprocessableEntity, "unprocessable entity")
		return
	}

	output, err := h.createBudgetUseCase.Execute(ctx, user.ID, input)
	if err != nil {
		span.RecordError(err)
		responses.Error(w, http.StatusBadRequest, "error creating budget")
		return
	}
	responses.JSON(w, http.StatusCreated, output)
}
