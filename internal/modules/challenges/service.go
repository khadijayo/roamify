package challenges

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/khadijayo/roamify/internal/modules/users"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Service interface {
	ListChallenges() ([]Challenge, error)
	AcceptChallenge(userID uuid.UUID, req *AcceptChallengeRequest) (*UserChallengeProgress, error)
	CompleteChallenge(userID uuid.UUID, req *CompleteChallengeRequest) (*UserChallengeProgress, error)
	GetMyProgress(userID uuid.UUID) ([]UserChallengeProgress, error)
	CreateChallenge(req *CreateChallengeRequest) (*Challenge, error)
	GetLeaderboard(limit int) ([]LeaderboardEntry, error)
	ListTrivia(limit int) ([]TriviaQuestion, error)
	CreateTriviaQuestion(req *CreateTriviaQuestionRequest) (*TriviaQuestion, error)
	AnswerTrivia(userID uuid.UUID, req *AnswerTriviaRequest) (*TriviaAttempt, error)
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

	vp, err := s.userRepo.GetVibeProfile(userID)
	if err == nil {
		vp.RoamifyPoints += challenge.Points

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

func (s *service) GetLeaderboard(limit int) ([]LeaderboardEntry, error) {
	profiles, err := s.userRepo.ListTopVibeProfiles(limit)
	if err != nil {
		return nil, err
	}
	entries := make([]LeaderboardEntry, 0, len(profiles))
	for _, vp := range profiles {
		u, userErr := s.userRepo.FindByID(vp.UserID)
		if userErr != nil || u == nil {
			continue
		}
		entries = append(entries, LeaderboardEntry{
			UserID:           u.ID,
			FullName:         u.FullName,
			AvatarURL:        u.AvatarURL,
			ExplorerLevel:    vp.ExplorerLevel,
			RoamifyPoints:    vp.RoamifyPoints,
			CountriesVisited: vp.CountriesVisited,
		})
	}
	return entries, nil
}

func (s *service) ListTrivia(limit int) ([]TriviaQuestion, error) {
	return s.repo.FindActiveTrivia(limit)
}

func (s *service) CreateTriviaQuestion(req *CreateTriviaQuestionRequest) (*TriviaQuestion, error) {
	points := req.Points
	if points <= 0 {
		points = 50
	}
	q := &TriviaQuestion{
		Question:      req.Question,
		Choices:       pq.StringArray(req.Choices),
		CorrectAnswer: req.CorrectAnswer,
		Points:        points,
		IsActive:      true,
	}
	if err := s.repo.CreateTriviaQuestion(q); err != nil {
		return nil, err
	}
	return q, nil
}

func (s *service) AnswerTrivia(userID uuid.UUID, req *AnswerTriviaRequest) (*TriviaAttempt, error) {
	question, err := s.repo.FindTriviaByID(req.QuestionID)
	if err != nil {
		return nil, errors.New("trivia question not found")
	}
	correct := req.Answer == question.CorrectAnswer
	awarded := 0
	if correct {
		awarded = question.Points
	}
	attempt := &TriviaAttempt{
		UserID:           userID,
		TriviaQuestionID: req.QuestionID,
		SelectedAnswer:   req.Answer,
		IsCorrect:        correct,
		AwardedPoints:    awarded,
	}
	if err := s.repo.CreateTriviaAttempt(attempt); err != nil {
		return nil, err
	}
	if awarded > 0 {
		vp, vpErr := s.userRepo.GetVibeProfile(userID)
		if vpErr == nil {
			vp.RoamifyPoints += awarded
			newLevel := (vp.RoamifyPoints / 500) + 1
			if newLevel > vp.ExplorerLevel {
				vp.ExplorerLevel = newLevel
			}
			_ = s.userRepo.UpsertVibeProfile(vp)
		}
	}
	return attempt, nil
}
