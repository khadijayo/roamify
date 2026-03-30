package notifications

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SettingsService handles notification preference toggles.

type SettingsService interface {
	GetSettings(userID uuid.UUID) (*UserNotificationSetting, error)
	UpdateSettings(userID uuid.UUID, req *UpdateNotificationSettingsRequest) (*UserNotificationSetting, error)
}

type settingsService struct {
	repo SettingsRepository
}

func NewSettingsService(repo SettingsRepository) SettingsService {
	return &settingsService{repo: repo}
}

func (s *settingsService) GetSettings(userID uuid.UUID) (*UserNotificationSetting, error) {
	settings, err := s.repo.FindByUser(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Auto-create with sensible defaults on first access
			defaults := &UserNotificationSetting{
				UserID:                 userID,
				TripRemindersEnabled:   true,
				SquadUpdatesEnabled:    true,
				PriceDropAlertsEnabled: false,
			}
			if err := s.repo.UpsertSettings(defaults); err != nil {
				return nil, err
			}
			return defaults, nil
		}
		return nil, err
	}
	return settings, nil
}

func (s *settingsService) UpdateSettings(userID uuid.UUID, req *UpdateNotificationSettingsRequest) (*UserNotificationSetting, error) {
	settings, err := s.GetSettings(userID)
	if err != nil {
		return nil, err
	}
	if req.TripRemindersEnabled != nil {
		settings.TripRemindersEnabled = *req.TripRemindersEnabled
	}
	if req.SquadUpdatesEnabled != nil {
		settings.SquadUpdatesEnabled = *req.SquadUpdatesEnabled
	}
	if req.PriceDropAlertsEnabled != nil {
		settings.PriceDropAlertsEnabled = *req.PriceDropAlertsEnabled
	}
	if err := s.repo.UpsertSettings(settings); err != nil {
		return nil, err
	}
	return settings, nil
}