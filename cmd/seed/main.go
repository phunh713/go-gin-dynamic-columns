package main

import (
	"fmt"
	"gin-demo/internal/application/config"
)

func main() {
	// Load config
	configEnv := config.LoadEnv()

	// Connect to database
	db := config.NewDB(configEnv)

	fmt.Println("Starting seed...")

	// Seed dynamic columns
	SeedDynamicColumns(db)

	fmt.Println("Companies seeded successfully")
	fmt.Println("Seeding completed!")
}
