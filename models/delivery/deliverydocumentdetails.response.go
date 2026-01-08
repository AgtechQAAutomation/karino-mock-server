package delivery

import "time"

type DeliveryNotesResponse struct {
	DeliveryNotes []DeliveryNote `json:"deliveryNotes"`
}

type DeliveryNote struct {
	ERPDeliveryDocumentId   string    `json:"erpDeliveryDocumentId"`
	ERPDeliveryDocumentCode string    `json:"erpDeliveryDocumentCode"`
	ERPDeliveryDocumentDate time.Time `json:"erpDeliveryDocumentDate"`
	Items                   []DeliveryItem `json:"items"`
}

type DeliveryItem struct {
	ERPItemID          string            `json:"erpItemID"`
	StockKeepingUnit   string            `json:"stock_keeping_unit"`
	Quantity           float64           `json:"quantity"`
	SalesOrder         DeliverySalesOrder `json:"salesOrder"`
}

type DeliverySalesOrder struct {
	TempERPSalesOrderId string `json:"tempERPSalesOrderId"`
	ERPSalesOrderId     string `json:"erpSalesOrderId"`
	ERPSalesOrderCode   string `json:"erpSalesOrderCode"`
	SPICSalesOrderId    string `json:"spicSalesOrderId"`
	ERPItemID           string `json:"erpItemID"`
	OrderItemID         string `json:"order_item_id"`
}
