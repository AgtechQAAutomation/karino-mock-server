package sales

type SalesOrderListResponse struct {
	TempERPSalesOrderId string `json:"tempERPSalesOrderId"`
	ErpSalesOrderId     string `json:"erpSalesOrderId"`
	ErpSalesOrderCode   string `json:"erpSalesOrderCode"`
	SpicSalesOrderId    string `json:"spicSalesOrderId"`
	CreatedAt           string `json:"created_at"`
	UpdatedAt           string `json:"updated_at"`
}

type ListSalesOrderResponse struct {
	Data       []SalesOrderListResponse `json:"data"`
	Pagination PaginationInfo           `json:"pagination"`
}

// PaginationInfo matches the required pagination format
type PaginationInfo struct {
	Page        int  `json:"page"`
	Limit       int  `json:"limit"`
	TotalItems  int  `json:"total_items"`
	TotalPages  int  `json:"total_pages"`
	HasPrevious bool `json:"has_previous"`
	HasNext     bool `json:"has_next"`
}
