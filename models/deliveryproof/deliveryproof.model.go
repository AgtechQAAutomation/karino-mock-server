package deliveryproof

type CreateDeliveryDocumentProofSchema struct {
	Waybill      WaybillProof       `json:"waybill"`
	WaybillItems []WaybillItemProof `json:"waybill_items"`
}

type WaybillProof struct {
	ContractID             string                `json:"contract_id"`
	OrderID                string                `json:"order_id"`
	RegionID               int                   `json:"region_id"`
	RegionPartID           int                   `json:"region_part_id"`
	SettlementID           int                   `json:"settlement_id"`
	SettlementPartID       int                   `json:"settlement_part_id"`
	CustomZone1ID          int                   `json:"custom_zone1_id"`
	CustomZone2ID          int                   `json:"custom_zone2_id"`
	SalesOrderID           string                `json:"sales_order_id"`
	SponsorName            string                `json:"sponsor_name"`
	CustomerID             string                `json:"customerId"`
	DeliveryNoteID         string                `json:"deliveryNoteId"`
	DeliveryNoteDocument   string                `json:"deliveryNoteDocument"`
	DeliveryPhotos         []DeliveryPhotoProof  `json:"deliveryPhotos"`
}

type DeliveryPhotoProof struct {
	URL1 string `json:"url1"`
	URL2 string `json:"url2"`
}

type WaybillItemProof struct {
	Name              string  `json:"name"`
	NumberOfUnits     int     `json:"number_of_units"`
	Quantity          float64 `json:"quantity"`
	QuantityUnitKey   string  `json:"quantity_unit_key"`
	UnitPrice         string  `json:"unit_price"`
	Price             string  `json:"price"`
	PriceUnitKey      string  `json:"price_unit_key"`
	Status            string  `json:"status"`
	StockKeepingUnit  string  `json:"stock_keeping_unit"`
}


type CreateDocumentdeliveryProofSuccessResponse struct{
  Success bool `json:"sucess"`
  Data  CreateDocumentdeliveryProofResponse `json:"data"`
    
}

type CreateDocumentdeliveryProofResponse struct {
	TempERPProofId  string `json:"tempERPProofId"`
    OrderId string `json:"orderId"`
    Message string `json:"Message"`
}
