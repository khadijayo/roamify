package main

import (
	"fmt"
	"log"
	"os"

	"roamify/internal/models" 
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"github.com/joho/godotenv"
)

var DB *gorm.DB

func ConnectDB() *gorm.DB {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Africa/Algiers",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("Database connected")
	return db
}

func AutoMigrate(db *gorm.DB) {
	log.Println("⏳ Running AutoMigrate...")

	err := db.AutoMigrate(
		
		&models.User{},
		&models.City{},
		&models.Category{},
		&models.Place{},
		&models.PlaceImage{},
		&models.Review{},
		&models.Favorite{},

		&models.Trip{},
		&models.TripPlace{},
		&models.SavedTrip{},

		//&models.Agency{},
		//&models.AgencyPackage{},

		&models.TravelGroup{},
		&models.GroupMember{},
		&models.GroupMessage{},
	)

	if err != nil {
		log.Fatal("AutoMigrate failed:", err)
	}

	log.Println(" AutoMigrate completed successfully")
}

func main() {
	
	godotenv.Load()  

	DB = ConnectDB()
	AutoMigrate(DB)

}