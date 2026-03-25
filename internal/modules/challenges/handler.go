package challenges

import (
	"strconv"

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

func (h *Handler) ListChallenges(c *gin.Context) {
	list, err := h.svc.ListChallenges()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, "challenges fetched", list)
}

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

func (h *Handler) GetMyProgress(c *gin.Context) {
	userID := middleware.GetUserID(c)
	progress, err := h.svc.GetMyProgress(userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, "progress fetched", progress)
}

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

func (h *Handler) GetLeaderboard(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	rows, err := h.svc.GetLeaderboard(limit)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, "leaderboard fetched", rows)
}

func (h *Handler) ListTrivia(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	rows, err := h.svc.ListTrivia(limit)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, "trivia fetched", rows)
}

func (h *Handler) CreateTrivia(c *gin.Context) {
	var req CreateTriviaQuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	q, err := h.svc.CreateTriviaQuestion(&req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Created(c, "trivia question created", q)
}

func (h *Handler) AnswerTrivia(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req AnswerTriviaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	attempt, err := h.svc.AnswerTrivia(userID, &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, "trivia answered", attempt)
}
