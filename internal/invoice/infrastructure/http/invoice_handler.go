package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/responses"
	"github.com/go-chi/chi/v5"

	"github.com/jailtonjunior94/financial/internal/invoice/application/dtos"
	"github.com/jailtonjunior94/financial/internal/invoice/application/usecase"
	"github.com/jailtonjunior94/financial/pkg/api/httperrors"
	"github.com/jailtonjunior94/financial/pkg/api/middlewares"
)

type InvoiceHandler struct {
	o11y                       observability.Observability
	errorHandler               httperrors.ErrorHandler
	createPurchaseUseCase      usecase.CreatePurchaseUseCase
	updatePurchaseUseCase      usecase.UpdatePurchaseUseCase
	deletePurchaseUseCase      usecase.DeletePurchaseUseCase
	getInvoiceUseCase          usecase.GetInvoiceUseCase
	listInvoicesByMonthUseCase usecase.ListInvoicesByMonthUseCase
	listInvoicesByCardUseCase  usecase.ListInvoicesByCardUseCase
}

func NewInvoiceHandler(
	o11y observability.Observability,
	errorHandler httperrors.ErrorHandler,
	createPurchaseUseCase usecase.CreatePurchaseUseCase,
	updatePurchaseUseCase usecase.UpdatePurchaseUseCase,
	deletePurchaseUseCase usecase.DeletePurchaseUseCase,
	getInvoiceUseCase usecase.GetInvoiceUseCase,
	listInvoicesByMonthUseCase usecase.ListInvoicesByMonthUseCase,
	listInvoicesByCardUseCase usecase.ListInvoicesByCardUseCase,
) *InvoiceHandler {
	return &InvoiceHandler{
		o11y:                       o11y,
		errorHandler:               errorHandler,
		createPurchaseUseCase:      createPurchaseUseCase,
		updatePurchaseUseCase:      updatePurchaseUseCase,
		deletePurchaseUseCase:      deletePurchaseUseCase,
		getInvoiceUseCase:          getInvoiceUseCase,
		listInvoicesByMonthUseCase: listInvoicesByMonthUseCase,
		listInvoicesByCardUseCase:  listInvoicesByCardUseCase,
	}
}

// CreatePurchase creates a new purchase with installments.
func (h *InvoiceHandler) CreatePurchase(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "invoice_handler.create_purchase")
	defer span.End()

	user, err := middlewares.GetUserFromContext(ctx)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	var input *dtos.PurchaseCreateInput
	if err = json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	if err := h.createPurchaseUseCase.Execute(ctx, user.ID, input); err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	responses.JSON(w, http.StatusCreated, map[string]string{
		"message": "Purchase created successfully",
	})
}

// UpdatePurchase updates all installments of a purchase.
func (h *InvoiceHandler) UpdatePurchase(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "invoice_handler.update_purchase")
	defer span.End()

	_, err := middlewares.GetUserFromContext(ctx)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	var input *dtos.PurchaseUpdateInput
	if err = json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	itemID := chi.URLParam(r, "id")
	if err := h.updatePurchaseUseCase.Execute(ctx, itemID, input); err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	responses.JSON(w, http.StatusOK, map[string]string{
		"message": "Purchase updated successfully",
	})
}

// DeletePurchase deletes all installments of a purchase.
func (h *InvoiceHandler) DeletePurchase(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "invoice_handler.delete_purchase")
	defer span.End()

	_, err := middlewares.GetUserFromContext(ctx)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	itemID := chi.URLParam(r, "id")
	if err := h.deletePurchaseUseCase.Execute(ctx, itemID); err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	responses.JSON(w, http.StatusNoContent, nil)
}

// GetInvoice retrieves a single invoice with its items.
func (h *InvoiceHandler) GetInvoice(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "invoice_handler.get_invoice")
	defer span.End()

	_, err := middlewares.GetUserFromContext(ctx)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	invoiceID := chi.URLParam(r, "id")
	output, err := h.getInvoiceUseCase.Execute(ctx, invoiceID)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	responses.JSON(w, http.StatusOK, output)
}

// ListInvoicesByMonth lists all invoices for a user in a specific month.
func (h *InvoiceHandler) ListInvoicesByMonth(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "invoice_handler.list_invoices_by_month")
	defer span.End()

	user, err := middlewares.GetUserFromContext(ctx)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	// Get month from query parameter (e.g., ?month=2025-01)
	month := r.URL.Query().Get("month")
	if month == "" {
		h.errorHandler.HandleError(w, r, fmt.Errorf("month parameter is required"))
		return
	}

	output, err := h.listInvoicesByMonthUseCase.Execute(ctx, user.ID, month)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	responses.JSON(w, http.StatusOK, output)
}

// ListInvoicesByCard lists all invoices for a specific card.
func (h *InvoiceHandler) ListInvoicesByCard(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "invoice_handler.list_invoices_by_card")
	defer span.End()

	_, err := middlewares.GetUserFromContext(ctx)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	cardID := chi.URLParam(r, "cardId")
	output, err := h.listInvoicesByCardUseCase.Execute(ctx, cardID)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	responses.JSON(w, http.StatusOK, output)
}
