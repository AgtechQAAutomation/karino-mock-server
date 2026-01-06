package models

// ListFarmersResponse is the top-level list response
type ListFarmersResponse struct {
	Data       []FarmerResponse `json:"data"`
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
