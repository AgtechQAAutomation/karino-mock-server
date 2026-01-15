package main

import (
<<<<<<< HEAD
	"github.com/shyamsundaar/karino-mock-server/models/farmers"
=======
	"github.com/shyamsundaar/karino-mock-server/models/delivery"
	"github.com/shyamsundaar/karino-mock-server/models/deliveryproof"
	models "github.com/shyamsundaar/karino-mock-server/models/farmers"
	"github.com/shyamsundaar/karino-mock-server/models/sales"

>>>>>>> main
	// "github.com/shyamsundaar/karino-mock-server/models/sales"
	"gorm.io/gen"
)

func main() {
	// Initialize the generator
	g := gen.NewGenerator(gen.Config{
		OutPath: "./query", // Path relative to this file
		Mode:    gen.WithDefaultQuery | gen.WithQueryInterface | gen.WithoutContext,
	})

	// Use the structs from your models package
	g.ApplyBasic(
		models.FarmerDetails{},
<<<<<<< HEAD
=======
		sales.SalesOrder{},
		sales.SalesOrderItem{},
		delivery.CreateDeliveryDocuments{},
		deliveryproof.Waybill{},
		deliveryproof.WaybillItem{},
>>>>>>> main
	)

	// Build the type-safe DAO
	g.Execute()
}
