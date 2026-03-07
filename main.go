package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"roamify/config"
	"roamify/internal/models"
	"roamify/internal/router"
	"roamify/seed"
)

func main() {
	// 1. Load .env
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from environment")
	}

	// 2. Load & validate config (crashes early if a required var is missing)
	cfg := config.Load()

	// 3. Connect DB
	db := connectDB(cfg)

	// 4. Migrate
	migrate(db)

	// 5. Seed (dev only)
	if cfg.Env != "production" {
		seed.Run(db)
	}

	// 6. Start server
	r := router.Setup(db, cfg)
	log.Printf("🚀 Roamify API running on http://localhost:%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func connectDB(cfg *config.Config) *gorm.DB {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Africa/Algiers",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort,
	)

	logLevel := logger.Silent
	if cfg.Env != "production" {
		logLevel = logger.Info
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	log.Println("✅ Database connected")
	return db
}

func migrate(db *gorm.DB) {
	err := db.AutoMigrate(
		&models.Wilaya{},
		&models.Place{},
		&models.PlaceImage{},
		&models.PlaceActivity{},
		&models.PlaceSeason{},
		&models.DestinationDetail{},
		&models.HotelDetail{},
		&models.RestaurantDetail{},
		&models.AgencyDetail{},
		&models.User{},
		&models.Review{},
		&models.Bookmark{},
		&models.TravelTip{},
	)
	if err != nil {
		log.Fatalf("❌ Migration failed: %v", err)
	}
	log.Println("✅ Database migrated")
}