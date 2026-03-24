package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/khadijayo/roamify/config"
	"github.com/khadijayo/roamify/internal/modules/challenges"
	"github.com/khadijayo/roamify/internal/modules/notifications"
	"github.com/khadijayo/roamify/internal/modules/passport"
	"github.com/khadijayo/roamify/internal/modules/posts"
	"github.com/khadijayo/roamify/internal/modules/trips"
	"github.com/khadijayo/roamify/internal/modules/users"
	"github.com/khadijayo/roamify/internal/modules/wishlist"
	"github.com/khadijayo/roamify/pkg/middleware"
)

func main() {
	// 1. Load config
	config.Load()

	// 2. Connect DB
	config.ConnectDB()

	// 3. AutoMigrate all models
	db := config.DB
	err := db.AutoMigrate(
		// Users & identity
		&users.User{},
		&users.VibeProfile{},
		&notifications.UserNotificationSetting{},

		// Trips
		&trips.Trip{},
		&trips.TripMember{},
		&trips.TripItineraryItem{},
		&trips.TripExpense{},
		&trips.ChatMessage{},

		// Social
		&posts.Post{},
		&posts.PostTag{},
		&posts.PostLike{},

		// Wishlist
		&wishlist.WishlistItem{},
		&wishlist.WishlistCollection{},
		&wishlist.WishlistCollectionItem{},

		// Gamification
		&challenges.Challenge{},
		&challenges.UserChallengeProgress{},

		// Passport
		&passport.PassportVaultRecord{},
		&passport.PassportStamp{},
	)
	if err != nil {
		log.Fatalf("[migrate] AutoMigrate failed: %v", err)
	}
	log.Println("[migrate] all tables migrated successfully")

	// 4. Setup Gin
	if config.App.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS())

	// 5. Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "roamify-api"})
	})

	// 6. Wire up all modules
	api := r.Group("/api/v1")
	wireModules(api)

	// 7. Start server
	addr := fmt.Sprintf(":%s", config.App.Port)
	log.Printf("[roamify] server starting on %s (env: %s)", addr, config.App.AppEnv)
	if err := r.Run(addr); err != nil {
		log.Fatalf("[roamify] server failed: %v", err)
	}
}

func wireModules(api *gin.RouterGroup) {
	db := config.DB

	// ✅ Create auth middleware ONCE
	authMiddleware := middleware.Auth(config.App.JWTSecret)

	// ---- Users ----
	userRepo := users.NewRepository(db)
	userSvc := users.NewService(
		userRepo,
		config.App.JWTSecret,
		config.App.JWTExpiryHours,
	)
	userHandler := users.NewHandler(userSvc)
	users.RegisterRoutes(api, userHandler, authMiddleware)

	// ---- Notifications ----
	notifRepo := notifications.NewRepository(db)
	notifSvc := notifications.NewService(notifRepo)
	notifHandler := notifications.NewHandler(notifSvc)
	notifications.RegisterRoutes(api, notifHandler, authMiddleware)

	// ---- Trips ----
	tripRepo := trips.NewRepository(db)
	tripSvc := trips.NewService(tripRepo)
	tripHandler := trips.NewHandler(tripSvc)
	trips.RegisterRoutes(api, tripHandler, authMiddleware)

	// ---- Posts ----
	postRepo := posts.NewRepository(db)
	postSvc := posts.NewService(postRepo)
	postHandler := posts.NewHandler(postSvc)
	posts.RegisterRoutes(api, postHandler, authMiddleware)

	// ---- Wishlist ----
	wishlistRepo := wishlist.NewRepository(db)
	wishlistSvc := wishlist.NewService(wishlistRepo)
	wishlistHandler := wishlist.NewHandler(wishlistSvc)
	wishlist.RegisterRoutes(api, wishlistHandler, authMiddleware)

	// ---- Challenges ----
	challengeRepo := challenges.NewRepository(db)
	challengeSvc := challenges.NewService(challengeRepo, userRepo)
	challengeHandler := challenges.NewHandler(challengeSvc)
	challenges.RegisterRoutes(api, challengeHandler, authMiddleware)

	// ---- Passport ----
	passportRepo := passport.NewRepository(db)
	passportSvc := passport.NewService(passportRepo, userRepo)
	passportHandler := passport.NewHandler(passportSvc)
	passport.RegisterRoutes(api, passportHandler, authMiddleware)
}
