package sales

type SalesOrderAmountResponse struct {
	Message             string  `json:"Message"`
	TempERPSalesOrderId string  `json:"tempERPSalesOrderId"`
	ErpSalesOrderId     string  `json:"erpSalesOrderId"`
	ErpSalesOrderCode   string  `json:"erpSalesOrderCode"`
	SpicSalesOrderId    string  `json:"spicSalesOrderId"`
	CreatedAt           string  `json:"createdAt"`
	UpdatedAt           string  `json:"updatedAt"`
	OrderValue          float64 `json:"orderValue"`
	TaxAmount           float64 `json:"taxAmount"`
	TotalAmount         float64 `json:"totalAmount"`
}
