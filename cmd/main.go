package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/khadijayo/roamify/config"
	"github.com/khadijayo/roamify/internal/modules/challenges"
	"github.com/khadijayo/roamify/internal/modules/discovery"
	"github.com/khadijayo/roamify/internal/modules/notifications"
	"github.com/khadijayo/roamify/internal/modules/passport"
	"github.com/khadijayo/roamify/internal/modules/posts"
	"github.com/khadijayo/roamify/internal/modules/trips"
	"github.com/khadijayo/roamify/internal/modules/users"
	"github.com/khadijayo/roamify/internal/modules/wishlist"
	"github.com/khadijayo/roamify/pkg/middleware"
)

func main() {

	config.Load()
	config.ConnectDB()
	config.AutoMigrate()

	// 4. Gin release mode in production
	if config.App.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 5. Router with recovery, logging, and CORS
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS())

	// 6. Health check endpoint — used by Railway/Render to confirm the server is up
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "roamify-api"})
	})

	swaggerDir, err := resolveSwaggerDir()
	if err != nil {
		log.Printf("[swagger] static assets unavailable: %v", err)
		r.GET("/swagger", func(c *gin.Context) {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "swagger assets not found on server",
			})
		})
		r.GET("/swagger/*any", func(c *gin.Context) {
			c.JSON(http.StatusNotFound, gin.H{"error": "swagger assets not found"})
		})
	} else {
		log.Printf("[swagger] serving static docs from %s", swaggerDir)
		r.GET("/swagger", func(c *gin.Context) {
			c.Redirect(http.StatusFound, "/swagger/index.html")
		})
		r.Static("/swagger", swaggerDir)
	}


	// 7. Swagger UI (static files in ./docs/swagger/)
	r.GET("/swagger", func(c *gin.Context) {
		c.Redirect(302, "/swagger/index.html")
	})
	r.Static("/swagger", "./docs/swagger")

	// 8. All module routes under /api/v1
	api := r.Group("/api/v1")
	wireModules(api)

	// 9. Start server
	addr := fmt.Sprintf(":%s", config.App.Port)
	log.Printf("[roamify] server starting on %s (env: %s)", addr, config.App.AppEnv)
	if err := r.Run(addr); err != nil {
		log.Fatalf("[roamify] server failed: %v", err)
	}
}


func resolveSwaggerDir() (string, error) {
	var candidates []string

	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates,
			filepath.Join(wd, "docs", "swagger"),
			filepath.Join(wd, "..", "docs", "swagger"),
		)
	}

	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		candidates = append(candidates,
			filepath.Join(exeDir, "docs", "swagger"),
			filepath.Join(exeDir, "..", "docs", "swagger"),
			filepath.Join(exeDir, "..", "..", "docs", "swagger"),
		)
	}

	seen := map[string]struct{}{}
	for _, candidate := range candidates {
		clean := filepath.Clean(candidate)
		if _, ok := seen[clean]; ok {
			continue
		}
		seen[clean] = struct{}{}

		info, err := os.Stat(clean)
		if err == nil && info.IsDir() {
			return clean, nil
		}
	}

	return "", fmt.Errorf("docs/swagger directory not found in runtime paths")
}

func wireModules(api *gin.RouterGroup) {
	db := config.DB
	auth := middleware.Auth(config.App.JWTSecret)

	// ---- Users ----
	userRepo := users.NewRepository(db)
	userSvc := users.NewService(userRepo, config.App.JWTSecret, config.App.JWTExpiryHours)
	userHandler := users.NewHandler(userSvc)
	users.RegisterRoutes(api, userHandler, auth)

	// ---- Notification settings (toggles: trip reminders, squad updates, price drops) ----
	notifSettingsRepo := notifications.NewSettingsRepository(db)
	notifSettingsSvc := notifications.NewSettingsService(notifSettingsRepo)
	notifSettingsHandler := notifications.NewSettingsHandler(notifSettingsSvc)
	notifications.RegisterSettingsRoutes(api, notifSettingsHandler, auth)

	// ---- In-app notifications (inbox, unread count, mark read) ----
	notifRepo := notifications.NewNotificationRepository(db)
	notifSvc := notifications.NewNotificationService(notifRepo)
	notifHandler := notifications.NewNotificationHandler(notifSvc)
	notifications.RegisterNotificationRoutes(api, notifHandler, auth)

	// ---- Trips (CRUD, squad, blueprint, treasury, chat, map pins) ----
	tripRepo := trips.NewRepository(db)
	tripSvc := trips.NewService(tripRepo)
	tripHandler := trips.NewHandler(tripSvc)
	trips.RegisterRoutes(api, tripHandler, auth)

	// ---- Posts (feed, likes, user grid) ----
	postRepo := posts.NewRepository(db)
	postSvc := posts.NewService(postRepo)
	postHandler := posts.NewHandler(postSvc)
	posts.RegisterRoutes(api, postHandler, auth)

	// ---- Wishlist (spots, collections) ----
	wishlistRepo := wishlist.NewRepository(db)
	wishlistSvc := wishlist.NewService(wishlistRepo)
	wishlistHandler := wishlist.NewHandler(wishlistSvc)
	wishlist.RegisterRoutes(api, wishlistHandler, auth)

	// ---- Challenges & gamification (leaderboard, trivia) ----
	challengeRepo := challenges.NewRepository(db)
	challengeSvc := challenges.NewService(challengeRepo, userRepo)
	challengeHandler := challenges.NewHandler(challengeSvc)
	challenges.RegisterRoutes(api, challengeHandler, auth)

	// ---- Passport (vault, stamps) ----
	passportRepo := passport.NewRepository(db)
	passportSvc := passport.NewService(passportRepo, userRepo)
	passportHandler := passport.NewHandler(passportSvc)
	passport.RegisterRoutes(api, passportHandler, auth)

	// ---- Discovery (home, atlas, search, price drops, Gemini AI assistant) ----
	personalizedSvc := discovery.NewPersonalizedService(userRepo)
	discoverySvc := discovery.NewService(config.App.GeminiKey)
	discoveryHandler := discovery.NewHandler(discoverySvc, personalizedSvc)
	discovery.RegisterRoutes(api, discoveryHandler, auth)

	fmt.Println("Registering user routes...")
}

