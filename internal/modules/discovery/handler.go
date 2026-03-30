package discovery

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/khadijayo/roamify/pkg/middleware"
	"github.com/khadijayo/roamify/pkg/response"
)

type Handler struct {
	svc             Service
	personalizedSvc *PersonalizedService
}

func NewHandler(svc Service, personalizedSvc *PersonalizedService) *Handler {
	return &Handler{svc: svc, personalizedSvc: personalizedSvc}
}

func (h *Handler) GetHomeDashboard(c *gin.Context) {
	data, err := h.svc.GetHomeDashboard()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, "home dashboard fetched", data)
}

func (h *Handler) VibeSearch(c *gin.Context) {
	query := c.Query("q")
	data, err := h.svc.VibeSearch(query)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, "vibe search results fetched", data)
}

func (h *Handler) GetAtlasLocations(c *gin.Context) {
	var filter AtlasFilterRequest
	_ = c.ShouldBindQuery(&filter)
	data, err := h.svc.GetAtlasLocations(&filter)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, "atlas locations fetched", data)
}

func (h *Handler) GetLocationDetail(c *gin.Context) {
	id, err := uuid.Parse(c.Param("locationId"))
	if err != nil {
		response.BadRequest(c, "invalid location id")
		return
	}
	loc, err := h.svc.GetLocation(id)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	if loc == nil {
		response.NotFound(c, "location not found")
		return
	}
	response.OK(c, "location detail fetched", loc)
}

func (h *Handler) GlobalSearch(c *gin.Context) {
	query := c.Query("q")
	data, err := h.svc.GlobalSearch(query)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, "global search results fetched", data)
}

func (h *Handler) TravelAssistant(c *gin.Context) {
	var req AssistantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	data, err := h.svc.TravelAssistant(&req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, "assistant response generated", data)
}

// GET /discovery/price-drops
// Returns price drop alerts personalised to the authenticated user's vibe profile.
// Falls back to all deals if the user has no vibe profile yet.
func (h *Handler) GetPersonalizedPriceDrops(c *gin.Context) {
	userID := middleware.GetUserID(c)
	deals, err := h.personalizedSvc.GetPersonalizedPriceDrops(userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, "price drops fetched", deals)
}

// GET /discovery/recommended
// Returns destinations ranked by how closely they match the user's interests/vibes.
func (h *Handler) GetRecommended(c *gin.Context) {
	userID := middleware.GetUserID(c)
	locs, err := h.personalizedSvc.RecommendedDestinations(userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, "recommended destinations fetched", locs)
}

// GET /discovery/atlas/geojson
// Returns the atlas locations as a GeoJSON FeatureCollection.
// Pass this directly to Mapbox as a source:
//
//	map.addSource('atlas', { type: 'geojson', data: response.data })
func (h *Handler) GetAtlasGeoJSON(c *gin.Context) {
	fc := GetAtlasGeoJSON()
	c.JSON(200, fc)
}

func (h *Handler) GenerateLocationsFromAnswers(c *gin.Context) {
	var req GenerateLocationsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	data, err := h.svc.GenerateLocationsFromAnswers(&req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, "generated discovery locations", data)
}
