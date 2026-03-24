package posts

import (
	"github.com/gin-gonic/gin"
	"github.com/khadijayo/roamify/pkg/middleware"
)

func RegisterRoutes(rg *gin.RouterGroup, h *Handler) {
	auth := middleware.Auth()

	p := rg.Group("/posts", auth)
	{
		p.POST("", h.CreatePost)
		p.GET("", h.GetFeed)
		p.GET("/:postId", h.GetPost)
		p.PATCH("/:postId", h.UpdatePost)
		p.DELETE("/:postId", h.DeletePost)
		p.POST("/:postId/like", h.LikePost)
		p.DELETE("/:postId/like", h.UnlikePost)
	}

	// User post grid — nested under /users
	rg.GET("/users/:userId/posts", auth, h.GetUserPosts)
}