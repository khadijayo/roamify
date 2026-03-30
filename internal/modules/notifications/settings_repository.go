package notifications

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SettingsRepository handles the user_notification_settings table.

type SettingsRepository interface {
	UpsertSettings(s *UserNotificationSetting) error
	FindByUser(userID uuid.UUID) (*UserNotificationSetting, error)
}

type settingsRepository struct {
	db *gorm.DB
}

func NewSettingsRepository(db *gorm.DB) SettingsRepository {
	return &settingsRepository{db: db}
}

func (r *settingsRepository) UpsertSettings(s *UserNotificationSetting) error {
	return r.db.Save(s).Error
}

func (r *settingsRepository) FindByUser(userID uuid.UUID) (*UserNotificationSetting, error) {
	var s UserNotificationSetting
	err := r.db.Where("user_id = ?", userID).First(&s).Error
	return &s, err
}