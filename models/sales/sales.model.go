package sales

import (
	"time"
	"github.com/go-playground/validator/v10"
)

// Order represents the main order payload
type Order struct {
	OrderID                 string       `json:"order_id"`
	OrderNumber             string       `json:"order_number"`
	ContractID              string       `json:"contract_id"`
	FarmerID                string       `json:"farmer_id"`
	FarmerName              string       `json:"farmer_name"`
	ClubID                  string       `json:"club_id"`
	ClubName                string       `json:"club_name"`
	FarmerResourceCategory  string       `json:"farmer_resource_category"`
	ContractCrop            string       `json:"contract_crop"`
	ContractCropVareity     string       `json:"contract_cropVareity"`
	ContractArea            float64      `json:"contractArea"`
	SponsorID               int           `json:"sponsor_id"`
	SponsorName             string        `json:"sponser_name"`
	BuyerID                 int           `json:"buyer_id"`
	BuyerName               string        `json:"buyer_name"`
	PackageSetCaptionPT     string        `json:"package_set_caption_pt"`
	RegionID                int           `json:"region_id"`
	RegionPartID            int           `json:"region_part_id"`
	SettlementID            int           `json:"settlement_id"`
	SettlementPartID        int           `json:"settlement_part_id"`
	CustomZone1ID           int           `json:"custom_zone1_id"`
	CustomZone2ID           int           `json:"custom_zone2_id"`
	PickupDate              time.Time     `json:"pickup_date"`
	CreatedBy               string        `json:"created_by"`
	CreatedAt               time.Time     `json:"created_at"`
	UpdatedAt               time.Time     `json:"updated_at"`
	OrderItems              []OrderItem   `json:"order_items"`
}
// Initialize the validator once for the package
var validate = validator.New()

// ErrorResponse defines the structure for API validation error messages
type ErrorResponse struct {
	Field string `json:"field"`
	Tag   string `json:"tag"`
	Value string `json:"value,omitempty"`
}