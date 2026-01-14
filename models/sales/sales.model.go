package sales

import (
	"time"
"strconv"
	// "github.com/go-playground/validator/v10"
	//"github.com/google/uuid"

	"gorm.io/gorm"
)

//
// =======================
// SalesOrder DB MODEL
// =======================
//

type SalesOrder struct {
	ID     uint   `gorm:"primaryKey;autoIncrement"`
	TempID string `gorm:"column:temp_id;not null" json:"tempId"`
	CoopID string `gorm:"column:coop_id;not null" json:"coopId"`

	ErpSalesOrderId   string `gorm:"column:erp_sales_order_id;size:64" json:"erp_sales_order_id"`
	ErpSalesOrderCode string `gorm:"column:erp_sales_order_code;size:64" json:"erp_sales_order_code"`

	OrderID     string `gorm:"column:order_id;size:64;uniqueIndex" json:"order_id"`
	OrderNumber string `gorm:"column:order_number;size:64" json:"order_number"`
	ContractID  string `gorm:"column:contract_id;size:64" json:"contract_id"`

	FarmerID   string `gorm:"column:farmer_id;size:64" json:"farmer_id"`
	FarmerName string `gorm:"column:farmer_name;size:128" json:"farmer_name"`

	ClubID   string `gorm:"column:club_id;size:64" json:"club_id"`
	ClubName string `gorm:"column:club_name;size:128" json:"club_name"`

	FarmerResourceCategory string `gorm:"column:farmer_resource_category;size:64" json:"farmer_resource_category"`

	ContractCrop        string  `gorm:"column:contract_crop;size:64" json:"contract_crop"`
	ContractCropVareity string  `gorm:"column:contract_crop_vareity;size:64" json:"contract_cropVareity"`
	ContractArea        float64 `gorm:"column:contract_area" json:"contractArea"`

	SponsorID   int    `gorm:"column:sponsor_id" json:"sponsor_id"`
	SponsorName string `gorm:"column:sponser_name;size:128" json:"sponser_name"`

	BuyerID   int    `gorm:"column:buyer_id" json:"buyer_id"`
	BuyerName string `gorm:"column:buyer_name;size:128" json:"buyer_name"`

	PackageSetCaptionPT string `gorm:"column:package_set_caption_pt;size:128" json:"package_set_caption_pt"`

	RegionID         int `gorm:"column:region_id" json:"region_id"`
	RegionPartID     int `gorm:"column:region_part_id" json:"region_part_id"`
	SettlementID     int `gorm:"column:settlement_id" json:"settlement_id"`
	SettlementPartID int `gorm:"column:settlement_part_id" json:"settlement_part_id"`

	CustomZone1ID int `gorm:"column:custom_zone1_id" json:"custom_zone1_id"`
	CustomZone2ID int `gorm:"column:custom_zone2_id" json:"custom_zone2_id"`

	PickupDate string `gorm:"column:pickup_date;default:null" json:"pickup_date"`
	CreatedBy  string `gorm:"column:created_by;size:64" json:"created_by"`

	CreatedAt *time.Time `gorm:"default:null"`
	UpdatedAt *time.Time `gorm:"default:null"`

	NoofOrderItems int              `gorm:"column:noof_order_items" json:"noofOrderItems"`
	OrderItems     []SalesOrderItem `gorm:"foreignKey:OrderID;references:OrderID" json:"order_items"`
}

func (SalesOrder) TableName() string {
	return "sales_orders"
}

//
// =======================
// GORM HOOK
// =======================
//

// BeforeCreate Hook to handle any logic before saving to DB
func (d *SalesOrder) BeforeCreate(tx *gorm.DB) (err error) {
	now := time.Now()

	// Fetch last TempID
	var lastTempID string

	err = tx.
		Model(&SalesOrder{}).
		Select("temp_id").
		Where("temp_id IS NOT NULL AND temp_id != ''").
		Order("id DESC").
		Limit(1).
		Scan(&lastTempID).Error

	// Default start value
	next := 1000

	if err == nil && lastTempID != "" {
		if n, convErr := strconv.Atoi(lastTempID); convErr == nil {
			next = n + 1
		}
	}

	d.TempID = strconv.Itoa(next)
	d.CreatedAt = &now
	d.UpdatedAt = &now

	return nil
}

//
// =======================
// SalesOrderItem MODEL
// =======================
//

type SalesOrderItem struct {
	ID      uint   `gorm:"primaryKey;autoIncrement"`
	OrderID string `gorm:"column:order_id;size:64;index;not null" json:"order_id"`

	OrderItemID     string `gorm:"column:order_item_id;size:64" json:"order_item_id"`
	OrderItemNumber string `gorm:"column:order_item_number;size:64" json:"order_item_number"`
	ErpItemID		string `gorm:"column:erp_item_id;size:64" json:"erp_item_id"`
	ErpItemID2		string `gorm:"column:erp_item_id_2;size:64" json:"erp_item_id_2"`

	StockKeepingUnit string `gorm:"column:stock_keeping_unit;size:64" json:"stock_keeping_unit"`
	ProductGroup     string `gorm:"column:product_group;size:64" json:"product_group"`

	InputItemID          string `gorm:"column:input_item_id;size:64" json:"input_item_id"`
	InputItemName        string `gorm:"column:input_item_name;size:128" json:"input_item_name"`
	InputItemNameCaption string `gorm:"column:input_item_name_caption;size:128" json:"input_item_name_caption"`

	Quantity        float64 `gorm:"column:quantity" json:"quantity"`
	QuantityUnitKey string  `gorm:"column:quantity_unit_key;size:32" json:"quantity_unit_key"`

	UnitPrice    float64 `gorm:"column:unit_price" json:"unit_price"`
	Price        string  `gorm:"column:price;size:32" json:"price"`
	PriceUnitKey string  `gorm:"column:price_unit_key;size:32" json:"price_unit_key"`

	NumberOfUnits int `gorm:"column:number_of_units" json:"number_of_units"`
}

func (SalesOrderItem) TableName() string {
	return "sales_order_items"
}

//
// =======================
// VALIDATION + ERRORS
// =======================
//

// var validate = validator.New()

type ErrorResponse struct {
	Field string `json:"field"`
	Tag   string `json:"tag"`
	Value string `json:"value,omitempty"`
}

//
// =======================
// CREATE REQUEST SCHEMA
// (Equivalent to CreateDetailSchema)
// =======================
//

type CreateSalesOrderSchema struct {
	OrderID     string `json:"order_id"`
	OrderNumber string `json:"order_number"`
	ContractID  string `json:"contract_id"`

	FarmerID   string `json:"farmer_id"`
	FarmerName string `json:"farmer_name"`

	ClubID   string `json:"club_id"`
	ClubName string `json:"club_name"`

	FarmerResourceCategory string `json:"farmer_resource_category"`

	ContractCrop        string  `json:"contract_crop"`
	ContractCropVareity string  `json:"contract_cropVareity"`
	ContractArea        float64 `json:"contractArea"`

	SponsorID   int    `json:"sponsor_id"`
	SponsorName string `json:"sponser_name"`

	BuyerID   int    `json:"buyer_id"`
	BuyerName string `json:"buyer_name"`

	PackageSetCaptionPT string `json:"package_set_caption_pt"`

	RegionID         int `json:"region_id"`
	RegionPartID     int `json:"region_part_id"`
	SettlementID     int `json:"settlement_id"`
	SettlementPartID int `json:"settlement_part_id"`

	CustomZone1ID int `json:"custom_zone1_id"`
	CustomZone2ID int `json:"custom_zone2_id"`

	PickupDate string `json:"pickup_date"`
	CreatedBy  string `json:"created_by"`

	OrderItems []SalesOrderItem `json:"order_items"`
}
