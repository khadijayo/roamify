package auth

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.RouterGroup, h *Handler) {
	authRoutes := r.Group("/auth")
	{
		authRoutes.POST("/register", h.Register)
		authRoutes.POST("/login", h.Login)
		authRoutes.POST("/social", h.SocialAuth)
		authRoutes.GET("/verify-email", h.VerifyEmail)
	}
}
