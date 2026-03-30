package challenges

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
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

	Challenge *Challenge `gorm:"foreignKey:ChallengeID" json:"challenge,omitempty"`
}

type TriviaQuestion struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Question      string         `gorm:"type:text;not null"                             json:"question"`
	Choices       pq.StringArray `gorm:"type:text[]"                               json:"choices"`
	CorrectAnswer string         `gorm:"type:varchar(255);not null"                     json:"-"`
	Points        int            `gorm:"default:50"                                     json:"points"`
	IsActive      bool           `gorm:"default:true"                                   json:"is_active"`
	CreatedAt     time.Time      `                                                      json:"created_at"`
	UpdatedAt     time.Time      `                                                      json:"updated_at"`
}

type TriviaAttempt struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID           uuid.UUID `gorm:"type:uuid;not null;index"                       json:"user_id"`
	TriviaQuestionID uuid.UUID `gorm:"type:uuid;not null;index"                       json:"trivia_question_id"`
	SelectedAnswer   string    `gorm:"type:varchar(255);not null"                     json:"selected_answer"`
	IsCorrect        bool      `gorm:"default:false"                                  json:"is_correct"`
	AwardedPoints    int       `gorm:"default:0"                                      json:"awarded_points"`
	CreatedAt        time.Time `                                                      json:"created_at"`
}

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

type CreateTriviaQuestionRequest struct {
	Question      string   `json:"question" binding:"required"`
	Choices       []string `json:"choices" binding:"required,min=2"`
	CorrectAnswer string   `json:"correct_answer" binding:"required"`
	Points        int      `json:"points"`
}

type AnswerTriviaRequest struct {
	QuestionID uuid.UUID `json:"question_id" binding:"required"`
	Answer     string    `json:"answer" binding:"required"`
}

type LeaderboardEntry struct {
	UserID           uuid.UUID `json:"user_id"`
	FullName         string    `json:"full_name"`
	AvatarURL        *string   `json:"avatar_url"`
	ExplorerLevel    int       `json:"explorer_level"`
	RoamifyPoints    int       `json:"roamify_points"`
	CountriesVisited int       `json:"countries_visited"`
}
