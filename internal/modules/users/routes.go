package users

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, h *Handler, auth gin.HandlerFunc) {
	// Public routes
	authRoutes := r.Group("/auth")
	{
		authRoutes.POST("/register", h.Register)
		authRoutes.POST("/login", h.Login)
	}

	// Protected routes
	userRoutes := r.Group("/users")
	userRoutes.Use(auth)
	{
		userRoutes.GET("/me", h.GetMe)
		userRoutes.PATCH("/me", h.UpdateMe)
		userRoutes.GET("/me/vibe", h.GetVibeProfile)
		userRoutes.PUT("/me/vibe", h.UpsertVibeProfile)
	}
}