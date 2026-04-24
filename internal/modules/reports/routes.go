package reports

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.RouterGroup, h *Handler, auth gin.HandlerFunc) {
	group := r.Group("/reports", auth)
	{
		group.POST("/", h.Create)
	}
}
