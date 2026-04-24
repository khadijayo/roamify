package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/khadijayo/roamify/pkg/middleware"
)

func RegisterRoutes(r *gin.RouterGroup, h *Handler, auth gin.HandlerFunc) {
	group := r.Group("/admin")
	group.Use(auth, middleware.AdminOnly())
	{
		group.GET("/users", h.ListUsers)
		group.PUT("/users/:id/ban", h.BanUser)
		group.PUT("/users/:id/unban", h.UnbanUser)
		group.DELETE("/users/:id", h.DeleteUser)

		group.GET("/posts", h.ListPosts)
		group.DELETE("/posts/:id", h.DeletePost)
		group.PUT("/posts/:id/hide", h.HidePost)
		group.PUT("/posts/:id/unhide", h.UnhidePost)

		group.GET("/reports", h.ListReports)
		group.PUT("/reports/:id/resolve", h.ResolveReport)
	}
}
