package models

type CreateSuccessFarmerResponse struct {
	Success bool           `json:"success"`
	Data    FarmerResponse `json:"data"`
}
type FarmerResponse struct {
	TempERPCustomerID string `json:"tempERPCustomerId"`
	ErpCustomerId     string `json:"erpCustomerId"`
	ErpVendorId       string `json:"erpVendorId"`
	FarmerId          string `json:"farmerId"`
	CreatedAt         string `json:"createdAt"`
	UpdatedAt         string `json:"updatedAt"`
	Message           string `json:"message"`
}

type ErrorFarmerResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
