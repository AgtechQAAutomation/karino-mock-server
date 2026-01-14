package deliveryproof

type DocumentdeliveryProof struct{
	ERPDeliveryDocumentId string `json:"erpDeliveryDocumentId"`
    ERPDeliveryDocumentCode string `json:"erpDeliveryDocumentCode"`
}

type ListDeliveryDocumentsResponse struct {
	Data       []DocumentdeliveryProof `json:"data"`
	Pagination PaginationInfo         `json:"pagination"`
}

type PaginationInfo struct{
    Page int `json:"page"`
    Limit int `json:"limit"`
	TotalItems int `json:"total_items"`
	TotalPages int `json:"total_pages"`
	HasPrevious bool `json:"has_previous"`
	HasNext bool `json:"has_next"`
}
type ErrorDeliveryDocumentProofResponse struct{
	Data       []DocumentdeliveryProof `json:"data"`
	Pagination PaginationInfo         `json:"pagination"`
}