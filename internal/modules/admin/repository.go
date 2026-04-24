package admin

import (
	"context"

	"github.com/google/uuid"
	"github.com/khadijayo/roamify/internal/modules/posts"
	"github.com/khadijayo/roamify/internal/modules/reports"
	"github.com/khadijayo/roamify/internal/modules/users"
	"gorm.io/gorm"
)

type Repository interface {
	ListUsers(ctx context.Context, page, limit int, query string) ([]users.User, int64, error)
	FindUserByID(ctx context.Context, id uuid.UUID) (*users.User, error)
	UpdateUser(ctx context.Context, user *users.User) error
	SoftDeleteUser(ctx context.Context, user *users.User) error

	ListPosts(ctx context.Context, page, limit int, query string) ([]posts.Post, int64, error)
	FindPostByID(ctx context.Context, id uuid.UUID) (*posts.Post, error)
	UpdatePost(ctx context.Context, post *posts.Post) error
	SoftDeletePost(ctx context.Context, post *posts.Post) error

	ListReports(ctx context.Context, page, limit int, status string) ([]reports.Report, int64, error)
	FindReportByID(ctx context.Context, id uuid.UUID) (*reports.Report, error)
	UpdateReport(ctx context.Context, report *reports.Report) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) ListUsers(ctx context.Context, page, limit int, query string) ([]users.User, int64, error) {
	var items []users.User
	var total int64

	db := r.db.WithContext(ctx).Model(&users.User{})
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
	if err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error; err != nil {
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

func (r *repository) ListPosts(ctx context.Context, page, limit int, query string) ([]posts.Post, int64, error) {
	var items []posts.Post
	var total int64

	db := r.db.WithContext(ctx).Model(&posts.Post{}).Preload("Tags")
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
	if err := r.db.WithContext(ctx).Preload("Tags").First(&post, "id = ?", id).Error; err != nil {
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
