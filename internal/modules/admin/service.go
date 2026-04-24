package admin

import (
	"context"
	"errors"
	"math"
	"strings"

	"github.com/google/uuid"
	"github.com/khadijayo/roamify/internal/modules/posts"
	"github.com/khadijayo/roamify/internal/modules/reports"
	"github.com/khadijayo/roamify/internal/modules/users"
	"github.com/khadijayo/roamify/pkg/response"
	"gorm.io/gorm"
)

var (
	ErrTargetNotFound     = errors.New("resource not found")
	ErrCannotModerateSelf = errors.New("you cannot perform this action on your own account")
)

type Service interface {
	ListUsers(ctx context.Context, page, limit int, query string) ([]users.User, *response.Meta, error)
	BanUser(ctx context.Context, targetID, actorID uuid.UUID) (*users.User, error)
	UnbanUser(ctx context.Context, targetID, actorID uuid.UUID) (*users.User, error)
	DeleteUser(ctx context.Context, targetID, actorID uuid.UUID) error

	ListPosts(ctx context.Context, page, limit int, query string) ([]posts.Post, *response.Meta, error)
	DeletePost(ctx context.Context, postID uuid.UUID) error
	HidePost(ctx context.Context, postID uuid.UUID) (*posts.Post, error)
	UnhidePost(ctx context.Context, postID uuid.UUID) (*posts.Post, error)

	ListReports(ctx context.Context, page, limit int, status string) ([]reports.Report, *response.Meta, error)
	ResolveReport(ctx context.Context, reportID uuid.UUID) (*reports.Report, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) ListUsers(ctx context.Context, page, limit int, query string) ([]users.User, *response.Meta, error) {
	page, limit = normalizePage(page, limit)
	items, total, err := s.repo.ListUsers(ctx, page, limit, strings.TrimSpace(query))
	if err != nil {
		return nil, nil, err
	}
	return items, newMeta(page, limit, total), nil
}

func (s *service) BanUser(ctx context.Context, targetID, actorID uuid.UUID) (*users.User, error) {
	if targetID == actorID {
		return nil, ErrCannotModerateSelf
	}

	user, err := s.repo.FindUserByID(ctx, targetID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTargetNotFound
		}
		return nil, err
	}

	user.IsBanned = true
	user.Status = users.StatusSuspended
	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *service) UnbanUser(ctx context.Context, targetID, actorID uuid.UUID) (*users.User, error) {
	if targetID == actorID {
		return nil, ErrCannotModerateSelf
	}

	user, err := s.repo.FindUserByID(ctx, targetID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTargetNotFound
		}
		return nil, err
	}

	user.IsBanned = false
	user.Status = users.StatusActive
	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *service) DeleteUser(ctx context.Context, targetID, actorID uuid.UUID) error {
	if targetID == actorID {
		return ErrCannotModerateSelf
	}

	user, err := s.repo.FindUserByID(ctx, targetID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTargetNotFound
		}
		return err
	}

	user.Status = users.StatusDeleted
	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return err
	}
	return s.repo.SoftDeleteUser(ctx, user)
}

func (s *service) ListPosts(ctx context.Context, page, limit int, query string) ([]posts.Post, *response.Meta, error) {
	page, limit = normalizePage(page, limit)
	items, total, err := s.repo.ListPosts(ctx, page, limit, strings.TrimSpace(query))
	if err != nil {
		return nil, nil, err
	}
	return items, newMeta(page, limit, total), nil
}

func (s *service) DeletePost(ctx context.Context, postID uuid.UUID) error {
	post, err := s.repo.FindPostByID(ctx, postID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTargetNotFound
		}
		return err
	}
	return s.repo.SoftDeletePost(ctx, post)
}

func (s *service) HidePost(ctx context.Context, postID uuid.UUID) (*posts.Post, error) {
	post, err := s.repo.FindPostByID(ctx, postID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTargetNotFound
		}
		return nil, err
	}
	post.IsHidden = true
	if err := s.repo.UpdatePost(ctx, post); err != nil {
		return nil, err
	}
	return post, nil
}

func (s *service) UnhidePost(ctx context.Context, postID uuid.UUID) (*posts.Post, error) {
	post, err := s.repo.FindPostByID(ctx, postID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTargetNotFound
		}
		return nil, err
	}
	post.IsHidden = false
	if err := s.repo.UpdatePost(ctx, post); err != nil {
		return nil, err
	}
	return post, nil
}

func (s *service) ListReports(ctx context.Context, page, limit int, status string) ([]reports.Report, *response.Meta, error) {
	page, limit = normalizePage(page, limit)
	items, total, err := s.repo.ListReports(ctx, page, limit, strings.TrimSpace(status))
	if err != nil {
		return nil, nil, err
	}
	return items, newMeta(page, limit, total), nil
}

func (s *service) ResolveReport(ctx context.Context, reportID uuid.UUID) (*reports.Report, error) {
	item, err := s.repo.FindReportByID(ctx, reportID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTargetNotFound
		}
		return nil, err
	}

	item.Status = reports.StatusResolved
	if err := s.repo.UpdateReport(ctx, item); err != nil {
		return nil, err
	}

	return item, nil
}

func normalizePage(page, limit int) (int, int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return page, limit
}

func newMeta(page, limit int, total int64) *response.Meta {
	return &response.Meta{
		Page:       page,
		PageSize:   limit,
		TotalItems: total,
		TotalPages: int(math.Ceil(float64(total) / float64(limit))),
	}
}
