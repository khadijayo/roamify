package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/khadijayo/roamify/config"
	"github.com/khadijayo/roamify/internal/modules/users"
	"github.com/khadijayo/roamify/internal/services"
	pkgjwt "github.com/khadijayo/roamify/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

const (
	verificationTTL             = 24 * time.Hour
	resetCodeTTL                = 10 * time.Minute
	resetCodeMaxAttempts        = 5
	resetCodeRateLimit          = 1 * time.Minute
	resendVerificationRateLimit = 1 * time.Minute

	googleUserInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"
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

	// Google OAuth errors
	ErrGoogleTokenExchange = errors.New("failed to exchange Google authorization code")
	ErrGoogleUserInfo      = errors.New("failed to fetch user info from Google")
	ErrGoogleMissingEmail  = errors.New("Google account did not return an email address")
)

// ── Request / response types ─────────────────────────────────────────────────

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

// googleUserInfo is the shape returned by Google's /oauth2/v2/userinfo endpoint.
type googleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
}

// ── Service interface ─────────────────────────────────────────────────────────

type Service interface {
	Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error)
	Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error)
	SocialAuth(ctx context.Context, req *SocialAuthRequest) (*AuthResponse, error)
	VerifyEmail(ctx context.Context, token string) (*VerifyEmailResponse, error)
	ForgotPassword(ctx context.Context, req *ForgotPasswordRequest) (*ActionResponse, error)
	VerifyResetCode(ctx context.Context, req *VerifyResetCodeRequest) (*ActionResponse, error)
	ResetPassword(ctx context.Context, req *ResetPasswordRequest) (*ActionResponse, error)
	ResendVerification(ctx context.Context, req *ResendVerificationRequest) (*ActionResponse, error)

	// Google OAuth
	GoogleCallback(ctx context.Context, code string, oauthCfg *oauth2.Config) (*AuthResponse, error)
}

// ── Service implementation ────────────────────────────────────────────────────

type service struct {
	repo           Repository
	jwtSecret      string
	jwtExpiryHours int
	httpClient     *http.Client
}

func NewService(repo Repository, cfg *config.Config) Service {
	return &service{
		repo:           repo,
		jwtSecret:      cfg.JWTSecret,
		jwtExpiryHours: cfg.JWTExpiryHours,
		httpClient:     &http.Client{Timeout: 10 * time.Second},
	}
}

// ── Existing service methods (unchanged) ─────────────────────────────────────

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
		// Allow password setup for social auth accounts (verified, but no password)
		// or for unverified accounts during re-registration
		if existing.IsVerified && existing.PasswordHash != nil {
			return nil, ErrEmailAlreadyRegistered
		}

		hashValue := string(hash)
		existing.FullName = fullName
		existing.PasswordHash = &hashValue
		existing.VerificationToken = &hashedToken
		existing.TokenExpiresAt = &expiresAt
		existing.VerificationSentAt = ptrTime(time.Now().UTC())
		// Only reset IsVerified to false for unverified accounts; keep verified accounts verified
		if !existing.IsVerified {
			existing.IsVerified = false
		}
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
		dbRole := user.Role
		provider := req.Provider
		user.AuthProvider = &provider
		user.ProviderID = &req.ProviderUserID
		user.FullName = strings.TrimSpace(req.FullName)
		user.AvatarURL = req.AvatarURL
		user.IsVerified = true
		user.VerificationToken = nil
		user.TokenExpiresAt = nil
		user.Role = dbRole
		if err := s.repo.UpdateUser(ctx, user); err != nil {
			return nil, err
		}
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

	return &VerifyEmailResponse{IsVerified: true}, nil
}

// ── Google OAuth ──────────────────────────────────────────────────────────────

// GoogleCallback exchanges the code for a token, fetches user info from Google,
// then finds-or-creates the user and returns a signed JWT — exactly like SocialAuth.
func (s *service) GoogleCallback(ctx context.Context, code string, oauthCfg *oauth2.Config) (*AuthResponse, error) {
	// Step 1: Exchange authorization code for Google access token
	googleToken, err := oauthCfg.Exchange(ctx, code)
	if err != nil {
		log.Printf("[auth/google] token exchange failed: %v", err)
		return nil, ErrGoogleTokenExchange
	}

	// Step 2: Use the access token to fetch the user's Google profile
	info, err := s.fetchGoogleUserInfo(ctx, googleToken.AccessToken)
	if err != nil {
		log.Printf("[auth/google] userinfo fetch failed: %v", err)
		return nil, ErrGoogleUserInfo
	}

	// Step 3: Validate that we got an email back
	if info.Email == "" {
		return nil, ErrGoogleMissingEmail
	}

	email := normalizeEmail(info.Email)

	// Step 4: Resolve full name – fall back gracefully if Google omits fields
	fullName := strings.TrimSpace(info.Name)
	if fullName == "" {
		fullName = strings.TrimSpace(info.GivenName + " " + info.FamilyName)
	}
	if fullName == "" {
		fullName = email // last resort
	}

	// Step 5: Build a SocialAuthRequest and reuse the existing SocialAuth logic.
	// This keeps the upsert/create/ban logic in one place.
	avatarURL := info.Picture
	req := &SocialAuthRequest{
		Provider:       "google",
		ProviderUserID: info.ID,
		Email:          &email,
		FullName:       fullName,
		AvatarURL:      &avatarURL,
	}

	return s.SocialAuth(ctx, req)
}

// fetchGoogleUserInfo calls Google's userinfo endpoint and returns the profile.
func (s *service) fetchGoogleUserInfo(ctx context.Context, accessToken string) (*googleUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, googleUserInfoURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("google userinfo returned %d: %s", resp.StatusCode, string(body))
	}

	var info googleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("decode google userinfo: %w", err)
	}

	return &info, nil
}

// ── Internal helpers ──────────────────────────────────────────────────────────

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

// generateStateToken creates a random CSRF state token for OAuth.
func generateStateToken() string {
	buf := make([]byte, 16)
	_, _ = rand.Read(buf)
	return base64.RawURLEncoding.EncodeToString(buf)
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
