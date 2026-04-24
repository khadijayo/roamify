package users

import (
	"errors"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Service interface {
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
	GetPublicProfile(userID uuid.UUID) (*User, error)
	SearchUsers(query string) ([]User, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
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
		vp.PreferredVibes = pq.StringArray(req.PreferredVibes)
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
		vp.Interests = pq.StringArray(req.Interests)
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

func (s *service) GetPublicProfile(userID uuid.UUID) (*User, error) {
	// Re-uses existing repo.FindByID — no new DB logic needed.
	return s.repo.FindByID(userID)
}

func (s *service) SearchUsers(query string) ([]User, error) {
	return s.repo.SearchUsers(query, 20)
}
