package notifications

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/khadijayo/roamify/pkg/middleware"
	"github.com/khadijayo/roamify/pkg/response"
)

// NotificationHandler serves the in-app notification inbox endpoints.

type NotificationHandler struct {
	svc NotificationService
}

func NewNotificationHandler(svc NotificationService) *NotificationHandler {
	return &NotificationHandler{svc: svc}
}

// GET /notifications?page=1&page_size=30
func (h *NotificationHandler) List(c *gin.Context) {
	userID := middleware.GetUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "30"))
	items, meta, err := h.svc.List(userID, page, pageSize)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OKPaginated(c, "notifications fetched", items, meta)
}

// GET /notifications/unread-count
// Returns {"count": N} — drives the bell badge dot in the app.
func (h *NotificationHandler) UnreadCount(c *gin.Context) {
	userID := middleware.GetUserID(c)
	count, err := h.svc.UnreadCount(userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, "unread count", gin.H{"count": count})
}

// PATCH /notifications/:notifId/read
func (h *NotificationHandler) MarkRead(c *gin.Context) {
	userID := middleware.GetUserID(c)
	notifID, err := uuid.Parse(c.Param("notifId"))
	if err != nil {
		response.BadRequest(c, "invalid notification id")
		return
	}
	if err := h.svc.MarkRead(notifID, userID); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, "notification marked as read", nil)
}

// PATCH /notifications/read-all
func (h *NotificationHandler) MarkAllRead(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if err := h.svc.MarkAllRead(userID); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, "all notifications marked as read", nil)
}