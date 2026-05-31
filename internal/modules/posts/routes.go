package posts

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, h *Handler, auth gin.HandlerFunc) {
	p := r.Group("/posts", auth)
	{
		p.GET("", h.GetFeedV2)
		p.POST("/", h.CreatePost)
		p.GET("/", h.GetFeedV2)
		p.GET("/feed", h.GetFeedV2)
		p.GET("/:postId/comments", h.GetComments)
		p.POST("/:postId/comments", h.AddComment)
		p.DELETE("/:postId/comments", h.DeleteComment)
		p.GET("/:postId", h.GetPost)
		p.PATCH("/:postId", h.UpdatePost)
		p.DELETE("/:postId", h.DeletePost)
		p.POST("/:postId/like", h.LikePost)
		p.DELETE("/:postId/like", h.UnlikePost)
	}

	r.GET("/feed", auth, h.GetFeedV2)
	r.GET("/users/:userId/posts", auth, h.GetUserPosts)
}
