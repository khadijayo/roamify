package users

import (
	"github.com/gin-gonic/gin"
	"github.com/khadijayo/roamify/pkg/middleware"
	"github.com/khadijayo/roamify/pkg/response"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// POST /auth/register
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	res, err := h.svc.Register(&req)
	if err != nil {
		response.Conflict(c, err.Error())
		return
	}
	response.Created(c, "account created successfully", res)
}

// POST /auth/login
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	res, err := h.svc.Login(&req)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}
	response.OK(c, "login successful", res)
}

// GET /users/me
func (h *Handler) GetMe(c *gin.Context) {
	userID := middleware.GetUserID(c)
	user, err := h.svc.GetProfile(userID)
	if err != nil {
		response.NotFound(c, "user not found")
		return
	}
	response.OK(c, "profile fetched", user)
}

// PATCH /users/me
func (h *Handler) UpdateMe(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	user, err := h.svc.UpdateProfile(userID, &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, "profile updated", user)
}

// GET /users/me/vibe
func (h *Handler) GetVibeProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	vp, err := h.svc.GetVibeProfile(userID)
	if err != nil {
		response.NotFound(c, "vibe profile not found")
		return
	}
	response.OK(c, "vibe profile fetched", vp)
}

// PUT /users/me/vibe
func (h *Handler) UpsertVibeProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req UpdateVibeProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	vp, err := h.svc.UpsertVibeProfile(userID, &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, "vibe profile saved", vp)
}
