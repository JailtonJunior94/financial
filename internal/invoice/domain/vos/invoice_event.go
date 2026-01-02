package vos

type InvoiceEvent string

const (
	InvoiceCreated      InvoiceEvent = "invoice_created"
	InvoiceUpdated      InvoiceEvent = "invoice_updated"
	InvoiceItemAdded    InvoiceEvent = "invoice_item_added"
	InvoiceItemRemoved  InvoiceEvent = "invoice_item_removed"
	InvoiceClosed       InvoiceEvent = "invoice_closed"
)

func (e InvoiceEvent) String() string {
	return string(e)
}
