package middleware

import (
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORS sets cross-origin headers.
//
// In production (RENDER env var is set) only the Vercel frontend origin is
// allowed. Locally, all origins are permitted so curl / Postman work without
// extra setup.
func CORS() gin.HandlerFunc {
	// Comma-separated list of allowed origins from env, e.g.:
	//   ALLOWED_ORIGINS=https://roamify-zeta.vercel.app,https://www.roamify.app
	allowedOrigins := parseOrigins(
		os.Getenv("ALLOWED_ORIGINS"),
		"https://roamify-zeta.vercel.app", // hard-coded safe default
	)

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		if isAllowed(origin, allowedOrigins) {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
		} else if os.Getenv("RENDER") == "" {
			// Local dev: allow everything so Postman / curl work fine
			c.Header("Access-Control-Allow-Origin", "*")
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		// Allow cookies for the OAuth state token
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func parseOrigins(envValue, defaultOrigin string) []string {
	origins := []string{defaultOrigin}
	for _, o := range strings.Split(envValue, ",") {
		o = strings.TrimSpace(o)
		if o != "" && o != defaultOrigin {
			origins = append(origins, o)
		}
	}
	return origins
}

func isAllowed(origin string, allowed []string) bool {
	for _, a := range allowed {
		if strings.EqualFold(origin, a) {
			return true
		}
	}
	return false
}
