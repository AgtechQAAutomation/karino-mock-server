package models

type FarmerCustomerResponse struct {
	TempERPCustomerID string `json:"tempERPCustomerId"`
	ErpCustomerId     string `json:"erpCustomerId"`
	// ErpVendorId       string `json:"erpVendorId"`
	FarmerId          string `json:"farmerId"`
	CreatedAt         string `json:"createdAt"`
	UpdatedAt         string `json:"updatedAt"`
	// Message           string `json:"message"`
}

type FarmerVendorResponse struct {
	TempERPCustomerID string `json:"tempERPCustomerId"`
	// ErpCustomerId     string `json:"erpCustomerId"`
	ErpVendorId       string `json:"erpVendorId"`
	FarmerId          string `json:"farmerId"`
	CreatedAt         string `json:"createdAt"`
	UpdatedAt         string `json:"updatedAt"`
	// Message           string `json:"message"`
}

// ListFarmersResponse is the top-level list response
type ListFarmersCustomersResponse struct {
	Data       []FarmerCustomerResponse `json:"data"`
	Pagination PaginationInfo   `json:"pagination"`
}

type ListFarmersVendorsResponse struct {
	Data       []FarmerVendorResponse `json:"data"`
	Pagination PaginationInfo   `json:"pagination"`
}

// PaginationInfo matches the required pagination format
type PaginationInfo struct {
	Page         int  `json:"page"`
	Limit        int  `json:"limit"`
	TotalItems   int  `json:"total_items"`
	TotalPages   int  `json:"total_pages"`
	HasPrevious  bool `json:"has_previous"`
	HasNext      bool `json:"has_next"`
}
