package challenges

import (
	"time"

	"github.com/google/uuid"
)

type ChallengeCategory string
type DifficultyLevel string
type ProgressStatus string

const (
	CategoryExplorer     ChallengeCategory = "explorer"
	CategorySocial       ChallengeCategory = "social"
	CategoryCollection   ChallengeCategory = "collection"
	CategoryGamification ChallengeCategory = "gamification"

	DifficultyEasy   DifficultyLevel = "easy"
	DifficultyMedium DifficultyLevel = "medium"
	DifficultyHard   DifficultyLevel = "hard"

	StatusAccepted  ProgressStatus = "accepted"
	StatusCompleted ProgressStatus = "completed"
)

type Challenge struct {
	ID          uuid.UUID         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Title       string            `gorm:"type:varchar(255);not null"                     json:"title"`
	Description string            `gorm:"type:text"                                      json:"description"`
	Category    ChallengeCategory `gorm:"type:varchar(50);default:'explorer'"            json:"category"`
	Difficulty  DifficultyLevel   `gorm:"type:varchar(20);default:'easy'"                json:"difficulty"`
	Points      int               `gorm:"default:100"                                    json:"points"`
	IsActive    bool              `gorm:"default:true"                                   json:"is_active"`
	CreatedAt   time.Time         `                                                      json:"created_at"`
	UpdatedAt   time.Time         `                                                      json:"updated_at"`
}

type UserChallengeProgress struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID        uuid.UUID      `gorm:"type:uuid;not null;index"                       json:"user_id"`
	ChallengeID   uuid.UUID      `gorm:"type:uuid;not null;index"                       json:"challenge_id"`
	Status        ProgressStatus `gorm:"type:varchar(20);default:'accepted'"            json:"status"`
	AwardedPoints int            `gorm:"default:0"                                      json:"awarded_points"`
	AcceptedAt    time.Time      `                                                      json:"accepted_at"`
	CompletedAt   *time.Time     `                                                      json:"completed_at"`
	CreatedAt     time.Time      `                                                      json:"created_at"`
	UpdatedAt     time.Time      `                                                      json:"updated_at"`

	// Associations
	Challenge *Challenge `gorm:"foreignKey:ChallengeID" json:"challenge,omitempty"`
}

// ---------------------------------------------------------------------------
// DTOs
// ---------------------------------------------------------------------------

type CreateChallengeRequest struct {
	Title       string            `json:"title"       binding:"required"`
	Description string            `json:"description"`
	Category    ChallengeCategory `json:"category"`
	Difficulty  DifficultyLevel   `json:"difficulty"`
	Points      int               `json:"points"`
}

type AcceptChallengeRequest struct {
	ChallengeID uuid.UUID `json:"challenge_id" binding:"required"`
}

type CompleteChallengeRequest struct {
	ChallengeID uuid.UUID `json:"challenge_id" binding:"required"`
}
