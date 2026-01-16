package deliveryproof

import "time"

type InvoicesResponse struct {
	Invoices []Invoice `json:"invoices"`
}

type Invoice struct {
	ERPInvoiceId   string        `json:"erpInvoiceId"`
	ERPInvoiceCode string        `json:"erpInvoiceCode"`
	ERPInvoiceDate string        `json:"erpInvoiceDate"`
	Items          []InvoiceItem `json:"items"`
}

type InvoiceItem struct {
	ERPItemID         string             `json:"erpItemID"`
	StockKeepingUnit  string             `json:"stock_keeping_unit"`
	Quantity          float64            `json:"quantity"`
	DeliveryNote      InvoiceDeliveryNote `json:"deliveryNote"`
}

type InvoiceDeliveryNote struct {
	TempERPDeliveryNoteId   string  `json:"tempERPDeliveryNoteId"`
	ERPDeliveryDocumentId   string  `json:"erpDeliveryDocumentId"`
	ERPDeliveryDocumentCode string  `json:"erpDeliveryDocumentCode"`
	ERPDeliveryDocumentDate *time.Time  `json:"erpDeliveryDocumentDate"`
	ERPItemID               string  `json:"erpItemID"`
	Quantity                float64 `json:"quantity"`
	OrderItemID             string  `json:"order_item_id"`
}
