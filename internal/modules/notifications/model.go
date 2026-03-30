package notifications

import (
	"time"

	"github.com/google/uuid"
)

// -----------------------------------------------------------------------
// Notification settings (per-user toggles shown in Profile → Settings)
// -----------------------------------------------------------------------

type UserNotificationSetting struct {
	ID                     uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID                 uuid.UUID `gorm:"type:uuid;uniqueIndex;not null"                 json:"user_id"`
	TripRemindersEnabled   bool      `gorm:"default:true"                                   json:"trip_reminders_enabled"`
	SquadUpdatesEnabled    bool      `gorm:"default:true"                                   json:"squad_updates_enabled"`
	PriceDropAlertsEnabled bool      `gorm:"default:false"                                  json:"price_drop_alerts_enabled"`
	UpdatedAt              time.Time `                                                      json:"updated_at"`
}

func (UserNotificationSetting) TableName() string { return "user_notification_settings" }

type UpdateNotificationSettingsRequest struct {
	TripRemindersEnabled   *bool `json:"trip_reminders_enabled"`
	SquadUpdatesEnabled    *bool `json:"squad_updates_enabled"`
	PriceDropAlertsEnabled *bool `json:"price_drop_alerts_enabled"`
}

// -----------------------------------------------------------------------
// In-app notification inbox (bell icon, unread count, mark read)
// -----------------------------------------------------------------------

type NotificationType string

const (
	NotifTripInvite         NotificationType = "trip_invite"
	NotifTripStatusChanged  NotificationType = "trip_status_changed"
	NotifMemberJoined       NotificationType = "member_joined"
	NotifPostLiked          NotificationType = "post_liked"
	NotifChallengeCompleted NotificationType = "challenge_completed"
	NotifNewFollower        NotificationType = "new_follower"
	NotifChatMessage        NotificationType = "chat_message"
	NotifPriceDrop          NotificationType = "price_drop"
)

type Notification struct {
	ID        uuid.UUID        `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID        `gorm:"type:uuid;not null;index"                       json:"user_id"`
	Type      NotificationType `gorm:"type:varchar(50);not null"                      json:"type"`
	Title     string           `gorm:"type:varchar(255);not null"                     json:"title"`
	Body      string           `gorm:"type:text"                                      json:"body"`
	RefID     *uuid.UUID       `gorm:"type:uuid"                                      json:"ref_id,omitempty"`
	RefType   *string          `gorm:"type:varchar(50)"                               json:"ref_type,omitempty"`
	IsRead    bool             `gorm:"default:false"                                  json:"is_read"`
	CreatedAt time.Time        `                                                      json:"created_at"`
}

func (Notification) TableName() string { return "notifications" }

// CreateNotificationInput is called internally by other services (not an HTTP DTO).
type CreateNotificationInput struct {
	UserID  uuid.UUID
	Type    NotificationType
	Title   string
	Body    string
	RefID   *uuid.UUID
	RefType *string
}