package models

// FarmerDetailResponse represents the detailed farmer view
type FarmerDetailResponse struct {
	FarmerID                    string     `json:"farmerId"`
	Name                   string     		`json:"name"`
	MobileNumber                string     `json:"mobile_number"`
	Cooperative				 	string     `json:"cooperative"`
	SettlementID                int        `json:"settlementId"`
	SettlementPartID            int        `json:"settlementPartId"`
	ZipCode                     string     `json:"zipCode"`
	FarmerKycTypeID             int        `json:"farmer_kyc_type_id"`
	FarmerKycType               string     `json:"farmer_kyc_type"`
	FarmerKycID                 string     `json:"farmer_kyc_id"`
	ClubID                      string     `json:"clubId"`
	ClubLeaderFarmerID          string     `json:"clubLeaderFarmerId"`
	Message		  				string 	   `json:"message"`
	EntityID					string     `json:"entityId"`
	CustomerCode				string     `json:"customerCode"`
	VendorCode					string     `json:"vendorCode"`
	CreatedDate         			string 	   `json:"createdAt"`
	UpdatedDate         			string 	   `json:"updatedAt"`
	// BankDetails         BankDetailsInfo `json:"bankDetails"`
}

// Nested object
// type BankDetailsInfo struct {
// 	IBAN  string `json:"IBAN"`
// 	SWIFT string `json:"SWIFT"`
// }
