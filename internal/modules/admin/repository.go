package admin

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/khadijayo/roamify/internal/modules/posts"
	"github.com/khadijayo/roamify/internal/modules/reports"
	"github.com/khadijayo/roamify/internal/modules/trips"
	"github.com/khadijayo/roamify/internal/modules/users"
	"gorm.io/gorm"
)

type Repository interface {
	ListUsers(ctx context.Context, page, limit int, status, query string) ([]users.User, int64, error)
	FindUserByID(ctx context.Context, id uuid.UUID) (*users.User, error)
	UpdateUser(ctx context.Context, user *users.User) error
	SoftDeleteUser(ctx context.Context, user *users.User) error
	CountUserPosts(ctx context.Context, userID uuid.UUID) (int64, error)
	CountUserComments(ctx context.Context, userID uuid.UUID) (int64, error)
	CountUserTrips(ctx context.Context, userID uuid.UUID) (int64, error)
	ListUserActivityPosts(ctx context.Context, userID uuid.UUID, limit int) ([]posts.Post, error)
	ListUserActivityComments(ctx context.Context, userID uuid.UUID, limit int) ([]posts.PostComment, error)
	ListUserActivityTrips(ctx context.Context, userID uuid.UUID, limit int) ([]trips.Trip, error)

	ListPosts(ctx context.Context, page, limit int, status, query string) ([]posts.Post, int64, error)
	FindPostByID(ctx context.Context, id uuid.UUID) (*posts.Post, error)
	UpdatePost(ctx context.Context, post *posts.Post) error
	SoftDeletePost(ctx context.Context, post *posts.Post) error

	ListComments(ctx context.Context, page, limit int) ([]posts.PostComment, int64, error)
	FindCommentByID(ctx context.Context, id uuid.UUID) (*posts.PostComment, error)
	DeleteComment(ctx context.Context, comment *posts.PostComment) error

	ListTrips(ctx context.Context, page, limit int) ([]trips.Trip, int64, error)
	DeleteTrip(ctx context.Context, tripID uuid.UUID) error

	ListReports(ctx context.Context, page, limit int, status string) ([]reports.Report, int64, error)
	FindReportByID(ctx context.Context, id uuid.UUID) (*reports.Report, error)
	UpdateReport(ctx context.Context, report *reports.Report) error
	GetStats(ctx context.Context) (*AdminStats, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) ListUsers(ctx context.Context, page, limit int, status, query string) ([]users.User, int64, error) {
	var items []users.User
	var total int64

	db := r.db.WithContext(ctx).Model(&users.User{})
	if status != "" {
		switch strings.ToLower(status) {
		case "active":
			db = db.Where("status = ?", users.StatusActive)
		case "banned":
			db = db.Where("is_banned = ?", true)
		case "suspended":
			db = db.Where("status = ?", users.StatusSuspended)
		}
	}
	if query != "" {
		like := "%" + query + "%"
		db = db.Where("email ILIKE ? OR full_name ILIKE ?", like, like)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := db.Order("created_at DESC").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *repository) FindUserByID(ctx context.Context, id uuid.UUID) (*users.User, error) {
	var user users.User
	if err := r.db.WithContext(ctx).Preload("VibeProfile").First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *repository) UpdateUser(ctx context.Context, user *users.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *repository) SoftDeleteUser(ctx context.Context, user *users.User) error {
	return r.db.WithContext(ctx).Delete(user).Error
}

func (r *repository) CountUserPosts(ctx context.Context, userID uuid.UUID) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&posts.Post{}).
		Where("author_user_id = ?", userID).
		Count(&total).Error
	return total, err
}

func (r *repository) CountUserComments(ctx context.Context, userID uuid.UUID) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&posts.PostComment{}).
		Where("user_id = ?", userID).
		Count(&total).Error
	return total, err
}

func (r *repository) CountUserTrips(ctx context.Context, userID uuid.UUID) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&trips.Trip{}).
		Select("DISTINCT trips.id").
		Joins("LEFT JOIN trip_members ON trip_members.trip_id = trips.id").
		Where("trips.owner_user_id = ? OR trip_members.user_id = ?", userID, userID).
		Count(&total).Error
	return total, err
}

func (r *repository) ListUserActivityPosts(ctx context.Context, userID uuid.UUID, limit int) ([]posts.Post, error) {
	var items []posts.Post
	err := r.db.WithContext(ctx).
		Where("author_user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&items).Error
	return items, err
}

func (r *repository) ListUserActivityComments(ctx context.Context, userID uuid.UUID, limit int) ([]posts.PostComment, error) {
	var items []posts.PostComment
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&items).Error
	return items, err
}

func (r *repository) ListUserActivityTrips(ctx context.Context, userID uuid.UUID, limit int) ([]trips.Trip, error) {
	var items []trips.Trip
	err := r.db.WithContext(ctx).
		Joins("LEFT JOIN trip_members ON trip_members.trip_id = trips.id").
		Where("trips.owner_user_id = ? OR trip_members.user_id = ?", userID, userID).
		Distinct("trips.id").
		Order("trips.created_at DESC").
		Limit(limit).
		Find(&items).Error
	return items, err
}

func (r *repository) ListPosts(ctx context.Context, page, limit int, status, query string) ([]posts.Post, int64, error) {
	var items []posts.Post
	var total int64

	db := r.db.WithContext(ctx).Model(&posts.Post{}).Preload("Tags")
	if status != "" {
		switch strings.ToLower(status) {
		case "hidden":
			db = db.Where("is_hidden = ?", true)
		case "visible":
			db = db.Where("is_hidden = ?", false)
		case "reported":
			db = db.Where("reports_count > 0")
		}
	}
	if query != "" {
		like := "%" + query + "%"
		db = db.Where("content ILIKE ? OR location ILIKE ?", like, like)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := db.Order("created_at DESC").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *repository) FindPostByID(ctx context.Context, id uuid.UUID) (*posts.Post, error) {
	var post posts.Post
	if err := r.db.WithContext(ctx).Preload("Tags").Preload("Comments").First(&post, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *repository) UpdatePost(ctx context.Context, post *posts.Post) error {
	return r.db.WithContext(ctx).Save(post).Error
}

func (r *repository) SoftDeletePost(ctx context.Context, post *posts.Post) error {
	return r.db.WithContext(ctx).Delete(post).Error
}

func (r *repository) ListComments(ctx context.Context, page, limit int) ([]posts.PostComment, int64, error) {
	var items []posts.PostComment
	var total int64

	db := r.db.WithContext(ctx).Model(&posts.PostComment{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := db.Order("created_at DESC").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *repository) FindCommentByID(ctx context.Context, id uuid.UUID) (*posts.PostComment, error) {
	var comment posts.PostComment
	if err := r.db.WithContext(ctx).First(&comment, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &comment, nil
}

func (r *repository) DeleteComment(ctx context.Context, comment *posts.PostComment) error {
	return r.db.WithContext(ctx).Delete(comment).Error
}

func (r *repository) ListTrips(ctx context.Context, page, limit int) ([]trips.Trip, int64, error) {
	var items []trips.Trip
	var total int64

	db := r.db.WithContext(ctx).Model(&trips.Trip{}).Preload("Members")
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := db.Order("created_at DESC").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *repository) DeleteTrip(ctx context.Context, tripID uuid.UUID) error {
	res := r.db.WithContext(ctx).Delete(&trips.Trip{}, "id = ?", tripID)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *repository) ListReports(ctx context.Context, page, limit int, status string) ([]reports.Report, int64, error) {
	var items []reports.Report
	var total int64

	db := r.db.WithContext(ctx).Model(&reports.Report{})
	if status != "" {
		db = db.Where("status = ?", status)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := db.Order("created_at DESC").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *repository) FindReportByID(ctx context.Context, id uuid.UUID) (*reports.Report, error) {
	var report reports.Report
	if err := r.db.WithContext(ctx).First(&report, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &report, nil
}

func (r *repository) UpdateReport(ctx context.Context, report *reports.Report) error {
	return r.db.WithContext(ctx).Save(report).Error
}

func (r *repository) GetStats(ctx context.Context) (*AdminStats, error) {
	var total, active, banned, newUsers int64
	if err := r.db.WithContext(ctx).Model(&users.User{}).
		Where("status != ?", users.StatusDeleted).
		Count(&total).Error; err != nil {
		return nil, err
	}

	if err := r.db.WithContext(ctx).Model(&users.User{}).
		Where("status = ? AND is_banned = ?", users.StatusActive, false).
		Count(&active).Error; err != nil {
		return nil, err
	}

	if err := r.db.WithContext(ctx).Model(&users.User{}).
		Where("is_banned = ?", true).
		Count(&banned).Error; err != nil {
		return nil, err
	}

	weekAgo := time.Now().UTC().AddDate(0, 0, -7)
	if err := r.db.WithContext(ctx).Model(&users.User{}).
		Where("created_at >= ?", weekAgo).
		Count(&newUsers).Error; err != nil {
		return nil, err
	}

	return &AdminStats{
		TotalUsers:       total,
		ActiveUsers:      active,
		NewUsersThisWeek: newUsers,
		BannedUsers:      banned,
	}, nil
}
