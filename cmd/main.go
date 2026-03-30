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
	// 1. Load config & connect to DB
	config.Load()
	config.ConnectDB()

	db := config.DB
	if err := db.AutoMigrate(
		&users.User{},
		&users.VibeProfile{},
		&users.UserFollow{},
		&users.UserPrivacySetting{},
		&notifications.UserNotificationSetting{},
		&trips.Trip{},
		&trips.TripMember{},
		&trips.TripItineraryItem{},
		&trips.TripExpense{},
		&trips.ChatMessage{},
		&posts.Post{},
		&posts.PostTag{},
		&posts.PostLike{},
		&wishlist.WishlistItem{},
		&wishlist.WishlistCollection{},
		&wishlist.WishlistCollectionItem{},
		&challenges.Challenge{},
		&challenges.UserChallengeProgress{},
		&challenges.TriviaQuestion{},
		&challenges.TriviaAttempt{},
		&passport.PassportVaultRecord{},
		&passport.PassportStamp{},
	); err != nil {
		log.Fatalf("[migrate] AutoMigrate failed: %v", err)
	}
	log.Println("[migrate] all tables migrated successfully")

	// 2. Set Gin mode
	if config.App.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 3. Create router with middleware
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS())

	// 4. Health endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "roamify-api"})
	})

	// 5. Swagger static files
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

	// 6. API routes
	api := r.Group("/api/v1")
	wireModules(api)

	// 7. Start server on Render's $PORT or fallback
	port := os.Getenv("PORT")
	if port == "" {
		port = config.App.Port // fallback for local dev
	}
	addr := fmt.Sprintf(":%s", port)
	log.Printf("[roamify] server starting on %s (env: %s)", addr, config.App.AppEnv)
	if err := r.Run(addr); err != nil {
		log.Fatalf("[roamify] server failed: %v", err)
	}
}

// resolveSwaggerDir dynamically finds ./docs/swagger
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
		if info, err := os.Stat(clean); err == nil && info.IsDir() {
			return clean, nil
		}
	}
	return "", fmt.Errorf("docs/swagger directory not found in runtime paths")
}

// wireModules registers all modules
func wireModules(api *gin.RouterGroup) {
	db := config.DB
	auth := middleware.Auth(config.App.JWTSecret)

	userRepo := users.NewRepository(db)
	userSvc := users.NewService(userRepo, config.App.JWTSecret, config.App.JWTExpiryHours)
	userHandler := users.NewHandler(userSvc)
	users.RegisterRoutes(api, userHandler, auth)

	// Notifications: settings + inbox
	notifSettingsRepo := notifications.NewSettingsRepository(db)
	notifSettingsSvc := notifications.NewSettingsService(notifSettingsRepo)
	notifSettingsHandler := notifications.NewSettingsHandler(notifSettingsSvc)
	notifications.RegisterSettingsRoutes(api, notifSettingsHandler, auth)

	notifRepo := notifications.NewNotificationRepository(db)
	notifSvc := notifications.NewNotificationService(notifRepo)
	notifHandler := notifications.NewNotificationHandler(notifSvc)
	notifications.RegisterNotificationRoutes(api, notifHandler, auth)

	tripRepo := trips.NewRepository(db)
	tripSvc := trips.NewService(tripRepo)
	tripHandler := trips.NewHandler(tripSvc)
	trips.RegisterRoutes(api, tripHandler, auth)

	postRepo := posts.NewRepository(db)
	postSvc := posts.NewService(postRepo)
	postHandler := posts.NewHandler(postSvc)
	posts.RegisterRoutes(api, postHandler, auth)

	wishlistRepo := wishlist.NewRepository(db)
	wishlistSvc := wishlist.NewService(wishlistRepo)
	wishlistHandler := wishlist.NewHandler(wishlistSvc)
	wishlist.RegisterRoutes(api, wishlistHandler, auth)

	challengeRepo := challenges.NewRepository(db)
	challengeSvc := challenges.NewService(challengeRepo, userRepo)
	challengeHandler := challenges.NewHandler(challengeSvc)
	challenges.RegisterRoutes(api, challengeHandler, auth)

	passportRepo := passport.NewRepository(db)
	passportSvc := passport.NewService(passportRepo, userRepo)
	passportHandler := passport.NewHandler(passportSvc)
	passport.RegisterRoutes(api, passportHandler, auth)

	personalizedSvc := discovery.NewPersonalizedService(userRepo)
	discoverySvc := discovery.NewService(config.App.GrokKey)
	discoveryHandler := discovery.NewHandler(discoverySvc, personalizedSvc)
	discovery.RegisterRoutes(api, discoveryHandler, auth)

	fmt.Println("Registered all modules")
}