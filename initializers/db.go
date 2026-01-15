package initializers

import (
	"fmt"
	"log"
	"os"

<<<<<<< HEAD
	"github.com/shyamsundaar/karino-mock-server/models/farmers"
=======
	"github.com/shyamsundaar/karino-mock-server/models/delivery"
	"github.com/shyamsundaar/karino-mock-server/models/deliveryproof"
	models "github.com/shyamsundaar/karino-mock-server/models/farmers"
	"github.com/shyamsundaar/karino-mock-server/models/products"
>>>>>>> main
	"github.com/shyamsundaar/karino-mock-server/models/sales"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func ConnectDB(config *Config) {
	var err error
	// dsn := fmt.Sprintf("user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=UTC")
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=UTC", config.DBUserName, config.DBUserPassword, config.DBHost, config.DBPort, config.DBName)

	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to the Database! \n", err.Error())
		os.Exit(1)
	}

	DB.Logger = logger.Default.LogMode(logger.Info)

	log.Println("Running Migrations")
<<<<<<< HEAD
	DB.AutoMigrate(&models.FarmerDetails{},&sales.SalesOrder{},&sales.SalesOrderItem{})
=======
	err = DB.AutoMigrate(&models.FarmerDetails{}, &sales.SalesOrder{}, &sales.SalesOrderItem{}, &products.Product{},
		&delivery.CreateDeliveryDocuments{},
		&deliveryproof.Waybill{}, &deliveryproof.WaybillItem{})
	SeedInitialData(DB)
	StartExpirationWorker(DB)

	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
>>>>>>> main

	log.Println("🚀 Connected Successfully to the Database")
}
