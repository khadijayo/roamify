package challenges

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	CreateChallenge(c *Challenge) error
	FindAllActive() ([]Challenge, error)
	FindByID(id uuid.UUID) (*Challenge, error)
	UpdateChallenge(c *Challenge) error

	AcceptChallenge(p *UserChallengeProgress) error
	FindProgress(userID, challengeID uuid.UUID) (*UserChallengeProgress, error)
	FindUserProgress(userID uuid.UUID) ([]UserChallengeProgress, error)
	UpdateProgress(p *UserChallengeProgress) error

	CreateTriviaQuestion(q *TriviaQuestion) error
	FindActiveTrivia(limit int) ([]TriviaQuestion, error)
	FindTriviaByID(id uuid.UUID) (*TriviaQuestion, error)
	CreateTriviaAttempt(a *TriviaAttempt) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateChallenge(c *Challenge) error {
	return r.db.Create(c).Error
}

func (r *repository) FindAllActive() ([]Challenge, error) {
	var challenges []Challenge
	err := r.db.Where("is_active = true").Order("difficulty ASC, points ASC").Find(&challenges).Error
	return challenges, err
}

func (r *repository) FindByID(id uuid.UUID) (*Challenge, error) {
	var c Challenge
	err := r.db.First(&c, "id = ?", id).Error
	return &c, err
}

func (r *repository) UpdateChallenge(c *Challenge) error {
	return r.db.Save(c).Error
}

func (r *repository) AcceptChallenge(p *UserChallengeProgress) error {
	return r.db.Create(p).Error
}

func (r *repository) FindProgress(userID, challengeID uuid.UUID) (*UserChallengeProgress, error) {
	var p UserChallengeProgress
	err := r.db.Where("user_id = ? AND challenge_id = ?", userID, challengeID).
		Preload("Challenge").First(&p).Error
	return &p, err
}

func (r *repository) FindUserProgress(userID uuid.UUID) ([]UserChallengeProgress, error) {
	var progress []UserChallengeProgress
	err := r.db.Where("user_id = ?", userID).
		Preload("Challenge").
		Order("accepted_at DESC").
		Find(&progress).Error
	return progress, err
}

func (r *repository) UpdateProgress(p *UserChallengeProgress) error {
	return r.db.Save(p).Error
}

func (r *repository) CreateTriviaQuestion(q *TriviaQuestion) error {
	return r.db.Create(q).Error
}

func (r *repository) FindActiveTrivia(limit int) ([]TriviaQuestion, error) {
	if limit < 1 || limit > 100 {
		limit = 10
	}
	var rows []TriviaQuestion
	err := r.db.Where("is_active = true").Order("created_at DESC").Limit(limit).Find(&rows).Error
	return rows, err
}

func (r *repository) FindTriviaByID(id uuid.UUID) (*TriviaQuestion, error) {
	var q TriviaQuestion
	err := r.db.First(&q, "id = ?", id).Error
	return &q, err
}

func (r *repository) CreateTriviaAttempt(a *TriviaAttempt) error {
	return r.db.Create(a).Error
}
