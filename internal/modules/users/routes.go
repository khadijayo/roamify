package users

import (
	"github.com/gin-gonic/gin"
	"github.com/khadijayo/roamify/pkg/middleware"
)

func RegisterRoutes(rg *gin.RouterGroup, h *Handler) {
	// Public auth routes
	auth := rg.Group("/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
	}

	// Protected user routes
	me := rg.Group("/users", middleware.Auth())
	{
		me.GET("/me", h.GetMe)
		me.PATCH("/me", h.UpdateMe)
		me.GET("/me/vibe", h.GetVibeProfile)
		me.PUT("/me/vibe", h.UpsertVibeProfile)
	}
}