// Roamify database seeder.
//
// Usage:
//
//	go run cmd/seed/main.go
//
// The command uses the same database environment variables as the API. It runs
// AutoMigrate for the normal app schema plus the seed-only presentation tables,
// clears previous @roamify.demo records, and inserts a fresh demo dataset.
package main

import (
	"log"
	"time"

	"github.com/khadijayo/roamify/config"
	"github.com/khadijayo/roamify/internal/seed"
)

func main() {
	start := time.Now()

	config.Load()
	config.ConnectDB()
	config.AutoMigrate()

	log.Println("[seed] starting Roamify presentation seed")
	if _, err := seed.Run(config.DB); err != nil {
		log.Fatalf("[seed] failed after %s: %v", time.Since(start).Round(time.Millisecond), err)
	}
	log.Printf("[seed] completed in %s", time.Since(start).Round(time.Millisecond))
	log.Printf("[seed] demo users all use password %q", "Roamify2026!")
}
