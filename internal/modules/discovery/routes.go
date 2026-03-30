package discovery

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.RouterGroup, h *Handler, auth gin.HandlerFunc) {
	d := r.Group("/discovery", auth)
	{
		d.GET("/home", h.GetHomeDashboard)
		d.GET("/vibe-search", h.VibeSearch)
		d.GET("/atlas", h.GetAtlasLocations)
		d.GET("/atlas/geojson", h.GetAtlasGeoJSON) // <-- NEW (before /:locationId)
		d.GET("/atlas/:locationId", h.GetLocationDetail)
		d.GET("/price-drops", h.GetPersonalizedPriceDrops) // <-- NEW
		d.GET("/recommended", h.GetRecommended)            // <-- NEW
		d.POST("/locations/generate", h.GenerateLocationsFromAnswers)
	}

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
