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
	"github.com/jailtonjunior94/financial/pkg/pagination"
)

type InvoiceHandler struct {
	o11y                                observability.Observability
	errorHandler                        httperrors.ErrorHandler
	createPurchaseUseCase               usecase.CreatePurchaseUseCase
	updatePurchaseUseCase               usecase.UpdatePurchaseUseCase
	deletePurchaseUseCase               usecase.DeletePurchaseUseCase
	getInvoiceUseCase                   usecase.GetInvoiceUseCase
	listInvoicesByMonthUseCase          usecase.ListInvoicesByMonthUseCase
	listInvoicesByMonthPaginatedUseCase usecase.ListInvoicesByMonthPaginatedUseCase
	listInvoicesByCardUseCase           usecase.ListInvoicesByCardUseCase
	listInvoicesByCardPaginatedUseCase  usecase.ListInvoicesByCardPaginatedUseCase
}

func NewInvoiceHandler(
	o11y observability.Observability,
	errorHandler httperrors.ErrorHandler,
	createPurchaseUseCase usecase.CreatePurchaseUseCase,
	updatePurchaseUseCase usecase.UpdatePurchaseUseCase,
	deletePurchaseUseCase usecase.DeletePurchaseUseCase,
	getInvoiceUseCase usecase.GetInvoiceUseCase,
	listInvoicesByMonthUseCase usecase.ListInvoicesByMonthUseCase,
	listInvoicesByMonthPaginatedUseCase usecase.ListInvoicesByMonthPaginatedUseCase,
	listInvoicesByCardUseCase usecase.ListInvoicesByCardUseCase,
	listInvoicesByCardPaginatedUseCase usecase.ListInvoicesByCardPaginatedUseCase,
) *InvoiceHandler {
	return &InvoiceHandler{
		o11y:                                o11y,
		errorHandler:                        errorHandler,
		createPurchaseUseCase:               createPurchaseUseCase,
		updatePurchaseUseCase:               updatePurchaseUseCase,
		deletePurchaseUseCase:               deletePurchaseUseCase,
		getInvoiceUseCase:                   getInvoiceUseCase,
		listInvoicesByMonthUseCase:          listInvoicesByMonthUseCase,
		listInvoicesByMonthPaginatedUseCase: listInvoicesByMonthPaginatedUseCase,
		listInvoicesByCardUseCase:           listInvoicesByCardUseCase,
		listInvoicesByCardPaginatedUseCase:  listInvoicesByCardPaginatedUseCase,
	}
}

// CreatePurchase godoc
//
//	@Summary		Criar compra (invoice item)
//	@Description	Registra uma nova compra no cartão de crédito com suporte a parcelamento.
//	@Description	Para cada parcela é criado um `InvoiceItem` na fatura correspondente ao mês de vencimento.
//	@Description
//	@Description	**Campos:**
//	@Description	- `card_id`: UUID do cartão
//	@Description	- `purchase_date`: data da compra em `YYYY-MM-DD`
//	@Description	- `total_amount`: valor total da compra (ex: `"1200.00"`)
//	@Description	- `currency`: `BRL` | `USD` | `EUR` (opcional, default `BRL`)
//	@Description	- `installment_total`: `1` para à vista, até `48` para parcelado
//	@Tags			invoices
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dtos.PurchaseCreateInput	true	"Dados da compra"
//	@Success		201		{object}	dtos.PurchaseCreateOutput	"Itens de fatura criados (um por parcela)"
//	@Failure		400		{object}	httperrors.ProblemDetail	"Dados inválidos"
//	@Failure		401		{object}	httperrors.ProblemDetail	"Não autenticado"
//	@Failure		404		{object}	httperrors.ProblemDetail	"Cartão ou categoria não encontrados"
//	@Failure		503		{object}	httperrors.ProblemDetail	"Serviço de cartão indisponível"
//	@Failure		500		{object}	httperrors.ProblemDetail	"Erro interno"
//	@Router			/api/v1/invoice-items [post]
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

	if validationErrs := input.Validate(); validationErrs.HasErrors() {
		h.errorHandler.HandleError(w, r, validationErrs)
		return
	}

	output, err := h.createPurchaseUseCase.Execute(ctx, user.ID, input)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	// Phase 2 Fix: Return created items with 201 Created
	responses.JSON(w, http.StatusCreated, output)
}

// UpdatePurchase godoc
//
//	@Summary		Atualizar compra (invoice item)
//	@Description	Atualiza os dados de uma compra. A atualização é propagada para todas as parcelas vinculadas.
//	@Description	O `id` é o ID de qualquer parcela da compra original.
//	@Tags			invoices
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string						true	"ID do item de fatura"	format(uuid)
//	@Param			request	body		dtos.PurchaseUpdateInput	true	"Dados atualizados da compra"
//	@Success		200		{object}	dtos.PurchaseUpdateOutput	"Itens atualizados (todas as parcelas)"
//	@Failure		400		{object}	httperrors.ProblemDetail	"Dados inválidos"
//	@Failure		401		{object}	httperrors.ProblemDetail	"Não autenticado"
//	@Failure		404		{object}	httperrors.ProblemDetail	"Item não encontrado"
//	@Failure		500		{object}	httperrors.ProblemDetail	"Erro interno"
//	@Router			/api/v1/invoice-items/{id} [put]
// UpdatePurchase updates all installments of a purchase.
func (h *InvoiceHandler) UpdatePurchase(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "invoice_handler.update_purchase")
	defer span.End()

	userID, err := middlewares.GetUserFromContext(ctx)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	var input *dtos.PurchaseUpdateInput
	if err = json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	if validationErrs := input.Validate(); validationErrs.HasErrors() {
		h.errorHandler.HandleError(w, r, validationErrs)
		return
	}

	itemID := chi.URLParam(r, "id")
	output, err := h.updatePurchaseUseCase.Execute(ctx, userID.ID, itemID, input)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	// Phase 2 Fix: Return updated items instead of message wrapper
	responses.JSON(w, http.StatusOK, output)
}

// DeletePurchase godoc
//
//	@Summary		Remover compra (invoice item)
//	@Description	Remove uma compra e todas as suas parcelas das faturas. Esta operação é irreversível.
//	@Description	O parâmetro `category_id` é opcional e usado para atualizar o orçamento após a remoção.
//	@Tags			invoices
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id			path	string	true	"ID do item de fatura"		format(uuid)
//	@Param			category_id	query	string	false	"ID da categoria (para atualizar orçamento)"	format(uuid)
//	@Success		204	"Compra removida com sucesso"
//	@Failure		401	{object}	httperrors.ProblemDetail	"Não autenticado"
//	@Failure		404	{object}	httperrors.ProblemDetail	"Item não encontrado"
//	@Failure		500	{object}	httperrors.ProblemDetail	"Erro interno"
//	@Router			/api/v1/invoice-items/{id} [delete]
// DeletePurchase deletes all installments of a purchase.
func (h *InvoiceHandler) DeletePurchase(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "invoice_handler.delete_purchase")
	defer span.End()

	userID, err := middlewares.GetUserFromContext(ctx)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	// Get categoryID from query param (optional, default to empty for now)
	categoryID := r.URL.Query().Get("category_id")

	itemID := chi.URLParam(r, "id")
	if err := h.deletePurchaseUseCase.Execute(ctx, userID.ID, itemID, categoryID); err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	responses.JSON(w, http.StatusNoContent, nil)
}

// GetInvoice godoc
//
//	@Summary		Buscar fatura por ID
//	@Description	Retorna os detalhes completos de uma fatura, incluindo todos os itens com valores por parcela,
//	@Description	rótulo de parcelamento (`installment_label`: ex: `"3/12"` ou `"À vista"`), data de vencimento e total.
//	@Tags			invoices
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string					true	"ID da fatura"	format(uuid)
//	@Success		200	{object}	dtos.InvoiceOutput		"Dados completos da fatura"
//	@Failure		401	{object}	httperrors.ProblemDetail	"Não autenticado"
//	@Failure		404	{object}	httperrors.ProblemDetail	"Fatura não encontrada"
//	@Failure		500	{object}	httperrors.ProblemDetail	"Erro interno"
//	@Router			/api/v1/invoices/{id} [get]
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

// ListInvoicesByMonth lists invoices for a user in a specific month with cursor-based pagination.
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

	// Parse cursor parameters (default: limit=20, max=100)
	params, err := pagination.ParseCursorParams(r, 20, 100)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	output, err := h.listInvoicesByMonthPaginatedUseCase.Execute(ctx, usecase.ListInvoicesByMonthPaginatedInput{
		UserID:         user.ID,
		ReferenceMonth: month,
		Limit:          params.Limit,
		Cursor:         params.Cursor,
	})
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	// Build paginated response
	response := pagination.NewPaginatedResponse(output.Invoices, params.Limit, output.NextCursor)
	responses.JSON(w, http.StatusOK, response)
}

// ListInvoicesByCard lists invoices for a specific card with cursor-based pagination.
func (h *InvoiceHandler) ListInvoicesByCard(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "invoice_handler.list_invoices_by_card")
	defer span.End()

	_, err := middlewares.GetUserFromContext(ctx)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	// Parse cursor parameters (default: limit=10, max=100)
	params, err := pagination.ParseCursorParams(r, 10, 100)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	cardID := chi.URLParam(r, "cardId")
	output, err := h.listInvoicesByCardPaginatedUseCase.Execute(ctx, usecase.ListInvoicesByCardPaginatedInput{
		CardID: cardID,
		Limit:  params.Limit,
		Cursor: params.Cursor,
	})
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	// Build paginated response
	response := pagination.NewPaginatedResponse(output.Invoices, params.Limit, output.NextCursor)
	responses.JSON(w, http.StatusOK, response)
}

// ListInvoices godoc
//
//	@Summary		Listar faturas
//	@Description	Retorna a lista paginada de faturas filtradas por mês ou por cartão.
//	@Description	**Exatamente um** dos parâmetros `month` ou `cardId` deve ser informado.
//	@Description
//	@Description	- **`?month=YYYY-MM`**: lista todas as faturas do usuário no mês especificado (ex: `?month=2025-01`)
//	@Description	- **`?cardId=uuid`**: lista todas as faturas de um cartão específico (ex: `?cardId=abc-123`)
//	@Tags			invoices
//	@Produce		json
//	@Security		BearerAuth
//	@Param			month	query		string	false	"Mês de referência (YYYY-MM). Mutuamente exclusivo com cardId."	example(2025-01)
//	@Param			cardId	query		string	false	"ID do cartão (UUID). Mutuamente exclusivo com month."			format(uuid)
//	@Param			limit	query		integer	false	"Itens por página (default: 20, max: 100)"						minimum(1)	maximum(100)	default(20)
//	@Param			cursor	query		string	false	"Cursor de paginação"
//	@Success		200		{object}	dtos.InvoicePaginatedOutput	"Lista paginada de faturas"
//	@Failure		400		{object}	httperrors.ProblemDetail							"Parâmetros ausentes ou inválidos"
//	@Failure		401		{object}	httperrors.ProblemDetail							"Não autenticado"
//	@Failure		500		{object}	httperrors.ProblemDetail							"Erro interno"
//	@Router			/api/v1/invoices [get]
// ListInvoices unifica listagem por month e por cardId via query params (Change 6).
func (h *InvoiceHandler) ListInvoices(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "invoice_handler.list_invoices")
	defer span.End()

	user, err := middlewares.GetUserFromContext(ctx)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	// Get query parameters
	month := r.URL.Query().Get("month")
	cardID := r.URL.Query().Get("cardId")

	// At least one filter is required
	if month == "" && cardID == "" {
		h.errorHandler.HandleError(w, r, fmt.Errorf("month or cardId parameter is required"))
		return
	}

	// Parse cursor parameters
	params, err := pagination.ParseCursorParams(r, 20, 100)
	if err != nil {
		h.errorHandler.HandleError(w, r, err)
		return
	}

	// Route to appropriate use case based on query params
	if month != "" {
		// List by month
		output, err := h.listInvoicesByMonthPaginatedUseCase.Execute(ctx, usecase.ListInvoicesByMonthPaginatedInput{
			UserID:         user.ID,
			ReferenceMonth: month,
			Limit:          params.Limit,
			Cursor:         params.Cursor,
		})
		if err != nil {
			h.errorHandler.HandleError(w, r, err)
			return
		}

		response := pagination.NewPaginatedResponse(output.Invoices, params.Limit, output.NextCursor)
		responses.JSON(w, http.StatusOK, response)
		return
	}

	if cardID != "" {
		// List by card
		output, err := h.listInvoicesByCardPaginatedUseCase.Execute(ctx, usecase.ListInvoicesByCardPaginatedInput{
			CardID: cardID,
			Limit:  params.Limit,
			Cursor: params.Cursor,
		})
		if err != nil {
			h.errorHandler.HandleError(w, r, err)
			return
		}

		response := pagination.NewPaginatedResponse(output.Invoices, params.Limit, output.NextCursor)
		responses.JSON(w, http.StatusOK, response)
		return
	}
}
