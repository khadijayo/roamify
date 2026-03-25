package users

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	CreateUser(user *User) error
	FindByEmail(email string) (*User, error)
	FindByProvider(provider, providerID string) (*User, error)
	FindByID(id uuid.UUID) (*User, error)
	UpdateUser(user *User) error
	UpsertVibeProfile(vp *VibeProfile) error
	GetVibeProfile(userID uuid.UUID) (*VibeProfile, error)
	ListTopVibeProfiles(limit int) ([]VibeProfile, error)
	CreateFollow(f *UserFollow) error
	DeleteFollow(followerID, followingID uuid.UUID) error
	FindFollow(followerID, followingID uuid.UUID) (*UserFollow, error)
	ListFollowers(userID uuid.UUID) ([]UserFollow, error)
	ListFollowing(userID uuid.UUID) ([]UserFollow, error)
	CountFollowers(userID uuid.UUID) (int64, error)
	CountFollowing(userID uuid.UUID) (int64, error)
	GetPrivacySettings(userID uuid.UUID) (*UserPrivacySetting, error)
	UpsertPrivacySettings(settings *UserPrivacySetting) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateUser(user *User) error {
	return r.db.Create(user).Error
}

func (r *repository) FindByEmail(email string) (*User, error) {
	var user User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *repository) FindByProvider(provider, providerID string) (*User, error) {
	var user User
	if err := r.db.Where("auth_provider = ? AND provider_id = ?", provider, providerID).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *repository) FindByID(id uuid.UUID) (*User, error) {
	var user User
	if err := r.db.Preload("VibeProfile").First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *repository) UpdateUser(user *User) error {
	return r.db.Save(user).Error
}

func (r *repository) UpsertVibeProfile(vp *VibeProfile) error {
	return r.db.Save(vp).Error
}

func (r *repository) GetVibeProfile(userID uuid.UUID) (*VibeProfile, error) {
	var vp VibeProfile
	if err := r.db.Where("user_id = ?", userID).First(&vp).Error; err != nil {
		return nil, err
	}
	return &vp, nil
}

func (r *repository) ListTopVibeProfiles(limit int) ([]VibeProfile, error) {
	if limit < 1 || limit > 100 {
		limit = 20
	}
	var rows []VibeProfile
	err := r.db.Order("roamify_points DESC, explorer_level DESC").Limit(limit).Find(&rows).Error
	return rows, err
}

func (r *repository) CreateFollow(f *UserFollow) error {
	return r.db.Create(f).Error
}

func (r *repository) DeleteFollow(followerID, followingID uuid.UUID) error {
	return r.db.Where("follower_id = ? AND following_id = ?", followerID, followingID).Delete(&UserFollow{}).Error
}

func (r *repository) FindFollow(followerID, followingID uuid.UUID) (*UserFollow, error) {
	var f UserFollow
	err := r.db.Where("follower_id = ? AND following_id = ?", followerID, followingID).First(&f).Error
	return &f, err
}

func (r *repository) ListFollowers(userID uuid.UUID) ([]UserFollow, error) {
	var rows []UserFollow
	err := r.db.Where("following_id = ?", userID).Order("created_at DESC").Find(&rows).Error
	return rows, err
}

func (r *repository) ListFollowing(userID uuid.UUID) ([]UserFollow, error) {
	var rows []UserFollow
	err := r.db.Where("follower_id = ?", userID).Order("created_at DESC").Find(&rows).Error
	return rows, err
}

func (r *repository) CountFollowers(userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&UserFollow{}).Where("following_id = ?", userID).Count(&count).Error
	return count, err
}

func (r *repository) CountFollowing(userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&UserFollow{}).Where("follower_id = ?", userID).Count(&count).Error
	return count, err
}

func (r *repository) GetPrivacySettings(userID uuid.UUID) (*UserPrivacySetting, error) {
	var settings UserPrivacySetting
	err := r.db.Where("user_id = ?", userID).First(&settings).Error
	return &settings, err
}

func (r *repository) UpsertPrivacySettings(settings *UserPrivacySetting) error {
	return r.db.Save(settings).Error
}
