package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jailtonjunior94/financial/internal/budget/application/dtos"
	"github.com/jailtonjunior94/financial/internal/budget/application/usecase"
	"github.com/jailtonjunior94/financial/pkg/api/httperrors"
	"github.com/jailtonjunior94/financial/pkg/api/middlewares"
	"github.com/jailtonjunior94/financial/pkg/pagination"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/responses"
)

type BudgetHandler struct {
	o11y                        observability.Observability
	errorHandler                httperrors.ErrorHandler
	createBudgetUseCase         usecase.CreateBudgetUseCase
	findBudgetUseCase           usecase.FindBudgetUseCase
	updateBudgetUseCase         usecase.UpdateBudgetUseCase
	deleteBudgetUseCase         usecase.DeleteBudgetUseCase
	listBudgetsPaginatedUseCase usecase.ListBudgetsPaginatedUseCase
}

func NewBudgetHandler(
	o11y observability.Observability,
	errorHandler httperrors.ErrorHandler,
	createBudgetUseCase usecase.CreateBudgetUseCase,
	findBudgetUseCase usecase.FindBudgetUseCase,
	updateBudgetUseCase usecase.UpdateBudgetUseCase,
	deleteBudgetUseCase usecase.DeleteBudgetUseCase,
	listBudgetsPaginatedUseCase usecase.ListBudgetsPaginatedUseCase,
) *BudgetHandler {
	return &BudgetHandler{
		o11y:                        o11y,
		errorHandler:                errorHandler,
		createBudgetUseCase:         createBudgetUseCase,
		findBudgetUseCase:           findBudgetUseCase,
		updateBudgetUseCase:         updateBudgetUseCase,
		deleteBudgetUseCase:         deleteBudgetUseCase,
		listBudgetsPaginatedUseCase: listBudgetsPaginatedUseCase,
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

	if validationErrs := input.Validate(); validationErrs.HasErrors() {
		h.errorHandler.HandleError(w, r, validationErrs)
		return
	}

	output, err := h.createBudgetUseCase.Execute(ctx, user.ID, input)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	responses.JSON(w, http.StatusCreated, output)
}

func (h *BudgetHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "budget_handler.list")
	defer span.End()

	user, err := middlewares.GetUserFromContext(ctx)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	// Parse cursor parameters (default: limit=20, max=100)
	params, err := pagination.ParseCursorParams(r, 20, 100)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	output, err := h.listBudgetsPaginatedUseCase.Execute(ctx, usecase.ListBudgetsPaginatedInput{
		UserID: user.ID,
		Limit:  params.Limit,
		Cursor: params.Cursor,
	})
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	// Build paginated response
	response := pagination.NewPaginatedResponse(output.Budgets, params.Limit, output.NextCursor)
	responses.JSON(w, http.StatusOK, response)
}

func (h *BudgetHandler) Find(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "budget_handler.find")
	defer span.End()

	budgetID := r.PathValue("id")
	if budgetID == "" {
		h.errorHandler.HandleError(w, r, fmt.Errorf("budget_id is required"))
		return
	}

	output, err := h.findBudgetUseCase.Execute(ctx, budgetID)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	responses.JSON(w, http.StatusOK, output)
}

func (h *BudgetHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "budget_handler.update")
	defer span.End()

	budgetID := r.PathValue("id")
	if budgetID == "" {
		h.errorHandler.HandleError(w, r, fmt.Errorf("budget_id is required"))
		return
	}

	var input *dtos.BudgetUpdateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	if validationErrs := input.Validate(); validationErrs.HasErrors() {
		h.errorHandler.HandleError(w, r, validationErrs)
		return
	}

	output, err := h.updateBudgetUseCase.Execute(ctx, budgetID, input)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	responses.JSON(w, http.StatusOK, output)
}

func (h *BudgetHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "budget_handler.delete")
	defer span.End()

	budgetID := r.PathValue("id")
	if budgetID == "" {
		h.errorHandler.HandleError(w, r, fmt.Errorf("budget_id is required"))
		return
	}

	err := h.deleteBudgetUseCase.Execute(ctx, budgetID)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	responses.JSON(w, http.StatusNoContent, nil)
}
