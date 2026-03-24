package users

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserStatus string

const (
	StatusActive    UserStatus = "active"
	StatusSuspended UserStatus = "suspended"
	StatusDeleted   UserStatus = "deleted"
)

type User struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Email       *string        `gorm:"type:varchar(255);uniqueIndex"                  json:"email"`
	FullName    string         `gorm:"type:varchar(255);not null"                     json:"full_name"`
	AvatarURL   *string        `gorm:"type:text"                                      json:"avatar_url"`
	PasswordHash *string       `gorm:"type:text"                                      json:"-"`
	Status      UserStatus     `gorm:"type:varchar(20);default:'active'"              json:"status"`
	LastLoginAt *time.Time     `                                                      json:"last_login_at"`
	CreatedAt   time.Time      `                                                      json:"created_at"`
	UpdatedAt   time.Time      `                                                      json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index"                                          json:"-"`

	// Associations
	VibeProfile         *VibeProfile         `gorm:"foreignKey:UserID"             json:"vibe_profile,omitempty"`
}

// ---------------------------------------------------------------------------

type TravelPace  string
type BudgetStyle string
type TravelWith  string
type ExplorerType string

const (
	PaceChill     TravelPace = "chill"
	PaceBalanced  TravelPace = "balanced"
	PaceAdventure TravelPace = "adventure"

	BudgetBackpacker BudgetStyle = "backpacker"
	BudgetMidRange   BudgetStyle = "mid-range"
	BudgetLuxury     BudgetStyle = "luxury"

	WithSolo    TravelWith = "solo"
	WithPartner TravelWith = "partner"
	WithSquad   TravelWith = "squad"
	WithFamily  TravelWith = "family"
)

type VibeProfile struct {
	ID                uuid.UUID    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID            uuid.UUID    `gorm:"type:uuid;uniqueIndex;not null"                 json:"user_id"`
	ExplorerType      string       `gorm:"type:varchar(100)"                              json:"explorer_type"`
	PreferredVibes    []string     `gorm:"type:text[];serializer:json"                    json:"preferred_vibes"`
	TravelPace        TravelPace   `gorm:"type:varchar(20)"                               json:"travel_pace"`
	BudgetStyle       BudgetStyle  `gorm:"type:varchar(20)"                               json:"budget_style"`
	TravelWith        TravelWith   `gorm:"type:varchar(20)"                               json:"travel_with"`
	Interests         []string     `gorm:"type:text[];serializer:json"                    json:"interests"`
	OnboardingComplete bool        `gorm:"default:false"                                  json:"onboarding_complete"`
	ExplorerLevel     int          `gorm:"default:1"                                      json:"explorer_level"`
	RoamifyPoints     int          `gorm:"default:0"                                      json:"roamify_points"`
	CountriesVisited  int          `gorm:"default:0"                                      json:"countries_visited"`
	CreatedAt         time.Time    `                                                      json:"created_at"`
	UpdatedAt         time.Time    `                                                      json:"updated_at"`
}

// ---------------------------------------------------------------------------
// DTOs

type RegisterRequest struct {
	FullName string `json:"full_name" binding:"required"`
	Email    string `json:"email"     binding:"required,email"`
	Password string `json:"password"  binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}

type UpdateProfileRequest struct {
	FullName  string  `json:"full_name"`
	AvatarURL *string `json:"avatar_url"`
}

type UpdateVibeProfileRequest struct {
	ExplorerType       string      `json:"explorer_type"`
	PreferredVibes     []string    `json:"preferred_vibes"`
	TravelPace         TravelPace  `json:"travel_pace"`
	BudgetStyle        BudgetStyle `json:"budget_style"`
	TravelWith         TravelWith  `json:"travel_with"`
	Interests          []string    `json:"interests"`
	OnboardingComplete *bool       `json:"onboarding_complete"`
}