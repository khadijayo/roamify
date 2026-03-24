package notifications

import (
	"github.com/gin-gonic/gin"
	"github.com/khadijayo/roamify/pkg/middleware"
)

func RegisterRoutes(rg *gin.RouterGroup, h *Handler) {
	n := rg.Group("/notifications", middleware.Auth())
	{
		n.GET("/settings", h.GetSettings)
		n.PATCH("/settings", h.UpdateSettings)
	}
}