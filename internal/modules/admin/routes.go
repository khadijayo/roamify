package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/khadijayo/roamify/pkg/middleware"
)

func RegisterRoutes(r *gin.RouterGroup, h *Handler, auth gin.HandlerFunc) {
	r.GET("/admin/login", h.AdminLoginPage)
	r.POST("/admin/login", h.AdminLogin)

	group := r.Group("/admin")
	group.Use(auth, middleware.RequireAdmin())
	{
		group.GET("/users", h.ListUsers)
		group.GET("/users/:id", h.GetUser)
		group.PATCH("/users/:id/role", h.UpdateUserRole)
		group.GET("/users/:id/activity", h.GetUserActivity)
		group.PUT("/users/:id/ban", h.BanUser)
		group.PUT("/users/:id/unban", h.UnbanUser)
		group.DELETE("/users/:id", h.DeleteUser)

		group.GET("/posts", h.ListPosts)
		group.GET("/posts/:id", h.GetPost)
		group.DELETE("/posts/:id", h.DeletePost)
		group.PUT("/posts/:id/hide", h.HidePost)
		group.PUT("/posts/:id/unhide", h.UnhidePost)

		group.GET("/comments", h.ListComments)
		group.DELETE("/comments/:id", h.DeleteComment)

		group.GET("/trips", h.ListTrips)
		group.DELETE("/trips/:id", h.DeleteTrip)

		group.GET("/reports", h.ListReports)
		group.PUT("/reports/:id/resolve", h.ResolveReport)

		group.GET("/stats", h.Stats)
	}
}
