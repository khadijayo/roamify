package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	// Server
	Port string
	Env  string

	// Database
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	// JWT
	JWTSecret          string
	JWTExpiryHours     int
	RefreshTokenExpiry int

	// Rate Limiting
	RateLimitRequests int
	RateLimitWindow   int // seconds

	// Security
	AllowedOrigins []string
	BcryptCost     int
}

func Load() *Config {
	cfg := &Config{
		Port:               getEnv("PORT", "8080"),
		Env:                getEnv("APP_ENV", "development"),
		DBHost:             mustGetEnv("DB_HOST"),
		DBPort:             getEnv("DB_PORT", "5432"),
		DBUser:             mustGetEnv("DB_USER"),
		DBPassword:         mustGetEnv("DB_PASSWORD"),
		DBName:             mustGetEnv("DB_NAME"),
		JWTSecret:          mustGetEnv("JWT_SECRET"),
		JWTExpiryHours:     getEnvInt("JWT_EXPIRY_HOURS", 72),
		RefreshTokenExpiry: getEnvInt("REFRESH_TOKEN_EXPIRY_DAYS", 7),
		RateLimitRequests:  getEnvInt("RATE_LIMIT_REQUESTS", 100),
		RateLimitWindow:    getEnvInt("RATE_LIMIT_WINDOW_SECONDS", 60),
		BcryptCost:         getEnvInt("BCRYPT_COST", 12),
		AllowedOrigins: []string{
			getEnv("FRONTEND_URL", "http://localhost:5173"),
		},
	}
	return cfg
}

func mustGetEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("❌ Required environment variable %q is not set", key)
	}
	return val
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		log.Printf("⚠️  Invalid value for %s, using default %d", key, fallback)
		return fallback
	}
	return i
}