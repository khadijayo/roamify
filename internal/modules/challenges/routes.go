package challenges

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, h *Handler, auth gin.HandlerFunc) {
	ch := r.Group("/challenges", auth)
	{
		ch.GET("/", h.ListChallenges)
		ch.GET("/leaderboard", h.GetLeaderboard)
		ch.POST("/", h.CreateChallenge)
		ch.POST("/accept", h.AcceptChallenge)
		ch.POST("/complete", h.CompleteChallenge)
		ch.GET("/my-progress", h.GetMyProgress)
		ch.GET("/trivia", h.ListTrivia)
		ch.POST("/trivia", h.CreateTrivia)
		ch.POST("/trivia/answer", h.AnswerTrivia)
	}
}
