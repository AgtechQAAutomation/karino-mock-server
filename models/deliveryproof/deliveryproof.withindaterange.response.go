package deliveryproof

type DeliveryDocumentDTO struct {
	ERPDeliveryDocumentId   string `json:"erpDeliveryDocumentId"`
	ERPDeliveryDocumentCode string `json:"erpDeliveryDocumentCode"`
}

type PaginationInfo struct {
	Page        int  `json:"page"`
	Limit       int  `json:"limit"`
	TotalItems  int  `json:"total_items"`
	TotalPages  int  `json:"total_pages"`
	HasPrevious bool `json:"has_previous"`
	HasNext     bool `json:"has_next"`
}

type ListDeliveryDocumentsResponse struct {
	Data       []DeliveryDocumentDTO `json:"data"`
	Pagination PaginationInfo        `json:"pagination"`
}

type ErrorDeliveryDocumentProofResponse struct{
	Data       []DeliveryDocumentDTO `json:"data"`
	Pagination PaginationInfo         `json:"pagination"`
}