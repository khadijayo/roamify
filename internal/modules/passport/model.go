package passport

import (
	"time"

	"github.com/google/uuid"
)

type PassportVaultRecord struct {
	ID                         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID                     uuid.UUID `gorm:"type:uuid;uniqueIndex;not null"                 json:"user_id"`
	EncryptedPayload           string    `gorm:"type:text;not null"                             json:"-"`
	PassportImageURL           *string   `gorm:"type:text"                                      json:"passport_image_url"`
	ExtractedFullNameMasked    *string   `gorm:"type:varchar(255)"                              json:"extracted_full_name_masked"`
	ExtractedNationality       *string   `gorm:"type:varchar(100)"                              json:"extracted_nationality"`
	ExtractedPassportNumMasked *string   `gorm:"type:varchar(50)"                               json:"extracted_passport_number_masked"`
	ExtractedDateOfBirth       *string   `gorm:"type:varchar(20)"                               json:"extracted_date_of_birth"`
	ExtractedExpiryDate        *string   `gorm:"type:varchar(20)"                               json:"extracted_expiry_date"`
	ExtractedCountryCode       *string   `gorm:"type:char(3)"                                   json:"extracted_country_code"`
	LockVersion                int       `gorm:"default:1"                                      json:"lock_version"`
	CreatedAt                  time.Time `                                                      json:"created_at"`
	UpdatedAt                  time.Time `                                                      json:"updated_at"`
}

func (PassportVaultRecord) TableName() string { return "passport_vault_records" }

type PassportStamp struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"                    json:"id"`
	UserID      uuid.UUID `gorm:"type:uuid;not null;index"                                           json:"user_id"`
	Country     string    `gorm:"type:varchar(100);not null"                                         json:"country"`
	CountryCode string    `gorm:"type:char(3);not null"                                              json:"country_code"`
	DateVisited time.Time `gorm:"type:date;not null"                                                 json:"date_visited"`
	CreatedAt   time.Time `                                                                          json:"created_at"`
}

func (PassportStamp) TableName() string { return "passport_stamps" }

type UpsertVaultRequest struct {
	EncryptedPayload           string  `json:"encrypted_payload"            binding:"required"`
	PassportImageURL           *string `json:"passport_image_url"`
	ExtractedFullNameMasked    *string `json:"extracted_full_name_masked"`
	ExtractedNationality       *string `json:"extracted_nationality"`
	ExtractedPassportNumMasked *string `json:"extracted_passport_number_masked"`
	ExtractedDateOfBirth       *string `json:"extracted_date_of_birth"`
	ExtractedExpiryDate        *string `json:"extracted_expiry_date"`
	ExtractedCountryCode       *string `json:"extracted_country_code"`
}

type AddStampRequest struct {
	Country     string    `json:"country"      binding:"required"`
	CountryCode string    `json:"country_code" binding:"required"`
	DateVisited time.Time `json:"date_visited" binding:"required"`
}
