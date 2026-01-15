package initializers

import (
	"log"

	"github.com/AgtechQAAutomation/karino-mock-server/models/products"
	"gorm.io/gorm"
)

func SeedInitialData(db *gorm.DB) {

	var count int64

	// 🔹 Check if products already exist
	db.Model(&products.Product{}).Count(&count)
	if count > 0 {
		log.Println("ℹ️ Products already seeded, skipping")
		return
	}

	// 🔹 Seed 10 Products
	productList := []products.Product{
		{ProductCode: "IIT-101"},
		{ProductCode: "IIT-102"},
		{ProductCode: "IIT-103"},
		{ProductCode: "IIT-104"},
		{ProductCode: "IIT-105"},
		{ProductCode: "IIT-106"},
		{ProductCode: "IIT-107"},
		{ProductCode: "IIT-108"},
		{ProductCode: "IIT-109"},
		{ProductCode: "IIT-110"},
	}

	if err := db.Create(&productList).Error; err != nil {
		log.Fatal("❌ Failed to seed products:", err)
	}

	log.Println("✅ Products seeded successfully")
}
