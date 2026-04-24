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

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	items, meta, err := h.svc.ListUsers(c.Request.Context(), page, limit, c.Query("q"))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.OKPaginated(c, "users fetched", items, meta)
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

	items, meta, err := h.svc.ListPosts(c.Request.Context(), page, limit, c.Query("q"))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.OKPaginated(c, "posts fetched", items, meta)
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
