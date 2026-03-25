package posts

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, h *Handler, auth gin.HandlerFunc) {
	p := r.Group("/posts", auth)
	{
		p.POST("/", h.CreatePost)
		p.GET("/", h.GetFeed)
		p.GET("/:postId", h.GetPost)
		p.PATCH("/:postId", h.UpdatePost)
		p.DELETE("/:postId", h.DeletePost)
		p.POST("/:postId/like", h.LikePost)
		p.DELETE("/:postId/like", h.UnlikePost)
	}

	r.GET("/users/:userId/posts", auth, h.GetUserPosts)
}
