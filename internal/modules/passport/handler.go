package passport

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/khadijayo/roamify/pkg/middleware"
	"github.com/khadijayo/roamify/pkg/response"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// ---- Vault ----

// PUT /passport/vault
func (h *Handler) UpsertVault(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req UpsertVaultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	record, err := h.svc.UpsertVault(userID, &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, "passport vault saved", record)
}

// GET /passport/vault
func (h *Handler) GetVault(c *gin.Context) {
	userID := middleware.GetUserID(c)
	record, err := h.svc.GetVault(userID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.OK(c, "passport vault fetched", record)
}

// DELETE /passport/vault
func (h *Handler) DeleteVault(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if err := h.svc.DeleteVault(userID); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, "passport vault deleted", nil)
}

// ---- Stamps ----

// POST /passport/stamps
func (h *Handler) AddStamp(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req AddStampRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	stamp, err := h.svc.AddStamp(userID, &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, "stamp added", stamp)
}

// GET /passport/stamps
func (h *Handler) GetStamps(c *gin.Context) {
	userID := middleware.GetUserID(c)
	stamps, err := h.svc.GetStamps(userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, "stamps fetched", stamps)
}

// DELETE /passport/stamps/:stampId
func (h *Handler) DeleteStamp(c *gin.Context) {
	userID := middleware.GetUserID(c)
	stampID, err := uuid.Parse(c.Param("stampId"))
	if err != nil {
		response.BadRequest(c, "invalid stamp id")
		return
	}
	if err := h.svc.DeleteStamp(stampID, userID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, "stamp deleted", nil)
}