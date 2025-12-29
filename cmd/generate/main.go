package main

import (
	"github.com/wpcodevo/golang-fiber-mysql/models"
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
		models.Detail{},
	)

	// Build the type-safe DAO
	g.Execute()
}
