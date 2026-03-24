package challenges

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, h *Handler, auth gin.HandlerFunc) {
	ch := r.Group("/challenges", auth)
	{
		ch.GET("/", h.ListChallenges)
		ch.POST("/", h.CreateChallenge) // admin/seeding
		ch.POST("/accept", h.AcceptChallenge)
		ch.POST("/complete", h.CompleteChallenge)
		ch.GET("/my-progress", h.GetMyProgress)
	}
}