package notifications

import (
	"time"

	"github.com/google/uuid"
)

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
