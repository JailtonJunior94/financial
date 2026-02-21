package vos

// InvoiceEvent representa os tipos de eventos de domínio de uma fatura.
//
// NOTA: Estes event types são parte de uma interface de domínio planejada para uma
// futura pipeline de eventos de granularidade fina (por fatura/item), separada do
// fluxo atual baseado em PurchaseEvent (por compra).
// Atualmente não estão conectados ao outbox nem a nenhum consumer.
// Não remover — servem de vocabulário canônico para a evolução do domínio.
type InvoiceEvent string

const (
	InvoiceCreated     InvoiceEvent = "invoice_created"
	InvoiceUpdated     InvoiceEvent = "invoice_updated"
	InvoiceItemAdded   InvoiceEvent = "invoice_item_added"
	InvoiceItemRemoved InvoiceEvent = "invoice_item_removed"
	InvoiceClosed      InvoiceEvent = "invoice_closed"
)

func (e InvoiceEvent) String() string {
	return string(e)
}
