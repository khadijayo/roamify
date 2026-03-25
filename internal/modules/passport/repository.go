package passport

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	UpsertVault(record *PassportVaultRecord) error
	FindVaultByUser(userID uuid.UUID) (*PassportVaultRecord, error)
	DeleteVault(userID uuid.UUID) error

	AddStamp(stamp *PassportStamp) error
	FindStampsByUser(userID uuid.UUID) ([]PassportStamp, error)
	FindStamp(userID uuid.UUID, countryCode string, dateVisited string) (*PassportStamp, error)
	DeleteStamp(stampID uuid.UUID) error
	CountStampsByUser(userID uuid.UUID) (int64, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) UpsertVault(record *PassportVaultRecord) error {

	return r.db.Save(record).Error
}

func (r *repository) FindVaultByUser(userID uuid.UUID) (*PassportVaultRecord, error) {
	var record PassportVaultRecord
	err := r.db.Where("user_id = ?", userID).First(&record).Error
	return &record, err
}

func (r *repository) DeleteVault(userID uuid.UUID) error {
	return r.db.Where("user_id = ?", userID).Delete(&PassportVaultRecord{}).Error
}

func (r *repository) AddStamp(stamp *PassportStamp) error {
	return r.db.Create(stamp).Error
}

func (r *repository) FindStampsByUser(userID uuid.UUID) ([]PassportStamp, error) {
	var stamps []PassportStamp
	err := r.db.Where("user_id = ?", userID).
		Order("date_visited DESC").
		Find(&stamps).Error
	return stamps, err
}

func (r *repository) FindStamp(userID uuid.UUID, countryCode string, dateVisited string) (*PassportStamp, error) {
	var stamp PassportStamp
	err := r.db.Where("user_id = ? AND country_code = ? AND date_visited = ?", userID, countryCode, dateVisited).
		First(&stamp).Error
	return &stamp, err
}

func (r *repository) DeleteStamp(stampID uuid.UUID) error {
	return r.db.Delete(&PassportStamp{}, "id = ?", stampID).Error
}

func (r *repository) CountStampsByUser(userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&PassportStamp{}).
		Where("user_id = ?", userID).
		Distinct("country_code").
		Count(&count).Error
	return count, err
}
