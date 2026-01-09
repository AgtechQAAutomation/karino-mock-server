package delivery

import (
	"time"
	"gorm.io/gorm"
)

type CreateDeliveryDocuments struct {
	Id 				uint   `json:"id" gorm:"primaryKey"`
	CoopID	  	string `json:"coop_id" gorm:"size:64;index;not null"`
	ErpSalesOrderCode string `gorm:"column:erp_sales_order_code;size:64" json:"erp_sales_order_code"`
	OrderID 	string `json:"order_id" gorm:"size:64;index;not null"`
	DeliveryDocumentID  	string `json:"delivery_document_id" gorm:"size:64;index;not null"`
	DeliveryDocumentCode  	string `json:"delivery_document_code" gorm:"size:64;index;not null"`
	OrderItemID 	string `json:"order_item_id" gorm:"size:64;index;not null"`
	CreatedAt  	*time.Time `json:"created_at"`
	UpdatedAt  	*time.Time `json:"updated_at"`
}


func (CreateDeliveryDocuments) TableName() string {
	return "delivery_documents"
}

// BeforeCreate Hook to handle any logic before saving to DB
func (d *CreateDeliveryDocuments) BeforeCreate(tx *gorm.DB) (err error) {
	var now = time.Now()
	d.CreatedAt = &now
	d.UpdatedAt = &now
	return nil
}

type CreateDeliveryDocumentSchema struct {
	ErpSalesOrderCode string `gorm:"column:erp_sales_order_code;size:64" json:"erp_sales_order_code"`
	OrderID 	string `json:"order_id" gorm:"size:64;index;not null"`
	NoofDeliveryDocuments  	int `json:"no_of_delivery_documents"`
}

type CreateDeliveryDocumentSuccessResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}