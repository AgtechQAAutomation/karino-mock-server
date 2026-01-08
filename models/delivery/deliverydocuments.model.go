package delivery

import "time"

type DeliverydocumentsList struct {
	CoopID              string `json:"coopId"`
    TempERPSalesOrderId string `json:"tempERPSalesOrderId"`
	ErpSalesOrderId     string `json:"erpSalesOrderId"`
	ErpSalesOrderCode   string `json:"erpSalesOrderCode"`
	SpicSalesOrderId    string `json:"spicSalesOrderId"`
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
}
