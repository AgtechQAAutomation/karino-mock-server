package main

import (
	"github.com/shyamsundaar/karino-mock-server/models/farmers"
	"github.com/shyamsundaar/karino-mock-server/models/sales"
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
		sales.SalesOrder{},	
		sales.SalesOrderItem{},
	)

	// Build the type-safe DAO
	g.Execute()
}
