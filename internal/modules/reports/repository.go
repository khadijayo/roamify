package reports

import (
	"context"

	"github.com/google/uuid"
	"github.com/khadijayo/roamify/internal/modules/posts"
	"github.com/khadijayo/roamify/internal/modules/users"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, report *Report) error
	CreateAndCountPost(ctx context.Context, report *Report) error
	TargetUserExists(ctx context.Context, id uuid.UUID) (bool, error)
	TargetPostExists(ctx context.Context, id uuid.UUID) (bool, error)
	List(ctx context.Context, status string, page, limit int) ([]Report, int64, error)
	FindByID(ctx context.Context, id uuid.UUID) (*Report, error)
	Update(ctx context.Context, report *Report) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, report *Report) error {
	return r.db.WithContext(ctx).Create(report).Error
}

func (r *repository) CreateAndCountPost(ctx context.Context, report *Report) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(report).Error; err != nil {
			return err
		}

		return tx.Model(&posts.Post{}).
			Where("id = ?", report.TargetID).
			UpdateColumn("reports_count", gorm.Expr("reports_count + 1")).
			Error
	})
}

func (r *repository) TargetUserExists(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&users.User{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *repository) TargetPostExists(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&posts.Post{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *repository) List(ctx context.Context, status string, page, limit int) ([]Report, int64, error) {
	var items []Report
	var total int64

	query := r.db.WithContext(ctx).Model(&Report{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *repository) FindByID(ctx context.Context, id uuid.UUID) (*Report, error) {
	var report Report
	if err := r.db.WithContext(ctx).First(&report, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &report, nil
}

func (r *repository) Update(ctx context.Context, report *Report) error {
	return r.db.WithContext(ctx).Save(report).Error
}
