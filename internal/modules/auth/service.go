package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/khadijayo/roamify/config"
	"github.com/khadijayo/roamify/internal/modules/users"
	"github.com/khadijayo/roamify/internal/services"
	pkgjwt "github.com/khadijayo/roamify/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	verificationTTL             = 24 * time.Hour
	resetCodeTTL                = 10 * time.Minute
	resetCodeMaxAttempts        = 5
	resetCodeRateLimit          = 1 * time.Minute
	resendVerificationRateLimit = 1 * time.Minute
)

var (
	ErrInvalidCredentials            = errors.New("invalid email or password")
	ErrEmailAlreadyRegistered        = errors.New("email already registered")
	ErrEmailNotVerified              = errors.New("please verify your email before logging in")
	ErrAccountBanned                 = errors.New("your account has been banned")
	ErrInvalidVerificationToken      = errors.New("invalid or expired verification token")
	ErrVerificationEmailFailed       = errors.New("failed to send verification email")
	ErrInvalidResetCode              = errors.New("invalid or expired reset code")
	ErrTooManyResetAttempts          = errors.New("too many invalid reset attempts")
	ErrResetCodeSendFailed           = errors.New("failed to send password reset code")
	ErrResetCodeRecentlySent         = errors.New("please wait before requesting another reset code")
	ErrVerificationResendRateLimited = errors.New("please wait before resending verification email")
	ErrVerificationResendFailed      = errors.New("failed to resend verification email")
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

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type VerifyResetCodeRequest struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required,len=6"`
}

type ResetPasswordRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Code        string `json:"code" binding:"required,len=6"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

type ResendVerificationRequest struct {
	Email string `json:"email" binding:"required,email"`
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
	IsVerified bool `json:"is_verified"`
}

type ActionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type Service interface {
	Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error)
	Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error)
	SocialAuth(ctx context.Context, req *SocialAuthRequest) (*AuthResponse, error)
	VerifyEmail(ctx context.Context, token string) (*VerifyEmailResponse, error)
	ForgotPassword(ctx context.Context, req *ForgotPasswordRequest) (*ActionResponse, error)
	VerifyResetCode(ctx context.Context, req *VerifyResetCodeRequest) (*ActionResponse, error)
	ResetPassword(ctx context.Context, req *ResetPasswordRequest) (*ActionResponse, error)
	ResendVerification(ctx context.Context, req *ResendVerificationRequest) (*ActionResponse, error)
}

type service struct {
	repo           Repository
	jwtSecret      string
	jwtExpiryHours int
}

func NewService(repo Repository, cfg *config.Config) Service {
	return &service{
		repo:           repo,
		jwtSecret:      cfg.JWTSecret,
		jwtExpiryHours: cfg.JWTExpiryHours,
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
		existing.VerificationSentAt = ptrTime(time.Now().UTC())
		existing.IsVerified = false
		existing.IsBanned = false
		existing.Role = users.RoleUser

		if err := s.repo.UpdateUser(ctx, existing); err != nil {
			return nil, err
		}

		s.sendVerificationAsync(email, rawToken)

		return &RegisterResponse{
			User:                  existing,
			VerificationSent:      true,
			VerificationExpiresAt: expiresAt,
		}, nil
	}

	hashValue := string(hash)
	user := &users.User{
		FullName:           fullName,
		Email:              &email,
		PasswordHash:       &hashValue,
		Role:               users.RoleUser,
		Status:             users.StatusActive,
		IsVerified:         false,
		IsBanned:           false,
		VerificationToken:  &hashedToken,
		TokenExpiresAt:     &expiresAt,
		VerificationSentAt: ptrTime(time.Now().UTC()),
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	s.sendVerificationAsync(email, rawToken)

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

	now := time.Now().UTC()
	user.LastLoginAt = &now
	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return nil, err
	}

	// Re-fetch from DB to guarantee the freshest role is in the token.
	// Prevents stale JWT role when role was updated externally (e.g. via pgAdmin).
	if fresh, fetchErr := s.repo.FindByID(ctx, user.ID); fetchErr == nil {
		log.Printf("[auth] login user_id=%s db_role=%s", fresh.ID, fresh.Role)
		user = fresh
	}

	token, err := s.issueToken(user)
	if err != nil {
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
		// Preserve the DB role — do NOT overwrite it during social login.
		// Only update profile/auth fields, never touch role.
		dbRole := user.Role
		provider := req.Provider
		user.AuthProvider = &provider
		user.ProviderID = &req.ProviderUserID
		user.FullName = strings.TrimSpace(req.FullName)
		user.AvatarURL = req.AvatarURL
		user.IsVerified = true
		user.VerificationToken = nil
		user.TokenExpiresAt = nil
		user.Role = dbRole // explicitly restore role to prevent accidental override
		if err := s.repo.UpdateUser(ctx, user); err != nil {
			return nil, err
		}
		// Re-fetch fresh from DB so issueToken uses the latest role
		if fresh, fetchErr := s.repo.FindByID(ctx, user.ID); fetchErr == nil {
			user = fresh
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

func (s *service) ForgotPassword(ctx context.Context, req *ForgotPasswordRequest) (*ActionResponse, error) {
	email := normalizeEmail(req.Email)
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &ActionResponse{Success: true, Message: "If an account exists, a reset code has been sent."}, nil
		}
		return nil, err
	}

	existing, err := s.repo.FindLatestPasswordResetCodeByEmail(ctx, email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existing != nil && existing.ExpiresAt.After(time.Now().UTC()) {
		if time.Since(existing.CreatedAt) < resetCodeRateLimit {
			return nil, ErrResetCodeRecentlySent
		}
	}

	rawCode, hashedCode, expiresAt, err := generateResetCode()
	if err != nil {
		return nil, err
	}

	reset := &users.PasswordResetCode{
		UserID:     user.ID,
		Email:      email,
		HashedCode: hashedCode,
		ExpiresAt:  expiresAt,
		Attempts:   0,
	}
	if err := s.repo.CreatePasswordResetCode(ctx, reset); err != nil {
		return nil, err
	}

	if err := services.SendPasswordResetCode(email, user.FullName, rawCode); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrResetCodeSendFailed, err)
	}

	return &ActionResponse{Success: true, Message: "If an account exists, a reset code has been sent."}, nil
}

func (s *service) VerifyResetCode(ctx context.Context, req *VerifyResetCodeRequest) (*ActionResponse, error) {
	email := normalizeEmail(req.Email)
	code := strings.TrimSpace(req.Code)
	if code == "" {
		return nil, ErrInvalidResetCode
	}

	reset, err := s.repo.FindLatestPasswordResetCodeByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidResetCode
		}
		return nil, err
	}

	if reset == nil || reset.UsedAt != nil || time.Now().UTC().After(reset.ExpiresAt) {
		return nil, ErrInvalidResetCode
	}

	if reset.Attempts >= resetCodeMaxAttempts {
		return nil, ErrTooManyResetAttempts
	}

	if !compareHashCode(code, reset.HashedCode) {
		reset.Attempts++
		if err := s.repo.UpdatePasswordResetCode(ctx, reset); err != nil {
			return nil, err
		}
		if reset.Attempts >= resetCodeMaxAttempts {
			return nil, ErrTooManyResetAttempts
		}
		return nil, ErrInvalidResetCode
	}

	return &ActionResponse{Success: true, Message: "Reset code is valid."}, nil
}

func (s *service) ResetPassword(ctx context.Context, req *ResetPasswordRequest) (*ActionResponse, error) {
	email := normalizeEmail(req.Email)
	code := strings.TrimSpace(req.Code)
	if code == "" {
		return nil, ErrInvalidResetCode
	}

	reset, err := s.repo.FindLatestPasswordResetCodeByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidResetCode
		}
		return nil, err
	}

	if reset == nil || reset.UsedAt != nil || time.Now().UTC().After(reset.ExpiresAt) {
		return nil, ErrInvalidResetCode
	}

	if reset.Attempts >= resetCodeMaxAttempts {
		return nil, ErrTooManyResetAttempts
	}

	if !compareHashCode(code, reset.HashedCode) {
		reset.Attempts++
		if err := s.repo.UpdatePasswordResetCode(ctx, reset); err != nil {
			return nil, err
		}
		if reset.Attempts >= resetCodeMaxAttempts {
			return nil, ErrTooManyResetAttempts
		}
		return nil, ErrInvalidResetCode
	}

	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, ErrInvalidResetCode
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	hashValue := string(hash)
	user.PasswordHash = &hashValue
	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return nil, err
	}

	if err := s.repo.DeletePasswordResetCodesByEmail(ctx, email); err != nil {
		return nil, err
	}

	return &ActionResponse{Success: true, Message: "Password reset successfully."}, nil
}

func (s *service) ResendVerification(ctx context.Context, req *ResendVerificationRequest) (*ActionResponse, error) {
	email := normalizeEmail(req.Email)
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &ActionResponse{Success: true, Message: "If an account exists, verification will be resent."}, nil
		}
		return nil, err
	}

	if user.IsVerified {
		return &ActionResponse{Success: true, Message: "Email already verified."}, nil
	}

	if user.VerificationSentAt != nil && time.Since(*user.VerificationSentAt) < resendVerificationRateLimit {
		return nil, ErrVerificationResendRateLimited
	}

	rawToken, hashedToken, expiresAt, err := generateVerificationToken()
	if err != nil {
		return nil, err
	}

	user.VerificationToken = &hashedToken
	user.TokenExpiresAt = &expiresAt
	now := time.Now().UTC()
	user.VerificationSentAt = &now
	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return nil, err
	}

	if err := services.SendVerificationEmail(email, rawToken); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrVerificationResendFailed, err)
	}

	return &ActionResponse{Success: true, Message: "Verification email resent."}, nil
}

func (s *service) VerifyEmail(ctx context.Context, token string) (*VerifyEmailResponse, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, ErrInvalidVerificationToken
	}

	user, err := s.findUserByVerificationToken(ctx, token)
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
		IsVerified: true,
	}, nil
}

func (s *service) sendVerificationAsync(email, rawToken string) {
	go func() {
		if err := services.SendVerificationEmail(email, rawToken); err != nil {
			log.Println("email failed:", err)
		}
	}()
}

func (s *service) findUserByVerificationToken(ctx context.Context, token string) (*users.User, error) {
	user, err := s.repo.FindByVerificationToken(ctx, hashToken(token))
	if err == nil {
		return user, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	return s.repo.FindByVerificationToken(ctx, token)
}

func (s *service) issueToken(user *users.User) (string, error) {
	email := "social-auth@roamify.local"
	if user.Email != nil {
		email = normalizeEmail(*user.Email)
	}

	role := strings.ToLower(strings.TrimSpace(string(user.Role)))
	if role == "" {
		role = string(users.RoleUser)
	}
	return pkgjwt.Generate(user.ID, email, role, s.jwtSecret, s.jwtExpiryHours)
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

func generateResetCode() (string, string, time.Time, error) {
	number, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", "", time.Time{}, err
	}

	raw := fmt.Sprintf("%06d", number.Int64())
	hashed := hashCode(raw)
	expiresAt := time.Now().UTC().Add(resetCodeTTL)
	return raw, hashed, expiresAt, nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func hashCode(code string) string {
	sum := sha256.Sum256([]byte(code))
	return hex.EncodeToString(sum[:])
}

func compareHashCode(code, hashed string) bool {
	return hashCode(code) == hashed
}

func ptrTime(t time.Time) *time.Time {
	return &t
}