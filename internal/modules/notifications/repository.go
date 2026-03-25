package notifications

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	UpsertSettings(s *UserNotificationSetting) error
	FindByUser(userID uuid.UUID) (*UserNotificationSetting, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) UpsertSettings(s *UserNotificationSetting) error {
	return r.db.Save(s).Error
}

func (r *repository) FindByUser(userID uuid.UUID) (*UserNotificationSetting, error) {
	var s UserNotificationSetting
	err := r.db.Where("user_id = ?", userID).First(&s).Error
	return &s, err
}
