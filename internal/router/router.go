package router

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"roamify/config"
	"roamify/internal/handlers"
	"roamify/internal/middleware"
)

func Setup(db *gorm.DB, cfg *config.Config) *gin.Engine {
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// ── CORS ─────────────────────────────────────────────────────
	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// ── Health check ─────────────────────────────────────────────
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "roamify-api"})
	})

	// ── Handlers ─────────────────────────────────────────────────
	placeH    := handlers.NewPlaceHandler(db)
	wilayaH   := handlers.NewWilayaHandler(db)
	authH     := handlers.NewAuthHandler(db, cfg)
	reviewH   := handlers.NewReviewHandler(db)
	bookmarkH := handlers.NewBookmarkHandler(db)
	tipH      := handlers.NewTipHandler(db)

	// ── API v1 ───────────────────────────────────────────────────
	v1 := r.Group("/api/v1")
	{
		// Public: Wilayas
		w := v1.Group("/wilayas")
		{
			w.GET("", wilayaH.List)
			w.GET("/:id", wilayaH.GetByID)
		}

		// Public: Places
		p := v1.Group("/places")
		{
			p.GET("", placeH.List)            
			p.GET("/featured", placeH.Featured)
			p.GET("/:slug", placeH.GetBySlug)
			p.GET("/:slug/reviews", reviewH.List)
		}

		// Public: Tips
		t := v1.Group("/tips")
		{
			t.GET("", tipH.List)
			t.GET("/:slug", tipH.GetBySlug)
		}

		// Public: Auth
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authH.Register)
			auth.POST("/login", authH.Login)
		}

		// Protected (JWT required)
		protected := v1.Group("")
		protected.Use(middleware.AuthRequired(cfg))
		{
			protected.POST("/auth/logout", authH.Logout)
			protected.POST("/places/:slug/reviews", reviewH.Create)

			bk := protected.Group("/user/bookmarks")
			{
				bk.GET("", bookmarkH.List)
				bk.POST("/:placeID", bookmarkH.Add)
				bk.DELETE("/:placeID", bookmarkH.Remove)
			}
		}
	}

	return r
}