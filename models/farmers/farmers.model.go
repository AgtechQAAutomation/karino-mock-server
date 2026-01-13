package models

import (
	"time"

	"github.com/go-playground/validator/v10"
	// "github.com/google/uuid"
	"strconv"
	"gorm.io/gorm"
)

// Detail represents the 'details' table in the database
type FarmerDetails struct {
	ID                          uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	TempID                      string     `gorm:"not null" json:"tempId"`
	CoopID                      string     `gorm:"not null" json:"coopId"`
	CustomerID                  string     `json:"customerId"`
	VendorID                    string     `json:"vendorId"`
	FarmerID                    string     `gorm:"not null" json:"farmerId"`
	FirstName                   string     `gorm:"not null" json:"firstName"`
	LastName                    string     `gorm:"not null" json:"lastName"`
	MobileNumber                string     `json:"mobile_number"`
	RegionID                    int        `json:"regionId"`
	RegionPartID                int        `json:"regionPartId"`
	SettlementID                int        `json:"settlementId"`
	SettlementPartID            int        `json:"settlementPartId"`
	CustomGeographyStructure1ID int     `json:"custom_geography_structure1_id"`
	CustomGeographyStructure2ID int     `json:"custom_geography_structure2_id"`
	ZipCode                     string     `json:"zipCode"`
	FarmerKycTypeID             int        `json:"farmer_kyc_type_id"`
	FarmerKycType               string     `json:"farmer_kyc_type"`
	FarmerKycID                 string     `json:"farmer_kyc_id"`
	ClubID                      string     `json:"clubId"`
	ClubName                    string     `json:"clubName"`
	ClubLeaderFarmerID          string     `json:"clubLeaderFarmerId" `
	RaithuCreatedDate           string     `json:"createdDate" gorm:"default:null"`
	RaithuUpdatedAt             string     `json:"updatedAt" gorm:"default:null"`
	CreatedAt                   time.Time  `gorm:"default:null" `
	UpdatedAt                   time.Time  `gorm:"default:null"`
	CustIDUpdateAt              *time.Time `gorm:"default:null"`
	VendorIDUpdateAt            *time.Time `gorm:"default:null"`
}

// BeforeCreate Hook to handle any logic before saving to DB
func (d *FarmerDetails) BeforeCreate(tx *gorm.DB) (err error) {
	now := time.Now()

	// Fetch last TempID
	var lastTempID string

	err = tx.
		Model(&FarmerDetails{}).
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
	d.CreatedAt = now
	d.UpdatedAt = now

	return nil
}


// Initialize the validator once for the package
var validate = validator.New()

// ErrorResponse defines the structure for API validation error messages
type ErrorResponse struct {
	Field string `json:"field"`
	Tag   string `json:"tag"`
	Value string `json:"value,omitempty"`
}

// ValidateStruct is a generic function that validates any struct against 'validate' tags
// It returns a slice of ErrorResponse pointers if validation fails
// func ValidateStruct[T any](payload T) []*ErrorResponse {
// 	var errors []*ErrorResponse

// 	// Execute validation
// 	err := validate.Struct(payload)

// 	if err != nil {
// 		// Cast the error to validator.ValidationErrors to access individual field errors
// 		for _, err := range err.(validator.ValidationErrors) {
// 			var element ErrorResponse
// 			element.Field = err.StructNamespace() // e.g., "CreateDetailSchema.FirstName"
// 			element.Tag = err.Tag()               // e.g., "required"
// 			element.Value = err.Param()           // e.g., "32" (for min=32)

// 			errors = append(errors, &element)
// 		}
// 	}

// 	return errors
// }

// CreateDetailSchema represents request body
// swagger:model CreateDetailSchema
type CreateDetailSchema struct {
	FarmerID           string `json:"farmerId" example:"string"`
	FirstName          string `json:"firstName" example:"string"`
	LastName           string `json:"lastName" example:"string"`
	MobileNumber       string `json:"mobile_number" example:"string"`
	RegionID           int    `json:"regionId" example:"0"`
	RegionPartID       int    `json:"regionPartID" example:"0"`
	SettlementID       int    `json:"settlementID" example:"0"`
	SettlementPartID   int    `json:"settlementPartID" example:"0"`
	CustomGeo1ID       int `json:"custom_geography_structure1_id" example:"0"`
	CustomGeo2ID       int `json:"custom_geography_structure2_id" example:"0"`
	ZipCode            string `json:"ZipCode" example:"string"`
	FarmerKycTypeID    int    `json:"farmer_kyc_type_id" example:"0"`
	FarmerKycType      string `json:"farmer_kyc_type" example:"string"`
	FarmerKycID        string `json:"farmer_kyc_id" example:"string"`
	ClubID             string `json:"clubId" example:"string"`
	ClubName           string `json:"clubName" example:"string"`
	ClubLeaderFarmerID string `json:"clubLeaderFarmerId" example:"string"`
	RaithuCreatedDate  string `json:"createdDate" example:"2025-12-30T05:03:17.863Z"`
	RaithuUpdatedAt    string `json:"updatedAt" example:"2025-12-30T05:03:17.863Z"`
}
