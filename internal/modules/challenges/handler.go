package challenges

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

// GET /challenges
func (h *Handler) ListChallenges(c *gin.Context) {
	list, err := h.svc.ListChallenges()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, "challenges fetched", list)
}

// POST /challenges/accept
func (h *Handler) AcceptChallenge(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req AcceptChallengeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	p, err := h.svc.AcceptChallenge(userID, &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, "challenge accepted", p)
}

// POST /challenges/complete
func (h *Handler) CompleteChallenge(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req CompleteChallengeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	p, err := h.svc.CompleteChallenge(userID, &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, "challenge completed! points awarded", p)
}

// GET /challenges/my-progress
func (h *Handler) GetMyProgress(c *gin.Context) {
	userID := middleware.GetUserID(c)
	progress, err := h.svc.GetMyProgress(userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, "progress fetched", progress)
}

// POST /challenges  (internal/admin route)
func (h *Handler) CreateChallenge(c *gin.Context) {
	var req CreateChallengeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	ch, err := h.svc.CreateChallenge(&req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Created(c, "challenge created", ch)
}