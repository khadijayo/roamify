package admin

import (
	"context"
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/khadijayo/roamify/internal/modules/posts"
	"github.com/khadijayo/roamify/internal/modules/users"
	"github.com/khadijayo/roamify/pkg/middleware"
	"github.com/khadijayo/roamify/pkg/response"
)

type Handler struct {
	svc Service
}

type UpdateUserRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=user admin"`
}

// AdminLoginRequest is the payload for admin login
// POST /admin/login
// { "email": "...", "password": "..." }
type AdminLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := c.Query("search")
	status := c.Query("status")

	items, meta, err := h.svc.ListUsers(c.Request.Context(), page, limit, status, search)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.OKPaginated(c, "users fetched", items, meta)
}

func (h *Handler) GetUser(c *gin.Context) {
	targetID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	item, err := h.svc.GetUserDetails(c.Request.Context(), targetID)
	if err != nil {
		switch {
		case errors.Is(err, ErrTargetNotFound):
			response.NotFound(c, err.Error())
		default:
			response.InternalError(c, err.Error())
		}
		return
	}

	response.OK(c, "user details fetched", item)
}

func (h *Handler) UpdateUserRole(c *gin.Context) {
	targetID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	var req UpdateUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	item, err := h.svc.ChangeUserRole(c.Request.Context(), targetID, users.UserRole(req.Role))
	if err != nil {
		switch {
		case errors.Is(err, ErrTargetNotFound):
			response.NotFound(c, err.Error())
		default:
			response.InternalError(c, err.Error())
		}
		return
	}

	response.OK(c, "user role updated", item)
}

func (h *Handler) GetUserActivity(c *gin.Context) {
	targetID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	items, err := h.svc.GetUserActivity(c.Request.Context(), targetID)
	if err != nil {
		switch {
		case errors.Is(err, ErrTargetNotFound):
			response.NotFound(c, err.Error())
		default:
			response.InternalError(c, err.Error())
		}
		return
	}

	response.OK(c, "user activity fetched", items)
}

func (h *Handler) BanUser(c *gin.Context) {
	h.moderateUser(c, h.svc.BanUser)
}

func (h *Handler) UnbanUser(c *gin.Context) {
	h.moderateUser(c, h.svc.UnbanUser)
}

func (h *Handler) DeleteUser(c *gin.Context) {
	targetID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	err = h.svc.DeleteUser(c.Request.Context(), targetID, middleware.GetUserID(c))
	if err != nil {
		switch {
		case errors.Is(err, ErrTargetNotFound):
			response.NotFound(c, err.Error())
		case errors.Is(err, ErrCannotModerateSelf):
			response.BadRequest(c, err.Error())
		default:
			response.InternalError(c, err.Error())
		}
		return
	}

	response.OK(c, "user deleted", gin.H{"id": targetID})
}

func (h *Handler) ListPosts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	status := c.Query("status")
	search := c.Query("search")

	items, meta, err := h.svc.ListPosts(c.Request.Context(), page, limit, status, search)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.OKPaginated(c, "posts fetched", items, meta)
}

func (h *Handler) GetPost(c *gin.Context) {
	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid post id")
		return
	}

	item, err := h.svc.GetPostDetails(c.Request.Context(), postID)
	if err != nil {
		if errors.Is(err, ErrTargetNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, "post fetched", item)
}

func (h *Handler) DeletePost(c *gin.Context) {
	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid post id")
		return
	}

	if err := h.svc.DeletePost(c.Request.Context(), postID); err != nil {
		if errors.Is(err, ErrTargetNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, "post deleted", gin.H{"id": postID})
}

func (h *Handler) HidePost(c *gin.Context) {
	h.togglePostVisibility(c, h.svc.HidePost)
}

func (h *Handler) UnhidePost(c *gin.Context) {
	h.togglePostVisibility(c, h.svc.UnhidePost)
}

func (h *Handler) ListComments(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	items, meta, err := h.svc.ListComments(c.Request.Context(), page, limit)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.OKPaginated(c, "comments fetched", items, meta)
}

func (h *Handler) DeleteComment(c *gin.Context) {
	commentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid comment id")
		return
	}

	if err := h.svc.DeleteComment(c.Request.Context(), commentID); err != nil {
		if errors.Is(err, ErrTargetNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, "comment deleted", gin.H{"id": commentID})
}

func (h *Handler) ListTrips(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	items, meta, err := h.svc.ListTrips(c.Request.Context(), page, limit)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.OKPaginated(c, "trips fetched", items, meta)
}

func (h *Handler) DeleteTrip(c *gin.Context) {
	tripID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid trip id")
		return
	}

	if err := h.svc.DeleteTrip(c.Request.Context(), tripID); err != nil {
		if errors.Is(err, ErrTargetNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, "trip deleted", gin.H{"id": tripID})
}

func (h *Handler) ListReports(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	items, meta, err := h.svc.ListReports(c.Request.Context(), page, limit, c.Query("status"))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.OKPaginated(c, "reports fetched", items, meta)
}

func (h *Handler) ResolveReport(c *gin.Context) {
	reportID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid report id")
		return
	}

	item, err := h.svc.ResolveReport(c.Request.Context(), reportID)
	if err != nil {
		if errors.Is(err, ErrTargetNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, "report resolved", item)
}

func (h *Handler) Stats(c *gin.Context) {
	stats, err := h.svc.GetStats(c.Request.Context())
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, "admin stats fetched", stats)
}

func (h *Handler) moderateUser(c *gin.Context, fn func(ctx context.Context, targetID, actorID uuid.UUID) (*users.User, error)) {
	targetID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	item, err := fn(c.Request.Context(), targetID, middleware.GetUserID(c))
	if err != nil {
		switch {
		case errors.Is(err, ErrTargetNotFound):
			response.NotFound(c, err.Error())
		case errors.Is(err, ErrCannotModerateSelf):
			response.BadRequest(c, err.Error())
		default:
			response.InternalError(c, err.Error())
		}
		return
	}

	response.OK(c, "user updated", item)
}

func (h *Handler) togglePostVisibility(c *gin.Context, fn func(ctx context.Context, postID uuid.UUID) (*posts.Post, error)) {
	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid post id")
		return
	}

	item, err := fn(c.Request.Context(), postID)
	if err != nil {
		if errors.Is(err, ErrTargetNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, "post updated", item)
}

// GET /admin/login
func (h *Handler) AdminLoginPage(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Admin login page (implement HTML if needed)"})
}

// POST /admin/login
func (h *Handler) AdminLogin(c *gin.Context) {
	var req AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	user, token, err := h.svc.AdminLogin(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		response.Forbidden(c, err.Error())
		return
	}
	c.JSON(200, gin.H{
		"token": token,
		"user": gin.H{
			"id":    user.ID,
			"email": user.Email,
			"role":  user.Role,
		},
	})
}
