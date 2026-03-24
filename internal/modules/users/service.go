package users

import (
	"errors"
	"time"

	"github.com/google/uuid"
	pkgjwt "github.com/khadijayo/roamify/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Service interface {
	Register(req *RegisterRequest) (*AuthResponse, error)
	Login(req *LoginRequest) (*AuthResponse, error)
	GetProfile(userID uuid.UUID) (*User, error)
	UpdateProfile(userID uuid.UUID, req *UpdateProfileRequest) (*User, error)
	GetVibeProfile(userID uuid.UUID) (*VibeProfile, error)
	UpsertVibeProfile(userID uuid.UUID, req *UpdateVibeProfileRequest) (*VibeProfile, error)
}

type service struct {
	repo           Repository
	jwtSecret      string
	jwtExpiryHours int
}

func NewService(repo Repository, jwtSecret string, jwtExpiryHours int) Service {
	return &service{
		repo:           repo,
		jwtSecret:      jwtSecret,
		jwtExpiryHours: jwtExpiryHours,
	}
}

func (s *service) Register(req *RegisterRequest) (*AuthResponse, error) {
	existing, err := s.repo.FindByEmail(req.Email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("email already registered")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	hashStr := string(hash)
	email := req.Email
	user := &User{
		FullName:     req.FullName,
		Email:        &email,
		PasswordHash: &hashStr,
		Status:       StatusActive,
	}

	if err := s.repo.CreateUser(user); err != nil {
		return nil, err
	}

	token, err := pkgjwt.Generate(user.ID, req.Email, s.jwtSecret, s.jwtExpiryHours)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{Token: token, User: user}, nil
}

func (s *service) Login(req *LoginRequest) (*AuthResponse, error) {
	user, err := s.repo.FindByEmail(req.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}
	if user.PasswordHash == nil {
		return nil, errors.New("this account uses social login")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	now := time.Now()
	user.LastLoginAt = &now
	_ = s.repo.UpdateUser(user)

	token, err := pkgjwt.Generate(user.ID, *user.Email, s.jwtSecret, s.jwtExpiryHours)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{Token: token, User: user}, nil
}

func (s *service) GetProfile(userID uuid.UUID) (*User, error) {
	return s.repo.FindByID(userID)
}

func (s *service) UpdateProfile(userID uuid.UUID, req *UpdateProfileRequest) (*User, error) {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if req.FullName != "" {
		user.FullName = req.FullName
	}
	if req.AvatarURL != nil {
		user.AvatarURL = req.AvatarURL
	}
	if err := s.repo.UpdateUser(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *service) GetVibeProfile(userID uuid.UUID) (*VibeProfile, error) {
	return s.repo.GetVibeProfile(userID)
}

func (s *service) UpsertVibeProfile(userID uuid.UUID, req *UpdateVibeProfileRequest) (*VibeProfile, error) {
	vp, err := s.repo.GetVibeProfile(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			vp = &VibeProfile{UserID: userID}
		} else {
			return nil, err
		}
	}

	if req.ExplorerType != "" {
		vp.ExplorerType = req.ExplorerType
	}
	if req.PreferredVibes != nil {
		vp.PreferredVibes = req.PreferredVibes
	}
	if req.TravelPace != "" {
		vp.TravelPace = req.TravelPace
	}
	if req.BudgetStyle != "" {
		vp.BudgetStyle = req.BudgetStyle
	}
	if req.TravelWith != "" {
		vp.TravelWith = req.TravelWith
	}
	if req.Interests != nil {
		vp.Interests = req.Interests
	}
	if req.OnboardingComplete != nil {
		vp.OnboardingComplete = *req.OnboardingComplete
	}

	if err := s.repo.UpsertVibeProfile(vp); err != nil {
		return nil, err
	}
	return vp, nil
}