package passport

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/khadijayo/roamify/internal/modules/users"
	"gorm.io/gorm"
)

type Service interface {
	UpsertVault(userID uuid.UUID, req *UpsertVaultRequest) (*PassportVaultRecord, error)
	GetVault(userID uuid.UUID) (*PassportVaultRecord, error)
	DeleteVault(userID uuid.UUID) error

	AddStamp(userID uuid.UUID, req *AddStampRequest) (*PassportStamp, error)
	GetStamps(userID uuid.UUID) ([]PassportStamp, error)
	DeleteStamp(stampID, userID uuid.UUID) error
}

type service struct {
	repo     Repository
	userRepo users.Repository
}

func NewService(repo Repository, userRepo users.Repository) Service {
	return &service{repo: repo, userRepo: userRepo}
}

func (s *service) UpsertVault(userID uuid.UUID, req *UpsertVaultRequest) (*PassportVaultRecord, error) {
	existing, err := s.repo.FindVaultByUser(userID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	record := &PassportVaultRecord{
		UserID:                     userID,
		EncryptedPayload:           req.EncryptedPayload,
		PassportImageURL:           req.PassportImageURL,
		ExtractedFullNameMasked:    req.ExtractedFullNameMasked,
		ExtractedNationality:       req.ExtractedNationality,
		ExtractedPassportNumMasked: req.ExtractedPassportNumMasked,
		ExtractedDateOfBirth:       req.ExtractedDateOfBirth,
		ExtractedExpiryDate:        req.ExtractedExpiryDate,
		ExtractedCountryCode:       req.ExtractedCountryCode,
	}

	if existing != nil {
		record.ID = existing.ID
		record.LockVersion = existing.LockVersion + 1
	}

	if err := s.repo.UpsertVault(record); err != nil {
		return nil, err
	}
	return record, nil
}

func (s *service) GetVault(userID uuid.UUID) (*PassportVaultRecord, error) {
	record, err := s.repo.FindVaultByUser(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("no passport vault found")
		}
		return nil, err
	}
	return record, nil
}

func (s *service) DeleteVault(userID uuid.UUID) error {
	return s.repo.DeleteVault(userID)
}

func (s *service) AddStamp(userID uuid.UUID, req *AddStampRequest) (*PassportStamp, error) {
	dateStr := req.DateVisited.Format("2006-01-02")
	countryCode := strings.ToUpper(req.CountryCode)

	existing, err := s.repo.FindStamp(userID, countryCode, dateStr)
	if err == nil && existing != nil {
		return nil, errors.New("stamp already exists for this country and date")
	}

	stamp := &PassportStamp{
		UserID:      userID,
		Country:     req.Country,
		CountryCode: countryCode,
		DateVisited: req.DateVisited,
	}
	if err := s.repo.AddStamp(stamp); err != nil {
		return nil, err
	}

	count, _ := s.repo.CountStampsByUser(userID)
	vp, err := s.userRepo.GetVibeProfile(userID)
	if err == nil {
		vp.CountriesVisited = int(count)
		_ = s.userRepo.UpsertVibeProfile(vp)
	}

	return stamp, nil
}

func (s *service) GetStamps(userID uuid.UUID) ([]PassportStamp, error) {
	return s.repo.FindStampsByUser(userID)
}

func (s *service) DeleteStamp(stampID, userID uuid.UUID) error {
	stamps, err := s.repo.FindStampsByUser(userID)
	if err != nil {
		return err
	}

	found := false
	for _, st := range stamps {
		if st.ID == stampID {
			found = true
			break
		}
	}
	if !found {
		return errors.New("stamp not found or not authorized")
	}
	if err := s.repo.DeleteStamp(stampID); err != nil {
		return err
	}

	count, _ := s.repo.CountStampsByUser(userID)
	vp, err := s.userRepo.GetVibeProfile(userID)
	if err == nil {
		vp.CountriesVisited = int(count)
		_ = s.userRepo.UpsertVibeProfile(vp)
	}
	return nil
}
