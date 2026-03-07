package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	FullName     string     `gorm:"not null" json:"full_name"`
	Email        string     `gorm:"not null;unique" json:"email"`
	PasswordHash string     `gorm:"not null" json:"-"`
	AvatarURL    string     `json:"avatar_url"`
	IsActive     bool       `gorm:"default:true" json:"is_active"`
	LastLoginAt  *time.Time `json:"last_login_at"`

	Reviews   []Review   `gorm:"foreignKey:UserID" json:"reviews,omitempty"`
	Bookmarks []Bookmark `gorm:"foreignKey:UserID" json:"bookmarks,omitempty"`
}

type DestinationType string

const (
	TypeResort      DestinationType = "resort"
	TypeDesert      DestinationType = "desert"
	TypeMountain    DestinationType = "mountain"
	TypeCulturalSite DestinationType = "cultural_site"
	TypeBeach       DestinationType = "beach"
	TypeMuseum      DestinationType = "museum"
	TypeHistorical  DestinationType = "historical"
)

type PlaceCategory string

const (
	CategoryDestination PlaceCategory = "destination" // natural/cultural spots
	CategoryHotel       PlaceCategory = "hotel"
	CategoryRestaurant  PlaceCategory = "restaurant"
	CategoryTravelAgency PlaceCategory = "travel_agency"
)

type Season string

const (
	SeasonSpring Season = "spring"
	SeasonSummer Season = "summer"
	SeasonAutumn Season = "autumn"
	SeasonWinter Season = "winter"
)

type ActivityType string

const (
	ActivityHiking       ActivityType = "hiking"
	ActivitySwimming     ActivityType = "swimming"
	ActivityCamping      ActivityType = "camping"
	ActivityFoodExperience ActivityType = "food_experience"
	ActivityAdventure    ActivityType = "adventure"
	ActivityRelaxation   ActivityType = "relaxation"
	ActivityCultural     ActivityType = "cultural"
	ActivityShopping     ActivityType = "shopping"
)



type Wilaya struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Name      string  `gorm:"not null;unique" json:"name"`        
	ArabicName string `gorm:"not null" json:"arabic_name"`        
	Code      int     `gorm:"not null;unique" json:"code"`        
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	CoverImage string `json:"cover_image"`

	Places []Place `gorm:"foreignKey:WilayaID" json:"places,omitempty"`
}

// Place is the unified model for destinations, hotels, restaurants, agencies
// The `category` field drives which sub-model applies
type Place struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Name        string        `gorm:"not null" json:"name"`
	Slug        string        `gorm:"not null;unique" json:"slug"`
	Category    PlaceCategory `gorm:"not null;index" json:"category"`
	Description string        `gorm:"type:text" json:"description"`

	WilayaID  uint    `gorm:"not null;index" json:"wilaya_id"`
	Wilaya    Wilaya  `gorm:"foreignKey:WilayaID" json:"wilaya,omitempty"`
	Address   string  `json:"address"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	MapsURL   string  `json:"maps_url"` 

	IsOpen    *bool  `json:"is_open"`    
	IsFeatured bool  `gorm:"default:false" json:"is_featured"`
	IsActive   bool  `gorm:"default:true" json:"is_active"`

	AverageRating float32 `gorm:"default:0" json:"average_rating"` 
	ReviewCount   int     `gorm:"default:0" json:"review_count"`

	Images      []PlaceImage  `gorm:"foreignKey:PlaceID" json:"images,omitempty"`
	Reviews     []Review      `gorm:"foreignKey:PlaceID" json:"reviews,omitempty"`
	Activities  []PlaceActivity `gorm:"foreignKey:PlaceID" json:"activities,omitempty"`
	BestSeasons []PlaceSeason `gorm:"foreignKey:PlaceID" json:"best_seasons,omitempty"`

	DestinationDetail *DestinationDetail `gorm:"foreignKey:PlaceID" json:"destination_detail,omitempty"`
	HotelDetail       *HotelDetail       `gorm:"foreignKey:PlaceID" json:"hotel_detail,omitempty"`
	RestaurantDetail  *RestaurantDetail  `gorm:"foreignKey:PlaceID" json:"restaurant_detail,omitempty"`
	AgencyDetail      *AgencyDetail      `gorm:"foreignKey:PlaceID" json:"agency_detail,omitempty"`
}

type DestinationDetail struct {
	ID      uint      `gorm:"primaryKey" json:"id"`
	PlaceID uuid.UUID `gorm:"type:uuid;not null;unique" json:"place_id"`

	Type          DestinationType `gorm:"not null" json:"type"` // beach, mountain, etc.
	EntryFee      float64         `json:"entry_fee"`            
	Currency      string          `gorm:"default:'DZD'" json:"currency"`
	VisitDuration string          `json:"visit_duration"`      
	Accessibility string          `json:"accessibility"`       
	TipsText      string          `gorm:"type:text" json:"tips_text"`
}


type HotelDetail struct {
	ID      uint      `gorm:"primaryKey" json:"id"`
	PlaceID uuid.UUID `gorm:"type:uuid;not null;unique" json:"place_id"`

	StarRating    int     `json:"star_rating"`    
	PricePerNight float64 `json:"price_per_night"`
	Currency      string  `gorm:"default:'DZD'" json:"currency"`
	PhoneNumber   string  `json:"phone_number"`
	Email         string  `json:"email"`
	Website       string  `json:"website"`
	CheckIn       string  `json:"check_in"`  
	CheckOut      string  `json:"check_out"` 
	HasPool       bool    `json:"has_pool"`
	HasWifi       bool    `json:"has_wifi"`
	HasParking    bool    `json:"has_parking"`
	HasRestaurant bool    `json:"has_restaurant"`
}

type RestaurantDetail struct {
	ID      uint      `gorm:"primaryKey" json:"id"`
	PlaceID uuid.UUID `gorm:"type:uuid;not null;unique" json:"place_id"`

	Cuisine       string  `json:"cuisine"`          
	PriceRange    string  `json:"price_range"`      
	PhoneNumber   string  `json:"phone_number"`
	OpeningHours  string  `json:"opening_hours"`   
	HasDelivery   bool    `json:"has_delivery"`
	HasTakeaway   bool    `json:"has_takeaway"`
	Halal         bool    `gorm:"default:true" json:"halal"`
}


type AgencyDetail struct {
	ID      uint      `gorm:"primaryKey" json:"id"`
	PlaceID uuid.UUID `gorm:"type:uuid;not null;unique" json:"place_id"`

	LicenseNumber string `json:"license_number"`
	PhoneNumber   string `json:"phone_number"`
	Email         string `json:"email"`
	Website       string `json:"website"`
	OpeningHours  string `json:"opening_hours"`
	ServicesOffered string `gorm:"type:text" json:"services_offered"` // comma-separated or JSON
}

type PlaceImage struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	PlaceID   uuid.UUID `gorm:"type:uuid;not null;index" json:"place_id"`
	URL       string    `gorm:"not null" json:"url"`
	AltText   string    `json:"alt_text"`
	IsCover   bool      `gorm:"default:false" json:"is_cover"` // main card image
	SortOrder int       `gorm:"default:0" json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
}


type PlaceActivity struct {
	ID       uint         `gorm:"primaryKey" json:"id"`
	PlaceID  uuid.UUID    `gorm:"type:uuid;not null;index" json:"place_id"`
	Activity ActivityType `gorm:"not null" json:"activity"`
}


type PlaceSeason struct {
	ID      uint      `gorm:"primaryKey" json:"id"`
	PlaceID uuid.UUID `gorm:"type:uuid;not null;index" json:"place_id"`
	Season  Season    `gorm:"not null" json:"season"`
}


type Review struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	PlaceID uuid.UUID `gorm:"type:uuid;not null;index" json:"place_id"`
	UserID  uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	User    User      `gorm:"foreignKey:UserID" json:"user,omitempty"`

	Rating  int    `gorm:"not null;check:rating >= 1 AND rating <= 5" json:"rating"`
	Title   string `json:"title"`
	Body    string `gorm:"type:text" json:"body"`
	IsVisible bool `gorm:"default:true" json:"is_visible"`
}


type Bookmark struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`

	UserID  uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	PlaceID uuid.UUID `gorm:"type:uuid;not null;index" json:"place_id"`
	Place   Place     `gorm:"foreignKey:PlaceID" json:"place,omitempty"`
}


type TravelTip struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Title     string    `gorm:"not null" json:"title"`
	Slug      string    `gorm:"not null;unique" json:"slug"`
	Body      string    `gorm:"type:text;not null" json:"body"`
	CoverImage string   `json:"cover_image"`
	WilayaID  *uint     `json:"wilaya_id"` 
	IsPublished bool    `gorm:"default:false" json:"is_published"`
	AuthorID  uuid.UUID `gorm:"type:uuid" json:"author_id"`
}