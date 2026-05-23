package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port           string
	DatabaseURL    string
	DBHost         string
	DBPort         string
	DBUser         string
	DBPassword     string
	DBName         string
	DBSSLMode      string
	JWTSecret      string
	JWTExpiryHours int
	AppEnv         string
	GrokKey        string
	AppBaseURL     string
	AdminEmails    []string

	SMTPHost      string
	SMTPPort      string
	SMTPUsername  string
	SMTPPassword  string
	SMTPFromEmail string
	SMTPFromName  string

	// Google OAuth
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
	FrontendURL        string
}

var App *Config

func Load() {
	// Load .env only locally (NOT on Render)
	if os.Getenv("RENDER") == "" {
		if err := godotenv.Load(); err != nil {
			log.Println("[config] no .env file found, reading from system environment")
		}
	}

	expiry, _ := strconv.Atoi(getEnv("JWT_EXPIRY_HOURS", "72"))

	App = &Config{
		Port:           getEnv("PORT", "8080"),
		DatabaseURL:    getEnv("DATABASE_URL", ""),
		DBHost:         getEnv("DB_HOST", "localhost"),
		DBPort:         getEnv("DB_PORT", "5432"),
		DBUser:         getEnv("DB_USER", "postgres"),
		DBPassword:     getEnv("DB_PASSWORD", ""),
		DBName:         getEnv("DB_NAME", "roamify"),
		DBSSLMode:      getEnv("DB_SSLMODE", "disable"),
		JWTSecret:      getEnv("JWT_SECRET", "change-me-in-production"),
		JWTExpiryHours: expiry,
		AppEnv:         getEnv("APP_ENV", "development"),
		// Prefer GROQ_API_KEY for Groq; keep GROK_KEY as backward-compatible fallback.
		GrokKey: getEnv("GROQ_API_KEY", getEnv("GROK_KEY", "")),

		// ✅ CRITICAL: always use env in production
		AppBaseURL: getEnv("APP_BASE_URL", "http://localhost:8080"),
		AdminEmails: parseCSVEnv(
			getEnv("ADMIN_EMAILS", getEnv("ADMIN_EMAIL", "")),
		),

		SMTPHost:      getEnv("SMTP_HOST", ""),
		SMTPPort:      getEnv("SMTP_PORT", "587"),
		SMTPUsername:  getEnv("SMTP_USERNAME", ""),
		SMTPPassword:  getEnv("SMTP_PASSWORD", ""),
		SMTPFromEmail: getEnv("SMTP_FROM_EMAIL", getEnv("SMTP_FROM", "")),
		SMTPFromName:  getEnv("SMTP_FROM_NAME", "Roamify"),

		// Google OAuth
		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL:  getEnv("GOOGLE_REDIRECT_URL", "http://localhost:8080/api/v1/auth/google/callback"),
		FrontendURL:        getEnv("FRONTEND_URL", "https://roamify-zeta.vercel.app"),
	}

	// 🔥 Fail fast in production if missing
	if os.Getenv("RENDER") != "" && App.AppBaseURL == "http://localhost:8080" {
		log.Fatal("❌ APP_BASE_URL must be set in Render environment variables")
	}

	if os.Getenv("RENDER") != "" && App.GoogleClientID == "" {
		log.Println("⚠️  GOOGLE_CLIENT_ID not set – Google OAuth will be disabled")
	}

	// ✅ Debug (remove later if you want)
	log.Println("✅ APP_BASE_URL =", App.AppBaseURL)
}

func (c *Config) DSN() string {
	if c.DatabaseURL != "" {
		return c.DatabaseURL
	}
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

func parseCSVEnv(raw string) []string {
	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		value := strings.ToLower(strings.TrimSpace(part))
		if value != "" {
			values = append(values, value)
		}
	}
	return values
}
