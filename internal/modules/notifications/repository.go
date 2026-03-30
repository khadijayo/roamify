package notifications

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// NotificationRepository handles the notifications inbox table.

type NotificationRepository interface {
	Create(n *Notification) error
	FindByUser(userID uuid.UUID, limit, offset int) ([]Notification, int64, error)
	FindUnreadCount(userID uuid.UUID) (int64, error)
	MarkRead(id, userID uuid.UUID) error
	MarkAllRead(userID uuid.UUID) error
}

type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(n *Notification) error {
	return r.db.Create(n).Error
}

func (r *notificationRepository) FindByUser(userID uuid.UUID, limit, offset int) ([]Notification, int64, error) {
	var items []Notification
	var count int64
	r.db.Model(&Notification{}).Where("user_id = ?", userID).Count(&count)
	err := r.db.
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&items).Error
	return items, count, err
}

func (r *notificationRepository) FindUnreadCount(userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&Notification{}).
		Where("user_id = ? AND is_read = false", userID).
		Count(&count).Error
	return count, err
}

func (r *notificationRepository) MarkRead(id, userID uuid.UUID) error {
	return r.db.Model(&Notification{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("is_read", true).Error
}

func (r *notificationRepository) MarkAllRead(userID uuid.UUID) error {
	return r.db.Model(&Notification{}).
		Where("user_id = ? AND is_read = false", userID).
		Update("is_read", true).Error
}