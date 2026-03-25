package passport

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, h *Handler, auth gin.HandlerFunc) {
	p := r.Group("/passport", auth)
	{

		p.PUT("/vault", h.UpsertVault)
		p.GET("/vault", h.GetVault)
		p.DELETE("/vault", h.DeleteVault)

		p.POST("/stamps", h.AddStamp)
		p.GET("/stamps", h.GetStamps)
		p.DELETE("/stamps/:stampId", h.DeleteStamp)
	}
}
