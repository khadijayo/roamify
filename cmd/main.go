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

	db := config.DB
	err := db.AutoMigrate(

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
	)
	if err != nil {
		log.Fatalf("[migrate] AutoMigrate failed: %v", err)
	}
	log.Println("[migrate] all tables migrated successfully")

	if config.App.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS())

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

	api := r.Group("/api/v1")
	wireModules(api)

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

	authMiddleware := middleware.Auth(config.App.JWTSecret)

	userRepo := users.NewRepository(db)
	userSvc := users.NewService(
		userRepo,
		config.App.JWTSecret,
		config.App.JWTExpiryHours,
	)
	userHandler := users.NewHandler(userSvc)
	users.RegisterRoutes(api, userHandler, authMiddleware)

	notifRepo := notifications.NewRepository(db)
	notifSvc := notifications.NewService(notifRepo)
	notifHandler := notifications.NewHandler(notifSvc)
	notifications.RegisterRoutes(api, notifHandler, authMiddleware)

	tripRepo := trips.NewRepository(db)
	tripSvc := trips.NewService(tripRepo)
	tripHandler := trips.NewHandler(tripSvc)
	trips.RegisterRoutes(api, tripHandler, authMiddleware)

	postRepo := posts.NewRepository(db)
	postSvc := posts.NewService(postRepo)
	postHandler := posts.NewHandler(postSvc)
	posts.RegisterRoutes(api, postHandler, authMiddleware)

	wishlistRepo := wishlist.NewRepository(db)
	wishlistSvc := wishlist.NewService(wishlistRepo)
	wishlistHandler := wishlist.NewHandler(wishlistSvc)
	wishlist.RegisterRoutes(api, wishlistHandler, authMiddleware)

	challengeRepo := challenges.NewRepository(db)
	challengeSvc := challenges.NewService(challengeRepo, userRepo)
	challengeHandler := challenges.NewHandler(challengeSvc)
	challenges.RegisterRoutes(api, challengeHandler, authMiddleware)

	passportRepo := passport.NewRepository(db)
	passportSvc := passport.NewService(passportRepo, userRepo)
	passportHandler := passport.NewHandler(passportSvc)
	passport.RegisterRoutes(api, passportHandler, authMiddleware)

	discoverySvc := discovery.NewService()
	discoveryHandler := discovery.NewHandler(discoverySvc)
	discovery.RegisterRoutes(api, discoveryHandler, authMiddleware)
}
