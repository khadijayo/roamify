package notifications

import (
	"math"

	"github.com/google/uuid"
	"github.com/khadijayo/roamify/pkg/response"
)

// NotificationService handles the in-app notification inbox.
// Call Fire() from any other service to create a notification.

type NotificationService interface {
	Fire(input *CreateNotificationInput) error
	List(userID uuid.UUID, page, pageSize int) ([]Notification, *response.Meta, error)
	UnreadCount(userID uuid.UUID) (int64, error)
	MarkRead(notifID, userID uuid.UUID) error
	MarkAllRead(userID uuid.UUID) error
}

type notificationService struct {
	repo NotificationRepository
}

func NewNotificationService(repo NotificationRepository) NotificationService {
	return &notificationService{repo: repo}
}

// Fire creates a notification. Called internally — not from HTTP.
// Example: notifSvc.Fire(&CreateNotificationInput{UserID: id, Type: NotifTripInvite, Title: "You were invited!"})
func (s *notificationService) Fire(input *CreateNotificationInput) error {
	n := &Notification{
		UserID:  input.UserID,
		Type:    input.Type,
		Title:   input.Title,
		Body:    input.Body,
		RefID:   input.RefID,
		RefType: input.RefType,
		IsRead:  false,
	}
	return s.repo.Create(n)
}

func (s *notificationService) List(userID uuid.UUID, page, pageSize int) ([]Notification, *response.Meta, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 30
	}
	offset := (page - 1) * pageSize
	items, total, err := s.repo.FindByUser(userID, pageSize, offset)
	if err != nil {
		return nil, nil, err
	}
	meta := &response.Meta{
		Page:       page,
		PageSize:   pageSize,
		TotalItems: total,
		TotalPages: int(math.Ceil(float64(total) / float64(pageSize))),
	}
	return items, meta, nil
}

func (s *notificationService) UnreadCount(userID uuid.UUID) (int64, error) {
	return s.repo.FindUnreadCount(userID)
}

func (s *notificationService) MarkRead(notifID, userID uuid.UUID) error {
	return s.repo.MarkRead(notifID, userID)
}

func (s *notificationService) MarkAllRead(userID uuid.UUID) error {
	return s.repo.MarkAllRead(userID)
}