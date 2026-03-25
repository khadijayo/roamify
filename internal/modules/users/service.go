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
	SocialAuth(req *SocialAuthRequest) (*AuthResponse, error)
	GetProfile(userID uuid.UUID) (*User, error)
	UpdateProfile(userID uuid.UUID, req *UpdateProfileRequest) (*User, error)
	GetVibeProfile(userID uuid.UUID) (*VibeProfile, error)
	UpsertVibeProfile(userID uuid.UUID, req *UpdateVibeProfileRequest) (*VibeProfile, error)
	FollowUser(followerID uuid.UUID, req *FollowUserRequest) error
	UnfollowUser(followerID, followingID uuid.UUID) error
	GetFollowers(userID uuid.UUID) ([]User, error)
	GetFollowing(userID uuid.UUID) ([]User, error)
	GetPrivacySettings(userID uuid.UUID) (*UserPrivacySetting, error)
	UpdatePrivacySettings(userID uuid.UUID, req *UpdatePrivacySettingsRequest) (*UserPrivacySetting, error)
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

func (s *service) SocialAuth(req *SocialAuthRequest) (*AuthResponse, error) {
	user, err := s.repo.FindByProvider(req.Provider, req.ProviderUserID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if user == nil {
		if req.Email != nil {
			existingByEmail, emailErr := s.repo.FindByEmail(*req.Email)
			if emailErr == nil && existingByEmail != nil {
				provider := req.Provider
				existingByEmail.AuthProvider = &provider
				existingByEmail.ProviderID = &req.ProviderUserID
				if req.AvatarURL != nil {
					existingByEmail.AvatarURL = req.AvatarURL
				}
				if err := s.repo.UpdateUser(existingByEmail); err != nil {
					return nil, err
				}
				user = existingByEmail
			}
		}

		if user == nil {
			provider := req.Provider
			user = &User{
				FullName:     req.FullName,
				Email:        req.Email,
				AvatarURL:    req.AvatarURL,
				AuthProvider: &provider,
				ProviderID:   &req.ProviderUserID,
				Status:       StatusActive,
			}
			if err := s.repo.CreateUser(user); err != nil {
				return nil, err
			}
		}
	}

	emailForToken := "social-auth@roamify.local"
	if user.Email != nil {
		emailForToken = *user.Email
	}
	token, err := pkgjwt.Generate(user.ID, emailForToken, s.jwtSecret, s.jwtExpiryHours)
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

func (s *service) FollowUser(followerID uuid.UUID, req *FollowUserRequest) error {
	if followerID == req.UserID {
		return errors.New("you cannot follow yourself")
	}
	_, err := s.repo.FindByID(req.UserID)
	if err != nil {
		return errors.New("target user not found")
	}
	_, err = s.repo.FindFollow(followerID, req.UserID)
	if err == nil {
		return errors.New("already following this user")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return s.repo.CreateFollow(&UserFollow{FollowerID: followerID, FollowingID: req.UserID})
}

func (s *service) UnfollowUser(followerID, followingID uuid.UUID) error {
	_, err := s.repo.FindFollow(followerID, followingID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("not following this user")
		}
		return err
	}
	return s.repo.DeleteFollow(followerID, followingID)
}

func (s *service) GetFollowers(userID uuid.UUID) ([]User, error) {
	rels, err := s.repo.ListFollowers(userID)
	if err != nil {
		return nil, err
	}
	users := make([]User, 0, len(rels))
	for _, rel := range rels {
		u, findErr := s.repo.FindByID(rel.FollowerID)
		if findErr == nil && u != nil {
			users = append(users, *u)
		}
	}
	return users, nil
}

func (s *service) GetFollowing(userID uuid.UUID) ([]User, error) {
	rels, err := s.repo.ListFollowing(userID)
	if err != nil {
		return nil, err
	}
	users := make([]User, 0, len(rels))
	for _, rel := range rels {
		u, findErr := s.repo.FindByID(rel.FollowingID)
		if findErr == nil && u != nil {
			users = append(users, *u)
		}
	}
	return users, nil
}

func (s *service) GetPrivacySettings(userID uuid.UUID) (*UserPrivacySetting, error) {
	settings, err := s.repo.GetPrivacySettings(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			defaults := &UserPrivacySetting{
				UserID:             userID,
				GhostModeEnabled:   false,
				DataSharingEnabled: true,
				MapVisibility:      "public",
			}
			if upErr := s.repo.UpsertPrivacySettings(defaults); upErr != nil {
				return nil, upErr
			}
			return defaults, nil
		}
		return nil, err
	}
	return settings, nil
}

func (s *service) UpdatePrivacySettings(userID uuid.UUID, req *UpdatePrivacySettingsRequest) (*UserPrivacySetting, error) {
	settings, err := s.GetPrivacySettings(userID)
	if err != nil {
		return nil, err
	}
	if req.GhostModeEnabled != nil {
		settings.GhostModeEnabled = *req.GhostModeEnabled
	}
	if req.DataSharingEnabled != nil {
		settings.DataSharingEnabled = *req.DataSharingEnabled
	}
	if req.MapVisibility != nil && *req.MapVisibility != "" {
		settings.MapVisibility = *req.MapVisibility
	}
	if err := s.repo.UpsertPrivacySettings(settings); err != nil {
		return nil, err
	}
	return settings, nil
}
