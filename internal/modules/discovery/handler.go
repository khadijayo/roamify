package discovery

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/khadijayo/roamify/pkg/response"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
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
