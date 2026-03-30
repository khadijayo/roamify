package notifications

import (
	"github.com/gin-gonic/gin"
	"github.com/khadijayo/roamify/pkg/middleware"
	"github.com/khadijayo/roamify/pkg/response"
)

// SettingsHandler serves GET/PATCH /notifications/settings

type SettingsHandler struct {
	svc SettingsService
}

func NewSettingsHandler(svc SettingsService) *SettingsHandler {
	return &SettingsHandler{svc: svc}
}

// GET /notifications/settings
func (h *SettingsHandler) GetSettings(c *gin.Context) {
	userID := middleware.GetUserID(c)
	settings, err := h.svc.GetSettings(userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, "notification settings fetched", settings)
}

// PATCH /notifications/settings
func (h *SettingsHandler) UpdateSettings(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req UpdateNotificationSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	settings, err := h.svc.UpdateSettings(userID, &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, "notification settings updated", settings)
}