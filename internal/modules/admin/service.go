package admin

import (
	"context"
	"errors"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/khadijayo/roamify/internal/modules/posts"
	"github.com/khadijayo/roamify/internal/modules/reports"
	"github.com/khadijayo/roamify/internal/modules/trips"
	"github.com/khadijayo/roamify/internal/modules/users"
	"github.com/khadijayo/roamify/pkg/response"
	"gorm.io/gorm"
)

var (
	ErrTargetNotFound     = errors.New("resource not found")
	ErrCannotModerateSelf = errors.New("you cannot perform this action on your own account")
)

type UserDetails struct {
	User          *users.User `json:"user"`
	TripsCount    int64       `json:"trips_count"`
	PostsCount    int64       `json:"posts_count"`
	CommentsCount int64       `json:"comments_count"`
}

type UserActivityItem struct {
	ID          uuid.UUID `json:"id"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type AdminStats struct {
	TotalUsers       int64 `json:"total_users"`
	ActiveUsers      int64 `json:"active_users"`
	NewUsersThisWeek int64 `json:"new_users_this_week"`
	BannedUsers      int64 `json:"banned_users"`
}

type Service interface {
	ListUsers(ctx context.Context, page, limit int, status, query string) ([]users.User, *response.Meta, error)
	GetUserDetails(ctx context.Context, targetID uuid.UUID) (*UserDetails, error)
	ChangeUserRole(ctx context.Context, targetID uuid.UUID, role users.UserRole) (*users.User, error)
	GetUserActivity(ctx context.Context, targetID uuid.UUID) ([]UserActivityItem, error)
	BanUser(ctx context.Context, targetID, actorID uuid.UUID) (*users.User, error)
	UnbanUser(ctx context.Context, targetID, actorID uuid.UUID) (*users.User, error)
	DeleteUser(ctx context.Context, targetID, actorID uuid.UUID) error

	ListPosts(ctx context.Context, page, limit int, status, query string) ([]posts.Post, *response.Meta, error)
	GetPostDetails(ctx context.Context, postID uuid.UUID) (*posts.Post, error)
	DeletePost(ctx context.Context, postID uuid.UUID) error
	HidePost(ctx context.Context, postID uuid.UUID) (*posts.Post, error)
	UnhidePost(ctx context.Context, postID uuid.UUID) (*posts.Post, error)

	ListComments(ctx context.Context, page, limit int) ([]posts.PostComment, *response.Meta, error)
	DeleteComment(ctx context.Context, commentID uuid.UUID) error

	ListTrips(ctx context.Context, page, limit int) ([]trips.Trip, *response.Meta, error)
	DeleteTrip(ctx context.Context, tripID uuid.UUID) error

	ListReports(ctx context.Context, page, limit int, status string) ([]reports.Report, *response.Meta, error)
	ResolveReport(ctx context.Context, reportID uuid.UUID) (*reports.Report, error)
	GetStats(ctx context.Context) (*AdminStats, error)

	// AdminLogin authenticates an admin user and returns a JWT if valid
	AdminLogin(ctx context.Context, email, password string) (*users.User, string, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) ListUsers(ctx context.Context, page, limit int, status, query string) ([]users.User, *response.Meta, error) {
	page, limit = normalizePage(page, limit)
	items, total, err := s.repo.ListUsers(ctx, page, limit, status, strings.TrimSpace(query))
	if err != nil {
		return nil, nil, err
	}
	return items, newMeta(page, limit, total), nil
}

func (s *service) GetUserDetails(ctx context.Context, targetID uuid.UUID) (*UserDetails, error) {
	user, err := s.repo.FindUserByID(ctx, targetID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTargetNotFound
		}
		return nil, err
	}

	postsCount, err := s.repo.CountUserPosts(ctx, targetID)
	if err != nil {
		return nil, err
	}

	commentsCount, err := s.repo.CountUserComments(ctx, targetID)
	if err != nil {
		return nil, err
	}

	tripsCount, err := s.repo.CountUserTrips(ctx, targetID)
	if err != nil {
		return nil, err
	}

	return &UserDetails{
		User:          user,
		PostsCount:    postsCount,
		CommentsCount: commentsCount,
		TripsCount:    tripsCount,
	}, nil
}

func (s *service) ChangeUserRole(ctx context.Context, targetID uuid.UUID, role users.UserRole) (*users.User, error) {
	user, err := s.repo.FindUserByID(ctx, targetID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTargetNotFound
		}
		return nil, err
	}

	user.Role = role
	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *service) GetUserActivity(ctx context.Context, targetID uuid.UUID) ([]UserActivityItem, error) {
	_, err := s.repo.FindUserByID(ctx, targetID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTargetNotFound
		}
		return nil, err
	}

	postsItems, err := s.repo.ListUserActivityPosts(ctx, targetID, 5)
	if err != nil {
		return nil, err
	}
	commentsItems, err := s.repo.ListUserActivityComments(ctx, targetID, 5)
	if err != nil {
		return nil, err
	}
	tripsItems, err := s.repo.ListUserActivityTrips(ctx, targetID, 5)
	if err != nil {
		return nil, err
	}

	var items []UserActivityItem
	for _, p := range postsItems {
		description := p.Content
		if len(description) > 120 {
			description = description[:120] + "..."
		}
		items = append(items, UserActivityItem{
			ID:          p.ID,
			Type:        "post",
			Description: description,
			CreatedAt:   p.CreatedAt,
		})
	}

	for _, cmt := range commentsItems {
		description := cmt.Content
		if len(description) > 120 {
			description = description[:120] + "..."
		}
		items = append(items, UserActivityItem{
			ID:          cmt.ID,
			Type:        "comment",
			Description: description,
			CreatedAt:   cmt.CreatedAt,
		})
	}

	for _, t := range tripsItems {
		description := t.Title
		if len(description) > 120 {
			description = description[:120] + "..."
		}
		items = append(items, UserActivityItem{
			ID:          t.ID,
			Type:        "trip",
			Description: description,
			CreatedAt:   t.CreatedAt,
		})
	}

	sort.SliceStable(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})

	if len(items) > 10 {
		items = items[:10]
	}

	return items, nil
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

func (s *service) ListPosts(ctx context.Context, page, limit int, status, query string) ([]posts.Post, *response.Meta, error) {
	page, limit = normalizePage(page, limit)
	items, total, err := s.repo.ListPosts(ctx, page, limit, status, strings.TrimSpace(query))
	if err != nil {
		return nil, nil, err
	}
	return items, newMeta(page, limit, total), nil
}

func (s *service) GetPostDetails(ctx context.Context, postID uuid.UUID) (*posts.Post, error) {
	post, err := s.repo.FindPostByID(ctx, postID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTargetNotFound
		}
		return nil, err
	}
	return post, nil
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

func (s *service) ListComments(ctx context.Context, page, limit int) ([]posts.PostComment, *response.Meta, error) {
	page, limit = normalizePage(page, limit)
	items, total, err := s.repo.ListComments(ctx, page, limit)
	if err != nil {
		return nil, nil, err
	}
	return items, newMeta(page, limit, total), nil
}

func (s *service) DeleteComment(ctx context.Context, commentID uuid.UUID) error {
	comment, err := s.repo.FindCommentByID(ctx, commentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTargetNotFound
		}
		return err
	}
	return s.repo.DeleteComment(ctx, comment)
}

func (s *service) ListTrips(ctx context.Context, page, limit int) ([]trips.Trip, *response.Meta, error) {
	page, limit = normalizePage(page, limit)
	items, total, err := s.repo.ListTrips(ctx, page, limit)
	if err != nil {
		return nil, nil, err
	}
	return items, newMeta(page, limit, total), nil
}

func (s *service) DeleteTrip(ctx context.Context, tripID uuid.UUID) error {
	if err := s.repo.DeleteTrip(ctx, tripID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTargetNotFound
		}
		return err
	}
	return nil
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

func (s *service) GetStats(ctx context.Context) (*AdminStats, error) {
	return s.repo.GetStats(ctx)
}

func (s *service) AdminLogin(ctx context.Context, email, password string) (*users.User, string, error) {
	var user users.User
	err := s.repo.GetUserByEmail(ctx, email, &user)
	if err != nil {
		return nil, "", errors.New("invalid credentials")
	}
	if user.Role != users.RoleAdmin {
		return nil, "", errors.New("not an admin user")
	}
	if !users.CheckPassword(user.PasswordHash, password) {
		return nil, "", errors.New("invalid credentials")
	}
	// Generate JWT
	secret := "" // TODO: inject config
	token, err := users.GenerateJWT(&user, secret)
	if err != nil {
		return nil, "", errors.New("failed to generate token")
	}
	return &user, token, nil
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
