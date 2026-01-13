package deliveryproof

type ListDeliveryDocumentsResponse struct {
	Data       []DocumentdeliveryProof `json:"data"`
	Pagination PaginationInfo         `json:"pagination"`
}

type DocumentdeliveryProof struct{
	ERPDeliveryDocumentId string `json:"erpDeliveryDocumentId"`
    ERPDeliveryDocumentCode string `json:"erpDeliveryDocumentCode"`
}

type PaginationInfo struct{
    Page uint `json:"page"`
    Limit uint `json:"limit"`
	Total_items uint `json:"total_items"`
	TotalPages uint `json:"total_pages"`
	HasPrevious bool `json:"has_previous"`
	HasNext bool `json:"has_next"`
}