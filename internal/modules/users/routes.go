package users

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, h *Handler, auth gin.HandlerFunc) {
	authRoutes := r.Group("/auth")
	{
		authRoutes.POST("/register", h.Register)
		authRoutes.POST("/login", h.Login)
		authRoutes.POST("/social", h.SocialAuth)
	}
	userRoutes := r.Group("/users")
    userRoutes.Use(auth)
  {
       userRoutes.GET("/me", h.GetMe)
       userRoutes.PATCH("/me", h.UpdateMe)
       userRoutes.GET("/me/vibe", h.GetVibeProfile)
       userRoutes.PUT("/me/vibe", h.UpsertVibeProfile)
       userRoutes.GET("/me/privacy", h.GetPrivacySettings)
       userRoutes.PATCH("/me/privacy", h.UpdatePrivacySettings)
       userRoutes.POST("/follow", h.FollowUser)
       userRoutes.DELETE("/follow/:userId", h.UnfollowUser)
       userRoutes.GET("/search", h.SearchUsers)           // <-- NEW (before /:userId)
       userRoutes.GET("/:userId", h.GetPublicProfile)
       userRoutes.GET("/:userId/followers", h.GetFollowers)
       userRoutes.GET("/:userId/following", h.GetFollowing)
  }

}