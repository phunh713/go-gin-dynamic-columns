package main

import (
	"flag"
	"fmt"
	"gin-demo/cmd/seed/dynamiccolumn"
	"gin-demo/internal/application/config"
	"log/slog"
	"strings"

	"gorm.io/gorm"
)

func main() {
	// Load config
	configEnv := config.LoadEnv()

	// Connect to database
	db := config.NewDB(configEnv)

	// Parse command-line flags
	domainsFlag := flag.String("domains", "", "Comma-separated list of domains to seed (e.g., companies,invoices)")
	flag.Parse()

	fmt.Println("Starting seed...")

	// Determine which seeds to run
	var domainsToSeed []string
	if *domainsFlag != "" {
		domainsToSeed = strings.Split(*domainsFlag, ",")
		for i := range domainsToSeed {
			domainsToSeed[i] = strings.TrimSpace(domainsToSeed[i])
		}
	}

	// Define all available seed functions
	seedFuncs := map[string]func(*gorm.DB, *slog.Logger){
		"dynamiccolumn": dynamiccolumn.Seed,
		"company":       SeedCompanies,
		"contract":      SeedContracts,
		"invoice":       SeedInvoices,
		"approval":      SeedApprovals,
		"deployment":    SeedDeployments,
		"employee":      SeedEmployees,
	}

	logger := config.NewLogger()
	// If no domains specified, run all seeds
	if len(domainsToSeed) == 0 {
		dynamiccolumn.Seed(db, logger)
		fmt.Println("Dynamic columns seeded successfully")

	} else {
		// Run only specified domains
		for _, domain := range domainsToSeed {
			if seedFunc, exists := seedFuncs[domain]; exists {
				seedFunc(db, logger)
				fmt.Printf("%s seeded successfully\n", domain)
			} else {
				fmt.Printf("Warning: Unknown domain '%s', skipping...\n", domain)
			}
		}
	}

	fmt.Println("Seeding completed!")
}
