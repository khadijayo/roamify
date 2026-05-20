package users

import (
	"golang.org/x/crypto/bcrypt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type UserStatus string
type UserRole string

const (
	StatusActive    UserStatus = "active"
	StatusSuspended UserStatus = "suspended"
	StatusDeleted   UserStatus = "deleted"

	RoleUser  UserRole = "user"
	RoleAdmin UserRole = "admin"
)

type User struct {
	ID                 uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Email              *string        `gorm:"type:varchar(255);uniqueIndex:idx_users_email" json:"email"`
	FullName           string         `gorm:"type:varchar(255);not null"                     json:"full_name"`
	AvatarURL          *string        `gorm:"type:text"                                      json:"avatar_url"`
	PasswordHash       *string        `gorm:"type:text"                                      json:"-"`
	AuthProvider       *string        `gorm:"type:varchar(50)"                               json:"auth_provider,omitempty"`
	ProviderID         *string        `gorm:"type:varchar(255)"                              json:"provider_id,omitempty"`
	Role               UserRole       `gorm:"type:varchar(20);default:'user';not null"       json:"role"`
	Status             UserStatus     `gorm:"type:varchar(20);default:'active'"              json:"status"`
	IsBanned           bool           `gorm:"default:false;not null"                         json:"is_banned"`
	IsVerified         bool           `gorm:"default:false;not null"                         json:"is_verified"`
	VerificationToken  *string        `gorm:"type:text;index"                                json:"-"`
	TokenExpiresAt     *time.Time     `                                                      json:"-"`
	VerificationSentAt *time.Time     `                                                      json:"-"`
	LastLoginAt        *time.Time     `                                                      json:"last_login_at"`
	CreatedAt          time.Time      `                                                      json:"created_at"`
	UpdatedAt          time.Time      `                                                      json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index"                                          json:"-"`

	VibeProfile *VibeProfile `gorm:"foreignKey:UserID"             json:"vibe_profile,omitempty"`
}

type UserFollow struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	FollowerID  uuid.UUID `gorm:"type:uuid;not null;index"                       json:"follower_id"`
	FollowingID uuid.UUID `gorm:"type:uuid;not null;index"                       json:"following_id"`
	CreatedAt   time.Time `                                                      json:"created_at"`
}

func (UserFollow) TableName() string { return "user_follows" }

type UserPrivacySetting struct {
	ID                 uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID             uuid.UUID `gorm:"type:uuid;uniqueIndex;not null"                 json:"user_id"`
	GhostModeEnabled   bool      `gorm:"default:false"                                  json:"ghost_mode_enabled"`
	DataSharingEnabled bool      `gorm:"default:true"                                   json:"data_sharing_enabled"`
	MapVisibility      string    `gorm:"type:varchar(20);default:'public'"              json:"map_visibility"`
	UpdatedAt          time.Time `                                                      json:"updated_at"`
}

func (UserPrivacySetting) TableName() string { return "user_privacy_settings" }

type TravelPace string
type BudgetStyle string
type TravelWith string
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
	ID                 uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID             uuid.UUID      `gorm:"type:uuid;uniqueIndex;not null"                 json:"user_id"`
	ExplorerType       string         `gorm:"type:varchar(100)"                              json:"explorer_type"`
	PreferredVibes     pq.StringArray `gorm:"type:text[]"                                 json:"preferred_vibes"`
	TravelPace         TravelPace     `gorm:"type:varchar(20)"                               json:"travel_pace"`
	BudgetStyle        BudgetStyle    `gorm:"type:varchar(20)"                               json:"budget_style"`
	TravelWith         TravelWith     `gorm:"type:varchar(20)"                               json:"travel_with"`
	Interests          pq.StringArray `gorm:"type:text[]"                                 json:"interests"`
	OnboardingComplete bool           `gorm:"default:false"                                  json:"onboarding_complete"`
	ExplorerLevel      int            `gorm:"default:1"                                      json:"explorer_level"`
	RoamifyPoints      int            `gorm:"default:0"                                      json:"roamify_points"`
	CountriesVisited   int            `gorm:"default:0"                                      json:"countries_visited"`
	CreatedAt          time.Time      `                                                      json:"created_at"`
	UpdatedAt          time.Time      `                                                      json:"updated_at"`
}

type PasswordResetCode struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID     uuid.UUID  `gorm:"type:uuid;not null;index"                      json:"user_id"`
	Email      string     `gorm:"type:varchar(255);not null;index"              json:"email"`
	HashedCode string     `gorm:"type:text;not null"                            json:"-"`
	ExpiresAt  time.Time  `gorm:"not null"                                      json:"expires_at"`
	Attempts   int        `gorm:"default:0"                                     json:"attempts"`
	UsedAt     *time.Time `gorm:"index"                                          json:"used_at"`
	CreatedAt  time.Time  `gorm:"autoCreateTime"                                json:"created_at"`
	UpdatedAt  time.Time  `gorm:"autoUpdateTime"                                json:"updated_at"`
}

func (PasswordResetCode) TableName() string {
	return "password_reset_codes"
}

type RegisterRequest struct {
	FullName string `json:"full_name" binding:"required"`
	Email    string `json:"email"     binding:"required,email"`
	Password string `json:"password"  binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type SocialAuthRequest struct {
	Provider       string  `json:"provider" binding:"required,oneof=google tiktok apple"`
	ProviderUserID string  `json:"provider_user_id" binding:"required"`
	Email          *string `json:"email"`
	FullName       string  `json:"full_name" binding:"required"`
	AvatarURL      *string `json:"avatar_url"`
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

type FollowUserRequest struct {
	UserID uuid.UUID `json:"user_id" binding:"required"`
}

type UpdatePrivacySettingsRequest struct {
	GhostModeEnabled   *bool   `json:"ghost_mode_enabled"`
	DataSharingEnabled *bool   `json:"data_sharing_enabled"`
	MapVisibility      *string `json:"map_visibility"`
}

type SearchUsersRequest struct {
	Q string `form:"q"`
}

func CheckPassword(hashPtr *string, password string) bool {
	if hashPtr == nil {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(*hashPtr), []byte(password)) == nil
}

func GenerateJWT(user *User, secret string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID.String(),
		"email":   user.Email,
		"role":    user.Role,
		"exp":     time.Now().Add(72 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
