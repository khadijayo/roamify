package discovery

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.RouterGroup, h *Handler, auth gin.HandlerFunc) {
	d := r.Group("/discovery", auth)
	{
		d.GET("/home", h.GetHomeDashboard)
		d.GET("/vibe-search", h.VibeSearch)
		d.GET("/atlas", h.GetAtlasLocations)
		d.GET("/atlas/:locationId", h.GetLocationDetail)
	}

	r.GET("/home", auth, h.GetHomeDashboard)
	r.GET("/vibe-search", auth, h.VibeSearch)
	r.GET("/atlas", auth, h.GetAtlasLocations)
	r.GET("/atlas/:locationId", auth, h.GetLocationDetail)

	s := r.Group("/search", auth)
	{
		s.GET("/global", h.GlobalSearch)
		s.GET("/results", h.GlobalSearch)
	}

	assistant := r.Group("/assistant", auth)
	{
		assistant.POST("/travel", h.TravelAssistant)
	}
}
