package models

import "time"


type User struct {
	ID        uint      `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Email     string    `json:"email" db:"email"`
	Password  string    `json:"-" db:"password"`
	Role      UserRole  `json:"role" db:"role"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type UserRole string

const (
	RoleAdmin  UserRole = "admin"
	RoleUser   UserRole = "user"
	RoleAgency UserRole = "agency"
)

type City struct {
	ID          uint   `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`
	Wilaya      string `json:"wilaya" db:"wilaya"`
	Description string `json:"description" db:"description"`
	CoverImage  string `json:"cover_image" db:"cover_image"`
}


type Category struct {
	ID   uint   `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
}


type Place struct {
	ID          uint      `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	CityID      uint      `json:"city_id" db:"city_id"`
	CategoryID  uint      `json:"category_id" db:"category_id"`
	Address     string    `json:"address" db:"address"`
	Latitude    float64   `json:"latitude" db:"latitude"`
	Longitude   float64   `json:"longitude" db:"longitude"`
	PriceRange  string    `json:"price_range" db:"price_range"`
	Phone       string    `json:"phone" db:"phone"`
	Website     string    `json:"website" db:"website"`
	CreatedBy   uint      `json:"created_by" db:"created_by"`
	IsVerified  bool      `json:"is_verified" db:"is_verified"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`


	City     *City     `json:"city,omitempty" db:"-"`
	Category *Category `json:"category,omitempty" db:"-"`
	Images   []PlaceImage `json:"images,omitempty" db:"-"`
	Reviews  []Review     `json:"reviews,omitempty" db:"-"`
}


type PlaceImage struct {
	ID       uint   `json:"id" db:"id"`
	PlaceID  uint   `json:"place_id" db:"place_id"`
	ImageURL string `json:"image_url" db:"image_url"`
}


type Review struct {
	ID        uint      `json:"id" db:"id"`
	UserID    uint      `json:"user_id" db:"user_id"`
	PlaceID   uint      `json:"place_id" db:"place_id"`
	Rating    int       `json:"rating" db:"rating"` // 1-5
	Comment   string    `json:"comment" db:"comment"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`

	// Associations
	User *User `json:"user,omitempty" db:"-"`
}


type Favorite struct {
	ID      uint `json:"id" db:"id"`
	UserID  uint `json:"user_id" db:"user_id"`
	PlaceID uint `json:"place_id" db:"place_id"`

	// Associations
	Place *Place `json:"place,omitempty" db:"-"`
}


type Trip struct {
	ID        uint      `json:"id" db:"id"`
	UserID    uint      `json:"user_id" db:"user_id"`
	Name      string    `json:"name" db:"name"`
	StartDate time.Time `json:"start_date" db:"start_date"`
	EndDate   time.Time `json:"end_date" db:"end_date"`

	// Associations
	Places []TripPlace `json:"places,omitempty" db:"-"`
}


type TripPlace struct {
	ID        uint      `json:"id" db:"id"`
	TripID    uint      `json:"trip_id" db:"trip_id"`
	PlaceID   uint      `json:"place_id" db:"place_id"`
	VisitDate time.Time `json:"visit_date" db:"visit_date"`

	// Associations
	Place *Place `json:"place,omitempty" db:"-"`
}

// ─────────────────────────────────────────
// GROUP CHATS
// ─────────────────────────────────────────


type TravelGroup struct {
	ID          uint            `json:"id" db:"id"`
	Name        string          `json:"name" db:"name"`
	Description string          `json:"description" db:"description"`
	CoverImage  string          `json:"cover_image" db:"cover_image"`
	CityID      *uint           `json:"city_id" db:"city_id"`         // optional: group tied to a city
	TripID      *uint           `json:"trip_id" db:"trip_id"`         // optional: group tied to a trip
	CreatedBy   uint            `json:"created_by" db:"created_by"`
	MaxMembers  int             `json:"max_members" db:"max_members"` // 0 = unlimited
	Status      TravelGroupStatus `json:"status" db:"status"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`

	// Associations
	Members  []GroupMember  `json:"members,omitempty" db:"-"`
	Messages []GroupMessage `json:"messages,omitempty" db:"-"`
	City     *City          `json:"city,omitempty" db:"-"`
}

type TravelGroupStatus string

const (
	GroupStatusOpen   TravelGroupStatus = "open"   // anyone can join
	GroupStatusClosed TravelGroupStatus = "closed" // invite only
	GroupStatusFull   TravelGroupStatus = "full"
)


type GroupMember struct {
	ID       uint            `json:"id" db:"id"`
	GroupID  uint            `json:"group_id" db:"group_id"`
	UserID   uint            `json:"user_id" db:"user_id"`
	Role     GroupMemberRole `json:"role" db:"role"`
	JoinedAt time.Time       `json:"joined_at" db:"joined_at"`

	// Associations
	User *User `json:"user,omitempty" db:"-"`
}

type GroupMemberRole string

const (
	GroupRoleAdmin  GroupMemberRole = "admin"
	GroupRoleMember GroupMemberRole = "member"
)


type GroupMessage struct {
	ID        uint      `json:"id" db:"id"`
	GroupID   uint      `json:"group_id" db:"group_id"`
	UserID    uint      `json:"user_id" db:"user_id"`
	Content   string    `json:"content" db:"content"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`

	// Associations
	User *User `json:"user,omitempty" db:"-"`
}

// ─────────────────────────────────────────
// COMMENT SECTIONS
// ─────────────────────────────────────────


type CommentTargetType string

const (
	CommentTargetPlace  CommentTargetType = "place"
	CommentTargetTrip   CommentTargetType = "trip"
	CommentTargetGroup  CommentTargetType = "group"
)

// Comment represents a comment on a place, trip, or group
type Comment struct {
	ID         uint              `json:"id" db:"id"`
	UserID     uint              `json:"user_id" db:"user_id"`
	TargetType CommentTargetType `json:"target_type" db:"target_type"`
	TargetID   uint              `json:"target_id" db:"target_id"`
	ParentID   *uint             `json:"parent_id" db:"parent_id"` // nil = top-level, set = reply
	Content    string            `json:"content" db:"content"`
	CreatedAt  time.Time         `json:"created_at" db:"created_at"`

	// Associations
	User    *User      `json:"user,omitempty" db:"-"`
	Replies []Comment  `json:"replies,omitempty" db:"-"`
}