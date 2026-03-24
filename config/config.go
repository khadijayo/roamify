package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port           string
	DBHost         string
	DBPort         string
	DBUser         string
	DBPassword     string
	DBName         string
	DBSSLMode      string
	JWTSecret      string
	JWTExpiryHours int
	AppEnv         string
}

var App *Config

func Load() {
	if err := godotenv.Load(); err != nil {
		log.Println("[config] no .env file found, reading from environment")
	}

	expiry, _ := strconv.Atoi(getEnv("JWT_EXPIRY_HOURS", "72"))

	App = &Config{
		Port:           getEnv("PORT", "8080"),
		DBHost:         getEnv("DB_HOST", "localhost"),
		DBPort:         getEnv("DB_PORT", "5432"),
		DBUser:         getEnv("DB_USER", "postgres"),
		DBPassword:     getEnv("DB_PASSWORD", ""),
		DBName:         getEnv("DB_NAME", "roamify"),
		DBSSLMode:      getEnv("DB_SSLMODE", "disable"),
		JWTSecret:      getEnv("JWT_SECRET", "change-me-in-production"),
		JWTExpiryHours: expiry,
		AppEnv:         getEnv("APP_ENV", "development"),
	}
}

func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=UTC",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}