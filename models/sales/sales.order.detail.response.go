package sales
type CreateSalesOrderResponse struct {
	Success bool                        `json:"success"`
	Data    CreateSalesOrderResponseData `json:"data"`
}

type CreateSalesOrderResponseData struct {
	TempERPSalesOrderId string `json:"tempERPSalesOrderId"`
	ErpSalesOrderId     string `json:"erpSalesOrderId"`
	ErpSalesOrderCode   string `json:"erpSalesOrderCode"`
	SpicSalesOrderId    string `json:"spicSalesOrderId"`
	CreatedAt           string `json:"createdAt"`
	UpdatedAt           string `json:"updatedAt"`
	Message             string `json:"Message"`
}
type ErrorSalesOrderResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
