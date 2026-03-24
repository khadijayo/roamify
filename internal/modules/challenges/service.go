package challenges

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/khadijayo/roamify/internal/modules/users"
	"gorm.io/gorm"
)

type Service interface {
	ListChallenges() ([]Challenge, error)
	AcceptChallenge(userID uuid.UUID, req *AcceptChallengeRequest) (*UserChallengeProgress, error)
	CompleteChallenge(userID uuid.UUID, req *CompleteChallengeRequest) (*UserChallengeProgress, error)
	GetMyProgress(userID uuid.UUID) ([]UserChallengeProgress, error)
	CreateChallenge(req *CreateChallengeRequest) (*Challenge, error) // admin use
}

type service struct {
	repo     Repository
	userRepo users.Repository
}

func NewService(repo Repository, userRepo users.Repository) Service {
	return &service{repo: repo, userRepo: userRepo}
}

func (s *service) ListChallenges() ([]Challenge, error) {
	return s.repo.FindAllActive()
}

func (s *service) AcceptChallenge(userID uuid.UUID, req *AcceptChallengeRequest) (*UserChallengeProgress, error) {
	challenge, err := s.repo.FindByID(req.ChallengeID)
	if err != nil {
		return nil, errors.New("challenge not found")
	}
	if !challenge.IsActive {
		return nil, errors.New("challenge is no longer active")
	}
	existing, err := s.repo.FindProgress(userID, req.ChallengeID)
	if err == nil && existing != nil {
		return nil, errors.New("challenge already accepted")
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	p := &UserChallengeProgress{
		UserID:      userID,
		ChallengeID: req.ChallengeID,
		Status:      StatusAccepted,
		AcceptedAt:  time.Now(),
	}
	if err := s.repo.AcceptChallenge(p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *service) CompleteChallenge(userID uuid.UUID, req *CompleteChallengeRequest) (*UserChallengeProgress, error) {
	progress, err := s.repo.FindProgress(userID, req.ChallengeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("challenge not accepted yet")
		}
		return nil, err
	}
	if progress.Status == StatusCompleted {
		return nil, errors.New("challenge already completed")
	}

	challenge, err := s.repo.FindByID(req.ChallengeID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	progress.Status = StatusCompleted
	progress.CompletedAt = &now
	progress.AwardedPoints = challenge.Points

	if err := s.repo.UpdateProgress(progress); err != nil {
		return nil, err
	}

	// Award points to vibe profile
	vp, err := s.userRepo.GetVibeProfile(userID)
	if err == nil {
		vp.RoamifyPoints += challenge.Points
		// Award XP / level up (every 500 points = next level)
		newLevel := (vp.RoamifyPoints / 500) + 1
		if newLevel > vp.ExplorerLevel {
			vp.ExplorerLevel = newLevel
		}
		_ = s.userRepo.UpsertVibeProfile(vp)
	}

	return progress, nil
}

func (s *service) GetMyProgress(userID uuid.UUID) ([]UserChallengeProgress, error) {
	return s.repo.FindUserProgress(userID)
}

func (s *service) CreateChallenge(req *CreateChallengeRequest) (*Challenge, error) {
	c := &Challenge{
		Title:       req.Title,
		Description: req.Description,
		Points:      req.Points,
		Category:    req.Category,
		Difficulty:  req.Difficulty,
		IsActive:    true,
	}
	return c, s.repo.CreateChallenge(c)
}
