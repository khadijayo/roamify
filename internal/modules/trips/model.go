package trips

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TripStatus string
type TripType string
type Role string
type JoinStatus string
type ItemType string

const (
	TripStatusPlanning  TripStatus = "planning"
	TripStatusOngoing   TripStatus = "ongoing"
	TripStatusCompleted TripStatus = "completed"
	TripStatusArchived  TripStatus = "archived"

	RoleOwner  Role = "owner"
	RoleAdmin  Role = "admin"
	RoleMember Role = "member"

	JoinStatusInvited  JoinStatus = "invited"
	JoinStatusJoined   JoinStatus = "joined"
	JoinStatusDeclined JoinStatus = "declined"

	ItemTypeActivity ItemType = "activity"
	ItemTypeFood     ItemType = "food"
	ItemTypeHotel    ItemType = "hotel"
)

type Trip struct {
	ID               uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OwnerUserID      uuid.UUID      `gorm:"type:uuid;not null;index"                       json:"owner_user_id"`
	Title            string         `gorm:"type:varchar(255);not null"                     json:"title"`
	Destination      string         `gorm:"type:varchar(255);not null"                     json:"destination"`
	TripType         TripType       `gorm:"type:varchar(50)"                               json:"trip_type"`
	VibeTags         []string       `gorm:"type:text[];serializer:json"                    json:"vibe_tags"`
	TravelersPlanned int            `gorm:"default:1"                                      json:"travelers_planned"`
	StartDate        *time.Time     `gorm:"type:timestamp"                                 json:"start_date"`
	EndDate          *time.Time     `gorm:"type:timestamp"                                 json:"end_date"`
	Budget           float64        `gorm:"type:numeric(12,2);default:0"                   json:"budget"`
	Spent            float64        `gorm:"type:numeric(12,2);default:0"                   json:"spent"`
	CoverImageURL    *string        `gorm:"type:text"                                      json:"cover_image_url"`
	Notes            *string        `gorm:"type:text"                                      json:"notes"`
	Status           TripStatus     `gorm:"type:varchar(20);default:'planning'"            json:"status"`
	CreatedAt        time.Time      `                                                      json:"created_at"`
	UpdatedAt        time.Time      `                                                      json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index"                                          json:"-"`

	// Associations
	Members        []TripMember        `gorm:"foreignKey:TripID" json:"members,omitempty"`
	ItineraryItems []TripItineraryItem `gorm:"foreignKey:TripID" json:"itinerary_items,omitempty"`
	Expenses       []TripExpense       `gorm:"foreignKey:TripID" json:"expenses,omitempty"`
}

type TripMember struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TripID     uuid.UUID  `gorm:"type:uuid;not null;index"                       json:"trip_id"`
	UserID     uuid.UUID  `gorm:"type:uuid;not null;index"                       json:"user_id"`
	Role       Role       `gorm:"type:varchar(20);default:'member'"              json:"role"`
	JoinStatus JoinStatus `gorm:"type:varchar(20);default:'invited'"             json:"join_status"`
	JoinedAt   *time.Time `                                                      json:"joined_at"`
	CreatedAt  time.Time  `                                                      json:"created_at"`
	UpdatedAt  time.Time  `                                                      json:"updated_at"`
}

type TripItineraryItem struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TripID          uuid.UUID  `gorm:"type:uuid;not null;index"                       json:"trip_id"`
	DayNumber       int        `gorm:"not null"                                        json:"day_number"`
	Title           string     `gorm:"type:varchar(255);not null"                      json:"title"`
	ItemType        ItemType   `gorm:"type:varchar(50)"                                json:"item_type"`
	StartTime       *time.Time `gorm:"type:timestamp"                                  json:"start_time"`
	LocationName    string     `gorm:"type:varchar(255)"                               json:"location_name"`
	Notes           *string    `gorm:"type:text"                                       json:"notes"`
	Lat             *float64   `gorm:"type:numeric(10,8)"                              json:"lat"`
	Lng             *float64   `gorm:"type:numeric(11,8)"                              json:"lng"`
	SortOrder       int        `gorm:"default:0"                                       json:"sort_order"`
	CreatedByUserID *uuid.UUID `gorm:"type:uuid;index"                                json:"created_by_user_id"`
	CreatedAt       time.Time  `                                                      json:"created_at"`
	UpdatedAt       time.Time  `                                                      json:"updated_at"`
}

type TripExpense struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TripID          uuid.UUID  `gorm:"type:uuid;not null;index"                       json:"trip_id"`
	CreatedByUserID *uuid.UUID `gorm:"type:uuid;index"                                json:"created_by_user_id"`
	Description     string     `gorm:"type:varchar(255);not null"                      json:"description"`
	Category        string     `gorm:"type:varchar(100)"                               json:"category"`
	Amount          float64    `gorm:"type:numeric(12,2);not null"                     json:"amount"`
	ExpenseDate     time.Time  `gorm:"type:date;not null"                              json:"expense_date"`
	CurrencyCode    string     `gorm:"type:varchar(3);default:'USD'"                   json:"currency_code"`
	CreatedAt       time.Time  `                                                      json:"created_at"`
	UpdatedAt       time.Time  `                                                      json:"updated_at"`
}

type ChatMessage struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TripID    uuid.UUID `gorm:"type:uuid;not null;index"                       json:"trip_id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null"                             json:"user_id"`
	Message   string    `gorm:"type:text;not null"                             json:"message"`
	CreatedAt time.Time `                                                      json:"created_at"`
}

// ---------------------------------------------------------------------------
// DTOs
// ---------------------------------------------------------------------------

type CreateTripRequest struct {
	Title            string     `json:"title"            binding:"required"`
	Destination      string     `json:"destination"      binding:"required"`
	TripType         TripType   `json:"trip_type"`
	VibeTags         []string   `json:"vibe_tags"`
	TravelersPlanned int        `json:"travelers_planned"`
	StartDate        *time.Time `json:"start_date"`
	EndDate          *time.Time `json:"end_date"`
	Budget           float64    `json:"budget"`
	CoverImageURL    *string    `json:"cover_image_url"`
	Notes            *string    `json:"notes"`
}

type UpdateTripRequest struct {
	Title            string     `json:"title"`
	Destination      string     `json:"destination"`
	TripType         TripType   `json:"trip_type"`
	VibeTags         []string   `json:"vibe_tags"`
	TravelersPlanned int        `json:"travelers_planned"`
	StartDate        *time.Time `json:"start_date"`
	EndDate          *time.Time `json:"end_date"`
	Budget           float64    `json:"budget"`
	CoverImageURL    *string    `json:"cover_image_url"`
	Notes            *string    `json:"notes"`
	Status           TripStatus `json:"status"`
}

type InviteMemberRequest struct {
	UserID uuid.UUID `json:"user_id" binding:"required"`
	Role   Role      `json:"role"`
}

type UpdateMemberStatusRequest struct {
	JoinStatus JoinStatus `json:"join_status" binding:"required"`
}

type CreateItineraryItemRequest struct {
	DayNumber    int        `json:"day_number"    binding:"required"`
	Title        string     `json:"title"         binding:"required"`
	ItemType     ItemType   `json:"item_type"`
	StartTime    *time.Time `json:"start_time"`
	LocationName string     `json:"location_name"`
	Notes        *string    `json:"notes"`
	Lat          *float64   `json:"lat"`
	Lng          *float64   `json:"lng"`
	SortOrder    int        `json:"sort_order"`
}

type UpdateItineraryItemRequest struct {
	DayNumber    int        `json:"day_number"`
	Title        string     `json:"title"`
	ItemType     ItemType   `json:"item_type"`
	StartTime    *time.Time `json:"start_time"`
	LocationName string     `json:"location_name"`
	Notes        *string    `json:"notes"`
	Lat          *float64   `json:"lat"`
	Lng          *float64   `json:"lng"`
	SortOrder    int        `json:"sort_order"`
}

type CreateExpenseRequest struct {
	Description  string    `json:"description"  binding:"required"`
	Category     string    `json:"category"`
	Amount       float64   `json:"amount"       binding:"required"`
	ExpenseDate  time.Time `json:"expense_date" binding:"required"`
	CurrencyCode string    `json:"currency_code"`
}

type UpdateExpenseRequest struct {
	Description  string    `json:"description"`
	Category     string    `json:"category"`
	Amount       float64   `json:"amount"`
	ExpenseDate  time.Time `json:"expense_date"`
	CurrencyCode string    `json:"currency_code"`
}
