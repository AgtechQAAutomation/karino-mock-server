package deliveryproof

import (
	"time"

	"gorm.io/gorm"
)

type Waybill struct {
	ID                   uint   `gorm:"primaryKey"`
	ContractID           string `gorm:"size:128"`
	OrderID              string `gorm:"size:128;index"`
	RegionID             int    `json:"region_id"`
	RegionPartID         int    `json:"region_part_id"`
	SettlementID         int    `json:"settlement_id"`
	SettlementPartID     int    `json:"settlement_part_id"`
	CustomZone1ID        int    `json:"custom_zone1_id"`
	CustomZone2ID        int    `json:"custom_zone2_id"`
	SalesOrderID         string `gorm:"size:128" json:"sales_order_id"`
	SponsorName          string `gorm:"size:255" json:"sponsor_name"`
	CustomerID           string `gorm:"size:128" json:"customerId"`
	DeliveryNoteID       string `gorm:"size:128" json:"deliveryNoteId"`
	DeliveryNoteDocument string `gorm:"type:text" json:"deliveryNoteDocument"`

	// Relationship: One Waybill has many WaybillItems
	Items []WaybillItem `gorm:"foreignKey:sales_order_id" json:"waybill_items"`

	// For the deliveryPhotos array, it's best stored as JSONB (Postgres) or LongText (MySQL)
	// Or you can create a separate table if you need to query by photo URL.
	DeliveryPhotos string `gorm:"type:jsonb" json:"deliveryPhotos"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (Waybill) TableName() string {
	return "way_bill"
}

type WaybillItem struct {
	ID              uint    `gorm:"primaryKey"`
	SalesOrderID    uint    `gorm:"index"` // Foreign Key
	Name            string  `gorm:"size:255"`
	NumberOfUnits   int     `json:"number_of_units"`
	Quantity        float64 `json:"quantity"`
	QuantityUnitKey string  `json:"quantity_unit_key"`

	// Use decimal for financial precision
	UnitPrice float64 `gorm:"type:decimal(10,2)" json:"unit_price"`
	Price     float64 `gorm:"type:decimal(10,2)" json:"price"`

	PriceUnitKey     string `gorm:"size:10" json:"price_unit_key"` // e.g., "USD"
	Status           string `gorm:"size:50" json:"status"`         // e.g., "DELIVERED"
	StockKeepingUnit string `gorm:"size:100" json:"stock_keeping_unit"`
}

func (WaybillItem) TableName() string {
	return "way_bill_items"
}

type CreateDeliveryDocumentProofSchema struct {
	Waybill      WaybillProof       `json:"waybill"`
	WaybillItems []WaybillItemProof `json:"waybill_items"`
}

type WaybillProof struct {
	ContractID             string `json:"contract_id"`
	OrderID                string `json:"order_id"`
	RegionID               int    `json:"region_id"`
	RegionPartID           int    `json:"region_part_id"`
	SettlementID           int    `json:"settlement_id"`
	SettlementPartID       int    `json:"settlement_part_id"`
	CustomZone1ID          int    `json:"custom_zone1_id"`
	CustomZone2ID          int    `json:"custom_zone2_id"`
	SalesOrderID           string `json:"sales_order_id"`
	SponsorName            string `json:"sponsor_name"`
	CustomerID             string `json:"customerId"`
	DeliveryNoteID         string `json:"deliveryNoteId"`
	DeliveryNoteDocument   string `json:"deliveryNoteDocument"`
	DeliveryPhotoProofURL1 string `json:"url1"`
	DeliveryPhotoProofURL2 string `json:"url2"`
}

type WaybillItemProof struct {
	Name             string  `json:"name"`
	NumberOfUnits    int     `json:"number_of_units"`
	Quantity         float64 `json:"quantity"`
	QuantityUnitKey  string  `json:"quantity_unit_key"`
	UnitPrice        string  `json:"unit_price"`
	Price            string  `json:"price"`
	PriceUnitKey     string  `json:"price_unit_key"`
	Status           string  `json:"status"`
	StockKeepingUnit string  `json:"stock_keeping_unit"`
}

type CreateDocumentdeliveryProofSuccessResponse struct {
	Success bool                                `json:"sucess"`
	Data    CreateDocumentdeliveryProofResponse `json:"data"`
}

type CreateDocumentdeliveryProofResponse struct {
	TempERPProofId string `json:"tempERPProofId"`
	OrderId        string `json:"orderId"`
	Message        string `json:"Message"`
}
