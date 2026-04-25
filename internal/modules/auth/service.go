package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/khadijayo/roamify/config"
	"github.com/khadijayo/roamify/internal/modules/users"
	emailsvc "github.com/khadijayo/roamify/internal/services/email"
	pkgjwt "github.com/khadijayo/roamify/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const verificationTTL = 15 * time.Minute

var (
	ErrInvalidCredentials       = errors.New("invalid email or password")
	ErrEmailAlreadyRegistered   = errors.New("email already registered")
	ErrEmailNotVerified         = errors.New("please verify your email before logging in")
	ErrAccountBanned            = errors.New("your account has been banned")
	ErrInvalidVerificationToken = errors.New("invalid or expired verification token")
	ErrVerificationEmailFailed  = errors.New("failed to send verification email")
)

type RegisterRequest struct {
	FullName string `json:"full_name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type SocialAuthRequest struct {
	Provider       string  `json:"provider" binding:"required,oneof=google tiktok apple"`
	ProviderUserID string  `json:"provider_user_id" binding:"required"`
	Email          *string `json:"email"`
	FullName       string  `json:"full_name" binding:"required"`
	AvatarURL      *string `json:"avatar_url"`
}

type RegisterResponse struct {
	User                  *users.User `json:"user"`
	VerificationSent      bool        `json:"verification_sent"`
	VerificationExpiresAt time.Time   `json:"verification_expires_at"`
}

type AuthResponse struct {
	Token string      `json:"token"`
	User  *users.User `json:"user"`
}

type VerifyEmailResponse struct {
	UserID     uuid.UUID `json:"user_id"`
	IsVerified bool      `json:"is_verified"`
}

type Service interface {
	Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error)
	Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error)
	SocialAuth(ctx context.Context, req *SocialAuthRequest) (*AuthResponse, error)
	VerifyEmail(ctx context.Context, token string) (*VerifyEmailResponse, error)
}

type service struct {
	repo           Repository
	emailService   emailsvc.Service
	jwtSecret      string
	jwtExpiryHours int
	appBaseURL     string
}

func NewService(repo Repository, emailService emailsvc.Service, cfg *config.Config) Service {
	return &service{
		repo:           repo,
		emailService:   emailService,
		jwtSecret:      cfg.JWTSecret,
		jwtExpiryHours: cfg.JWTExpiryHours,
		appBaseURL:     sanitizeAppBaseURL(cfg.AppBaseURL),
	}
}

func (s *service) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
	email := normalizeEmail(req.Email)
	fullName := strings.TrimSpace(req.FullName)
	if fullName == "" {
		return nil, errors.New("full_name is required")
	}

	existing, err := s.repo.FindByEmail(ctx, email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	rawToken, hashedToken, expiresAt, err := generateVerificationToken()
	if err != nil {
		return nil, err
	}

	if existing != nil {
		if existing.IsVerified {
			return nil, ErrEmailAlreadyRegistered
		}

		hashValue := string(hash)
		existing.FullName = fullName
		existing.PasswordHash = &hashValue
		existing.VerificationToken = &hashedToken
		existing.TokenExpiresAt = &expiresAt
		existing.IsVerified = false
		existing.IsBanned = false
		existing.Role = users.RoleUser

		if err := s.repo.UpdateUser(ctx, existing); err != nil {
			return nil, err
		}

		if err := s.sendVerification(ctx, email, fullName, rawToken); err != nil {
			return nil, err
		}

		return &RegisterResponse{
			User:                  existing,
			VerificationSent:      true,
			VerificationExpiresAt: expiresAt,
		}, nil
	}

	hashValue := string(hash)
	user := &users.User{
		FullName:          fullName,
		Email:             &email,
		PasswordHash:      &hashValue,
		Role:              users.RoleUser,
		Status:            users.StatusActive,
		IsVerified:        false,
		IsBanned:          false,
		VerificationToken: &hashedToken,
		TokenExpiresAt:    &expiresAt,
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	if err := s.sendVerification(ctx, email, fullName, rawToken); err != nil {
		return nil, err
	}

	return &RegisterResponse{
		User:                  user,
		VerificationSent:      true,
		VerificationExpiresAt: expiresAt,
	}, nil
}

func (s *service) Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error) {
	user, err := s.repo.FindByEmail(ctx, normalizeEmail(req.Email))
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	if user.IsBanned {
		return nil, ErrAccountBanned
	}
	if !user.IsVerified {
		return nil, ErrEmailNotVerified
	}
	if user.PasswordHash == nil {
		return nil, errors.New("this account uses social login")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := s.issueToken(user)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	user.LastLoginAt = &now
	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return nil, err
	}

	return &AuthResponse{Token: token, User: user}, nil
}

func (s *service) SocialAuth(ctx context.Context, req *SocialAuthRequest) (*AuthResponse, error) {
	user, err := s.repo.FindByProvider(ctx, req.Provider, req.ProviderUserID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if user == nil && req.Email != nil {
		user, err = s.repo.FindByEmail(ctx, normalizeEmail(*req.Email))
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}

	if user != nil {
		if user.IsBanned {
			return nil, ErrAccountBanned
		}
		provider := req.Provider
		user.AuthProvider = &provider
		user.ProviderID = &req.ProviderUserID
		user.FullName = strings.TrimSpace(req.FullName)
		user.AvatarURL = req.AvatarURL
		user.IsVerified = true
		user.VerificationToken = nil
		user.TokenExpiresAt = nil
		if err := s.repo.UpdateUser(ctx, user); err != nil {
			return nil, err
		}
	} else {
		provider := req.Provider
		user = &users.User{
			FullName:     strings.TrimSpace(req.FullName),
			Email:        normalizeOptionalEmail(req.Email),
			AvatarURL:    req.AvatarURL,
			AuthProvider: &provider,
			ProviderID:   &req.ProviderUserID,
			Role:         users.RoleUser,
			Status:       users.StatusActive,
			IsVerified:   true,
			IsBanned:     false,
		}
		if err := s.repo.CreateUser(ctx, user); err != nil {
			return nil, err
		}
	}

	token, err := s.issueToken(user)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	user.LastLoginAt = &now
	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return nil, err
	}

	return &AuthResponse{Token: token, User: user}, nil
}

func (s *service) VerifyEmail(ctx context.Context, token string) (*VerifyEmailResponse, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, ErrInvalidVerificationToken
	}

	user, err := s.repo.FindByVerificationToken(ctx, hashToken(token))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidVerificationToken
		}
		return nil, err
	}

	if user.TokenExpiresAt == nil || time.Now().UTC().After(user.TokenExpiresAt.UTC()) {
		user.VerificationToken = nil
		user.TokenExpiresAt = nil
		_ = s.repo.UpdateUser(ctx, user)
		return nil, ErrInvalidVerificationToken
	}

	user.IsVerified = true
	user.VerificationToken = nil
	user.TokenExpiresAt = nil
	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return nil, err
	}

	return &VerifyEmailResponse{
		UserID:     user.ID,
		IsVerified: true,
	}, nil
}

func (s *service) sendVerification(ctx context.Context, email, fullName, rawToken string) error {
	link := fmt.Sprintf("%s/api/v1/auth/verify-email?token=%s", s.appBaseURL, url.QueryEscape(rawToken))
	if err := s.emailService.SendVerificationEmail(ctx, email, fullName, link); err != nil {
		return fmt.Errorf("%w: %v", ErrVerificationEmailFailed, err)
	}
	return nil
}

func sanitizeAppBaseURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return strings.TrimRight(raw, "/")
	}

	// Keep only scheme + host to prevent malformed links like /swagger/#/... from env mistakes.
	u.Path = ""
	u.RawPath = ""
	u.RawQuery = ""
	u.Fragment = ""

	return strings.TrimRight(u.String(), "/")
}

func (s *service) issueToken(user *users.User) (string, error) {
	email := "social-auth@roamify.local"
	if user.Email != nil {
		email = *user.Email
	}
	return pkgjwt.Generate(user.ID, email, string(user.Role), s.jwtSecret, s.jwtExpiryHours)
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func normalizeOptionalEmail(email *string) *string {
	if email == nil {
		return nil
	}
	value := normalizeEmail(*email)
	return &value
}

func generateVerificationToken() (rawToken, hashedToken string, expiresAt time.Time, err error) {
	buf := make([]byte, 32)
	if _, err = rand.Read(buf); err != nil {
		return "", "", time.Time{}, err
	}

	rawToken = base64.RawURLEncoding.EncodeToString(buf)
	hashedToken = hashToken(rawToken)
	expiresAt = time.Now().UTC().Add(verificationTTL)
	return rawToken, hashedToken, expiresAt, nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
