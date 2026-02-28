package models

import "time"


type User struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Name      string    `json:"name" gorm:"not null"`
	Email     string    `json:"email" gorm:"uniqueIndex;not null"`
	Password  string    `json:"-" gorm:"not null"`
	Role      UserRole  `json:"role" gorm:"type:varchar(20);default:'user'"`
	CreatedAt time.Time `json:"created_at"`
}

type UserRole string

const (
	RoleAdmin  UserRole = "admin"
	RoleUser   UserRole = "user"
	RoleAgency UserRole = "agency"
)



type City struct {
	ID          uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	Name        string `json:"name" gorm:"not null"`
	Wilaya      string `json:"wilaya" gorm:"not null"`
	Description string `json:"description"`
	CoverImage  string `json:"cover_image"`
}



type Category struct {
	ID   uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	Name string `json:"name" gorm:"uniqueIndex;not null"`
}


type Place struct {
	ID          uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description"`
	CityID      uint      `json:"city_id" gorm:"not null"`
	CategoryID  uint      `json:"category_id" gorm:"not null"`
	Address     string    `json:"address"`
	Latitude    float64   `json:"latitude"`
	Longitude   float64   `json:"longitude"`
	PriceRange  string    `json:"price_range" gorm:"type:varchar(50)"`
	Phone       string    `json:"phone" gorm:"type:varchar(20)"`
	Website     string    `json:"website"`
	CreatedBy   uint      `json:"created_by" gorm:"not null"`
	IsVerified  bool      `json:"is_verified" gorm:"default:false"`
	CreatedAt   time.Time `json:"created_at"`

	// Associations
	City     City         `json:"city,omitempty" gorm:"foreignKey:CityID"`
	Category Category     `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
	Images   []PlaceImage `json:"images,omitempty" gorm:"foreignKey:PlaceID"`
	Reviews  []Review     `json:"reviews,omitempty" gorm:"foreignKey:PlaceID"`
}

type PlaceImage struct {
	ID       uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	PlaceID  uint   `json:"place_id" gorm:"not null;index"`
	ImageURL string `json:"image_url" gorm:"not null"`
}



type Review struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID    uint      `json:"user_id" gorm:"not null;index"`
	PlaceID   uint      `json:"place_id" gorm:"not null;index"`
	Rating    int       `json:"rating" gorm:"check:rating >= 1 AND rating <= 5"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`

	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

type Favorite struct {
	ID      uint `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID  uint `json:"user_id" gorm:"not null;uniqueIndex:idx_user_place"`
	PlaceID uint `json:"place_id" gorm:"not null;uniqueIndex:idx_user_place"`

	Place Place `json:"place,omitempty" gorm:"foreignKey:PlaceID"`
}


type Trip struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID    uint      `json:"user_id" gorm:"not null;index"`
	Name      string    `json:"name" gorm:"not null"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`

	Places []TripPlace `json:"places,omitempty" gorm:"foreignKey:TripID"`
}

type TripPlace struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	TripID    uint      `json:"trip_id" gorm:"not null;index"`
	PlaceID   uint      `json:"place_id" gorm:"not null"`
	VisitDate time.Time `json:"visit_date"`

	Place Place `json:"place,omitempty" gorm:"foreignKey:PlaceID"`
}


type SavedTrip struct {
	ID      uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID  uint      `json:"user_id" gorm:"not null;uniqueIndex:idx_saved_trip"`
	TripID  uint      `json:"trip_id" gorm:"not null;uniqueIndex:idx_saved_trip"`
	SavedAt time.Time `json:"saved_at" gorm:"autoCreateTime"`

	Trip Trip `json:"trip,omitempty" gorm:"foreignKey:TripID"`
}



type Agency struct {
	ID          uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID      uint      `json:"user_id" gorm:"uniqueIndex;not null"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description"`
	Address     string    `json:"address"`
	CityID      uint      `json:"city_id" gorm:"not null"`
	Phone       string    `json:"phone" gorm:"type:varchar(20)"`
	IsVerified  bool      `json:"is_verified" gorm:"default:false"`

	City     City            `json:"city,omitempty" gorm:"foreignKey:CityID"`
	Packages []AgencyPackage `json:"packages,omitempty" gorm:"foreignKey:AgencyID"`
}

type AgencyPackage struct {
	ID          uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	AgencyID    uint      `json:"agency_id" gorm:"not null;index"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Duration    int       `json:"duration"`
	MaxPeople   int       `json:"max_people"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	IsAvailable bool      `json:"is_available" gorm:"default:true"`
	CreatedAt   time.Time `json:"created_at"`

	Agency Agency       `json:"agency,omitempty" gorm:"foreignKey:AgencyID"`
	
}



type TravelGroup struct {
	ID          uint              `json:"id" gorm:"primaryKey;autoIncrement"`
	Name        string            `json:"name" gorm:"not null"`
	Description string            `json:"description"`
	CoverImage  string            `json:"cover_image"`
	CityID      *uint             `json:"city_id"`
	TripID      *uint             `json:"trip_id"`
	CreatedBy   uint              `json:"created_by" gorm:"not null"`
	MaxMembers  int               `json:"max_members" gorm:"default:0"`
	Status      TravelGroupStatus `json:"status" gorm:"type:varchar(20);default:'open'"`
	CreatedAt   time.Time         `json:"created_at"`

	Members  []GroupMember  `json:"members,omitempty" gorm:"foreignKey:GroupID"`
	Messages []GroupMessage `json:"messages,omitempty" gorm:"foreignKey:GroupID"`
	City     *City          `json:"city,omitempty" gorm:"foreignKey:CityID"`
}

type TravelGroupStatus string

const (
	GroupStatusOpen   TravelGroupStatus = "open"
	GroupStatusClosed TravelGroupStatus = "closed"
	GroupStatusFull   TravelGroupStatus = "full"
)

type GroupMember struct {
	ID       uint            `json:"id" gorm:"primaryKey;autoIncrement"`
	GroupID  uint            `json:"group_id" gorm:"not null;uniqueIndex:idx_group_user"`
	UserID   uint            `json:"user_id" gorm:"not null;uniqueIndex:idx_group_user"`
	Role     GroupMemberRole `json:"role" gorm:"type:varchar(20);default:'member'"`
	JoinedAt time.Time       `json:"joined_at" gorm:"autoCreateTime"`

	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

type GroupMemberRole string

const (
	GroupRoleAdmin  GroupMemberRole = "admin"
	GroupRoleMember GroupMemberRole = "member"
)

type GroupMessage struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	GroupID   uint      `json:"group_id" gorm:"not null;index"`
	UserID    uint      `json:"user_id" gorm:"not null"`
	Content   string    `json:"content" gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`

	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}
