package notifications

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, h *Handler, auth gin.HandlerFunc) {
	n := r.Group("/notifications", auth)
	{
		n.GET("/settings", h.GetSettings)
		n.PATCH("/settings", h.UpdateSettings)
	}
}