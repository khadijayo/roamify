package users

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	CreateUser(user *User) error
	FindByEmail(email string) (*User, error)
	FindByID(id uuid.UUID) (*User, error)
	UpdateUser(user *User) error
	UpsertVibeProfile(vp *VibeProfile) error
	GetVibeProfile(userID uuid.UUID) (*VibeProfile, error)
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