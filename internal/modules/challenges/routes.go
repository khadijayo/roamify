package challenges

import (
	"github.com/gin-gonic/gin"
	"github.com/khadijayo/roamify/pkg/middleware"
)

func RegisterRoutes(rg *gin.RouterGroup, h *Handler) {
	ch := rg.Group("/challenges", middleware.Auth())
	{
		ch.GET("", h.ListChallenges)
		ch.POST("", h.CreateChallenge) // admin/seeding
		ch.POST("/accept", h.AcceptChallenge)
		ch.POST("/complete", h.CompleteChallenge)
		ch.GET("/my-progress", h.GetMyProgress)
	}
}