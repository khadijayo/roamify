package reports

import (
	"context"
	"errors"
	"math"
	"strings"

	"github.com/google/uuid"
	"github.com/khadijayo/roamify/pkg/response"
)

var (
	ErrInvalidTarget = errors.New("invalid report target")
)

type Service interface {
	Create(ctx context.Context, reporterID uuid.UUID, req *CreateReportRequest) (*Report, error)
	List(ctx context.Context, status string, page, limit int) ([]Report, *response.Meta, error)
	Resolve(ctx context.Context, reportID uuid.UUID) (*Report, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, reporterID uuid.UUID, req *CreateReportRequest) (*Report, error) {
	reason := strings.TrimSpace(req.Reason)
	if reason == "" {
		return nil, errors.New("reason is required")
	}

	report := &Report{
		ReporterID: reporterID,
		TargetType: req.TargetType,
		TargetID:   req.TargetID,
		Reason:     reason,
		Status:     StatusPending,
	}

	switch req.TargetType {
	case TargetTypeUser:
		exists, err := s.repo.TargetUserExists(ctx, req.TargetID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, ErrInvalidTarget
		}
		if err := s.repo.Create(ctx, report); err != nil {
			return nil, err
		}
	case TargetTypePost:
		exists, err := s.repo.TargetPostExists(ctx, req.TargetID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, ErrInvalidTarget
		}
		if err := s.repo.CreateAndCountPost(ctx, report); err != nil {
			return nil, err
		}
	default:
		return nil, ErrInvalidTarget
	}

	return report, nil
}

func (s *service) List(ctx context.Context, status string, page, limit int) ([]Report, *response.Meta, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	items, total, err := s.repo.List(ctx, strings.TrimSpace(status), page, limit)
	if err != nil {
		return nil, nil, err
	}

	meta := &response.Meta{
		Page:       page,
		PageSize:   limit,
		TotalItems: total,
		TotalPages: int(math.Ceil(float64(total) / float64(limit))),
	}
	return items, meta, nil
}

func (s *service) Resolve(ctx context.Context, reportID uuid.UUID) (*Report, error) {
	report, err := s.repo.FindByID(ctx, reportID)
	if err != nil {
		return nil, err
	}

	report.Status = StatusResolved
	if err := s.repo.Update(ctx, report); err != nil {
		return nil, err
	}

	return report, nil
}
