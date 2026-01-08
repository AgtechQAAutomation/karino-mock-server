package delivery

type DeliverydocumentsListResponse struct {
	// CoopID              string `json:"coopId"`
    TempERPSalesOrderId string `json:"tempERPSalesOrderId"`
	ErpSalesOrderId     string `json:"erpSalesOrderId"`
	ErpSalesOrderCode   string `json:"erpSalesOrderCode"`
	SpicSalesOrderId    string `json:"spicSalesOrderId"`
	CreatedAt           string `json:"createdAt"`
	UpdatedAt           string `json:"updatedAt"`
}

type ErrorDeliverydocumentsResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}