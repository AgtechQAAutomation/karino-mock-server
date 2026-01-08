package products

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Product struct {
	ID              uint    `json:"id" gorm:"primaryKey"`
	ProductCode     string  `json:"product_code" gorm:"size:64;uniqueIndex;not null"`
}

func (s *Product) BeforeCreate(tx *gorm.DB) error {
	if s.ProductCode == "" {
		s.ProductCode = uuid.New().String()
	}
	return nil
}

func (Product) TableName() string {
	return "products"
}