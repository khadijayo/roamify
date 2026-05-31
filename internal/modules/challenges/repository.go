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
	GetUserPointTotals(limit int) ([]LeaderboardEntry, error)
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

func (r *repository) GetUserPointTotals(limit int) ([]LeaderboardEntry, error) {
	if limit < 1 || limit > 100 {
		limit = 20
	}

	var rows []LeaderboardEntry
	err := r.db.Table("users AS u").
		Select(`
			u.id AS user_id,
			u.full_name,
			u.avatar_url,
			COALESCE(vp.explorer_level, 1) AS explorer_level,
			COALESCE(vp.countries_visited, 0) AS countries_visited,
			COALESCE(cp.challenge_points, 0) AS challenge_points,
			COALESCE(tp.trivia_points, 0) AS trivia_points,
			COALESCE(cp.challenge_points, 0) + COALESCE(tp.trivia_points, 0) AS total_points,
			COALESCE(cp.challenge_points, 0) + COALESCE(tp.trivia_points, 0) AS roamify_points
		`).
		Joins("LEFT JOIN vibe_profiles vp ON vp.user_id = u.id").
		Joins(`
			LEFT JOIN (
				SELECT user_id, SUM(awarded_points) AS challenge_points
				FROM user_challenge_progresses
				WHERE status = ? AND awarded_points > 0
				GROUP BY user_id
			) cp ON cp.user_id = u.id
		`, StatusCompleted).
		Joins(`
			LEFT JOIN (
				SELECT user_id, SUM(awarded_points) AS trivia_points
				FROM trivia_attempts
				WHERE is_correct = true AND awarded_points > 0
				GROUP BY user_id
			) tp ON tp.user_id = u.id
		`).
		Where("u.deleted_at IS NULL").
		Order("total_points DESC, challenge_points DESC, trivia_points DESC, u.full_name ASC").
		Limit(limit).
		Scan(&rows).Error
	return rows, err
}
