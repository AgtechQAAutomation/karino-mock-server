package models

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Detail represents the 'details' table in the database
type Detail struct {
	ID                          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	TempID                      string    `gorm:"not null" json:"tempId"`
	CoopID                      string    `gorm:"not null" json:"coopId"`
	CustomerID                  string    `json:"customerId"`
	VendorID                    string    `json:"vendorId"`
	FarmerID                    string    `gorm:"not null" json:"farmerId"`
	FirstName                   string    `gorm:"not null" json:"firstName"`
	LastName                    string    `gorm:"not null" json:"lastName"`
	MobileNumber                string    `json:"mobile_number"`
	RegionID                    string    `json:"regionId"`
	RegionPartID                string    `json:"regionPartId"`
	SettlementID                string    `json:"settlementId"`
	SettlementPartID            string    `json:"settlementPartId"`
	CustomGeographyStructure1ID int       `json:"custom_geography_structure1_id"`
	CustomGeographyStructure2ID int       `json:"custom_geography_structure2_id"`
	ZipCode                     string    `json:"zipCode"`
	FarmerKycTypeID             string    `json:"farmer_kyc_type_id"`
	FarmerKycType               string    `json:"farmer_kyc_type"`
	FarmerKycID                 string    `json:"farmer_kyc_id"`
	ClubID                      string    `json:"clubId"`
	ClubName                    string    `json:"clubName"`
	ClubLeaderFarmerID          string    `json:"clubLeaderFarmerId"`
	RaithuCreatedDate           string    `json:"raithuCreatedDate"`
	RaithuUpdatedAt             string    `json:"raithuUpdatedAt"`
	CreatedAt                   time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt                   time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
	CustIDUpdateAt              string    `json:"custIdupdateAt"`
	VendorIDUpdateAt            string    `json:"vendorIdupdateAt"`
}

// BeforeCreate Hook to handle any logic before saving to DB
func (d *Detail) BeforeCreate(tx *gorm.DB) (err error) {
	if d.TempID == "" {
		d.TempID = uuid.New().String()
	}
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
func ValidateStruct[T any](payload T) []*ErrorResponse {
	var errors []*ErrorResponse

	// Execute validation
	err := validate.Struct(payload)

	if err != nil {
		// Cast the error to validator.ValidationErrors to access individual field errors
		for _, err := range err.(validator.ValidationErrors) {
			var element ErrorResponse
			element.Field = err.StructNamespace() // e.g., "CreateDetailSchema.FirstName"
			element.Tag = err.Tag()               // e.g., "required"
			element.Value = err.Param()           // e.g., "32" (for min=32)

			errors = append(errors, &element)
		}
	}

	return errors
}

type CreateDetailSchema struct {
	CoopID    string `json:"coopId" validate:"required"`
	FarmerID  string `json:"farmerId" validate:"required"`
	FirstName string `json:"firstName" validate:"required"`
	LastName  string `json:"lastName" validate:"required"`
	// Note: You can add custom validation here to reflect the SQL CHECK constraint
	FarmerKycID        string `json:"farmer_kyc_id" validate:"required_without=ClubLeaderFarmerID"`
	ClubLeaderFarmerID string `json:"clubLeaderFarmerId" validate:"required_without=FarmerKycID"`
}

type UpdateDetailSchema struct {
	FirstName    string `json:"firstName,omitempty"`
	LastName     string `json:"lastName,omitempty"`
	MobileNumber string `json:"mobile_number,omitempty"`
	ZipCode      string `json:"zipCode,omitempty"`
}
