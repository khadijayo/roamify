package auth

import (
	"context"

	"github.com/google/uuid"
	"github.com/khadijayo/roamify/internal/modules/users"
	"gorm.io/gorm"
)

type Repository interface {
	CreateUser(ctx context.Context, user *users.User) error
	UpdateUser(ctx context.Context, user *users.User) error
	FindByEmail(ctx context.Context, email string) (*users.User, error)
	FindByProvider(ctx context.Context, provider, providerID string) (*users.User, error)
	FindByVerificationToken(ctx context.Context, token string) (*users.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*users.User, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateUser(ctx context.Context, user *users.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *repository) UpdateUser(ctx context.Context, user *users.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *repository) FindByEmail(ctx context.Context, email string) (*users.User, error) {
	var user users.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *repository) FindByProvider(ctx context.Context, provider, providerID string) (*users.User, error) {
	var user users.User
	if err := r.db.WithContext(ctx).
		Where("auth_provider = ? AND provider_id = ?", provider, providerID).
		First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *repository) FindByVerificationToken(ctx context.Context, token string) (*users.User, error) {
	var user users.User
	if err := r.db.WithContext(ctx).
		Where("verification_token = ?", token).
		First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *repository) FindByID(ctx context.Context, id uuid.UUID) (*users.User, error) {
	var user users.User
	if err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
