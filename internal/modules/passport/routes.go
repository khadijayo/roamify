package passport

import (
	"github.com/gin-gonic/gin"
	"github.com/khadijayo/roamify/pkg/middleware"
)

func RegisterRoutes(rg *gin.RouterGroup, h *Handler) {
	p := rg.Group("/passport", middleware.Auth())
	{
		// Vault
		p.PUT("/vault", h.UpsertVault)
		p.GET("/vault", h.GetVault)
		p.DELETE("/vault", h.DeleteVault)

		// Stamps
		p.POST("/stamps", h.AddStamp)
		p.GET("/stamps", h.GetStamps)
		p.DELETE("/stamps/:stampId", h.DeleteStamp)
	}
}