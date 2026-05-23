// cmd/seed/main.go
//
// Roamify database seeder
//
// Usage:
//
//	go run cmd/seed/main.go
//
// Environment variables (same as the main app):
//
//	DATABASE_URL   – full Postgres DSN (takes precedence)
//	DB_HOST / DB_PORT / DB_USER / DB_PASSWORD / DB_NAME / DB_SSLMODE – individual params
//	APP_ENV        – set to "development" for verbose GORM logging
//
// The seeder is fully idempotent. Running it multiple times is safe.
package main

import (
	"log"
	"os"

	"github.com/khadijayo/roamify/config"
	"github.com/khadijayo/roamify/internal/seed"
)

func main() {
	// ── Bootstrap ──────────────────────────────────────────────────────────────
	if err := bootstrap(); err != nil {
		log.Fatalf("[seed] bootstrap failed: %v", err)
	}

	db := config.DB
	log.Println("[seed] 🌱 starting Roamify database seeder…")

	// ── 1. Users ───────────────────────────────────────────────────────────────
	seededUsers, err := seed.SeedUsers(db)
	if err != nil {
		log.Fatalf("[seed] SeedUsers: %v", err)
	}

	// ── 2. Trips ───────────────────────────────────────────────────────────────
	seededTrips, err := seed.SeedTrips(db, seededUsers)
	if err != nil {
		log.Fatalf("[seed] SeedTrips: %v", err)
	}

	// ── 3. Squads ──────────────────────────────────────────────────────────────
	if err := seed.SeedSquads(db, seededUsers, seededTrips); err != nil {
		log.Printf("[seed] SeedSquads (non-fatal): %v", err)
	}

	// ── 4. Comments ────────────────────────────────────────────────────────────
	if err := seed.SeedComments(db, seededUsers, seededTrips); err != nil {
		log.Printf("[seed] SeedComments (non-fatal): %v", err)
	}

	// ── 5. Likes ───────────────────────────────────────────────────────────────
	if err := seed.SeedLikes(db, seededUsers, seededTrips); err != nil {
		log.Printf("[seed] SeedLikes (non-fatal): %v", err)
	}

	log.Println("[seed] ✅ all done – Roamify is ready for demo!")
}

// bootstrap loads config and opens the database connection.
// It deliberately does NOT run migrations – run the main app once first.
func bootstrap() error {
	// Allow overriding the env file path for CI / Docker environments
	if envFile := os.Getenv("ENV_FILE"); envFile != "" {
		if err := os.Setenv("GODOTENV_PATH", envFile); err != nil {
			return err
		}
	}

	config.Load()
	config.ConnectDB()
	return nil
}
