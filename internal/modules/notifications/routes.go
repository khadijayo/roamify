package notifications

import "github.com/gin-gonic/gin"

// RegisterSettingsRoutes registers the notification preference toggles.
// Called from main.go wireModules().
func RegisterSettingsRoutes(r *gin.RouterGroup, h *SettingsHandler, auth gin.HandlerFunc) {
	n := r.Group("/notifications", auth)
	{
		n.GET("/settings", h.GetSettings)    // GET  /notifications/settings
		n.PATCH("/settings", h.UpdateSettings) // PATCH /notifications/settings
	}
}

// RegisterNotificationRoutes registers the in-app inbox endpoints.
// Called from main.go wireModules() separately from settings routes.
func RegisterNotificationRoutes(r *gin.RouterGroup, h *NotificationHandler, auth gin.HandlerFunc) {
	n := r.Group("/notifications", auth)
	{
		n.GET("", h.List)                        // GET   /notifications
		n.GET("/unread-count", h.UnreadCount)    // GET   /notifications/unread-count
		n.PATCH("/read-all", h.MarkAllRead)      // PATCH /notifications/read-all
		n.PATCH("/:notifId/read", h.MarkRead)    // PATCH /notifications/:notifId/read
	}
}