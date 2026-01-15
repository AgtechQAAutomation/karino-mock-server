package models

type CreateSuccessFarmerVendorResponse struct {
	Success bool           `json:"success"`
	Data    CreateFarmerVendorResponse `json:"data"`
}

type CreateSuccessFarmerCustomerResponse struct {
	Success bool           `json:"success"`
	Data    CreateFarmerCustomerResponse `json:"data"`
}

type CreateFarmerCustomerResponse struct {
	TempERPCustomerID string `json:"tempERPCustomerId"`
	ErpCustomerId     string `json:"erpCustomerId"`
	FarmerId          string `json:"farmerId"`
	CreatedAt         string `json:"createdAt"`
	UpdatedAt         string `json:"updatedAt"`
	Message           string `json:"Message"`
}

type CreateFarmerVendorResponse struct {
	TempERPCustomerID string `json:"tempERPCustomerId"`
	ErpVendorId       string `json:"erpVendorId"`
	FarmerId          string `json:"farmerId"`
	CreatedAt         string `json:"createdAt"`
	UpdatedAt         string `json:"updatedAt"`
	Message           string `json:"Message"`
}

type ErrorFarmerResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
