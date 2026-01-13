package models

// FarmerDetailResponse represents the detailed farmer view
type FarmerDetailResponse struct {
	FarmerID           string          `json:"FarmerID"`
	Name               string          `json:"Name"`
	MobileNumber       string          `json:"MobileNumber"`
	Cooperative        string          `json:"Cooperative"`
	SettlementID       int             `json:"SettlementID"`
	SettlementPartID   int             `json:"SettlementPartID"`
	ZipCode            string          `json:"ZipCode"`
	FarmerKycTypeID    int             `json:"FarmerKYCTypeID"`
	FarmerKycType      string          `json:"FarmerKYCType"`
	FarmerKycID        string          `json:"FarmerKYCID"`
	ClubID             string          `json:"ClubID"`
	ClubLeaderFarmerID string          `json:"ClubLeaderFarmerID"`
	Message            string          `json:"Message"`
	EntityID           string          `json:"EntityID"`
	CustomerCode       string          `json:"CustomerCode"`
	VendorCode         string          `json:"VendorCode"`
	CreatedDate        string          `json:"CreatedDate"`
	UpdatedDate        string          `json:"UpdatedDate"`
	BankDetails        BankDetailsInfo `json:"BankDetails"`
}

// Nested object
type BankDetailsInfo struct {
	IBAN  string `json:"IBAN"`
	SWIFT string `json:"SWIFT"`
}
