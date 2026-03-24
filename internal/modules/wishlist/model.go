package wishlist

import (
	"time"

	"github.com/google/uuid"
)

type WishlistCategory string

const (
	CategoryRestaurant  WishlistCategory = "restaurant"
	CategoryHotel       WishlistCategory = "hotel"
	CategoryAttraction  WishlistCategory = "attraction"
	CategoryCafe        WishlistCategory = "cafe"
	CategoryNightlife   WishlistCategory = "nightlife"
	CategoryNature      WishlistCategory = "nature"
	CategoryMuseum      WishlistCategory = "museum"
	CategoryShopping    WishlistCategory = "shopping"
)

type WishlistItem struct {
	ID            uuid.UUID        `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID        uuid.UUID        `gorm:"type:uuid;not null;index"                       json:"user_id"`
	Name          string           `gorm:"type:varchar(255);not null"                     json:"name"`
	Location      string           `gorm:"type:varchar(255)"                              json:"location"`
	Category      WishlistCategory `gorm:"type:varchar(30)"                               json:"category"`
	Notes         *string          `gorm:"type:text"                                      json:"notes"`
	EstimatedCost *float64         `gorm:"type:numeric(12,2)"                             json:"estimated_cost"`
	CreatedAt     time.Time        `                                                      json:"created_at"`
	UpdatedAt     time.Time        `                                                      json:"updated_at"`
}

type WishlistCollection struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"  json:"id"`
	UserID      uuid.UUID `gorm:"type:uuid;not null;index"                        json:"user_id"`
	Name        string    `gorm:"type:varchar(255);not null"                      json:"name"`
	Description *string   `gorm:"type:text"                                       json:"description"`
	Emoji       string    `gorm:"type:varchar(10)"                                json:"emoji"`
	ColorToken  string    `gorm:"type:varchar(50)"                                json:"color_token"`
	CreatedAt   time.Time `                                                       json:"created_at"`
	UpdatedAt   time.Time `                                                       json:"updated_at"`

	Items []WishlistItem `gorm:"many2many:wishlist_collection_items;joinForeignKey:CollectionID;joinReferences:WishlistItemID" json:"items,omitempty"`
}

type WishlistCollectionItem struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CollectionID   uuid.UUID `gorm:"type:uuid;not null;index"                       json:"collection_id"`
	WishlistItemID uuid.UUID `gorm:"type:uuid;not null;index"                       json:"wishlist_item_id"`
	CreatedAt      time.Time `                                                      json:"created_at"`
}

func (WishlistCollectionItem) TableName() string { return "wishlist_collection_items" }

// ---------------------------------------------------------------------------
// DTOs
// ---------------------------------------------------------------------------

type CreateItemRequest struct {
	Name          string           `json:"name"           binding:"required"`
	Location      string           `json:"location"`
	Category      WishlistCategory `json:"category"`
	Notes         *string          `json:"notes"`
	EstimatedCost *float64         `json:"estimated_cost"`
}

type UpdateItemRequest struct {
	Name          string           `json:"name"`
	Location      string           `json:"location"`
	Category      WishlistCategory `json:"category"`
	Notes         *string          `json:"notes"`
	EstimatedCost *float64         `json:"estimated_cost"`
}

type CreateCollectionRequest struct {
	Name        string  `json:"name"        binding:"required"`
	Description *string `json:"description"`
	Emoji       string  `json:"emoji"`
	ColorToken  string  `json:"color_token"`
}

type UpdateCollectionRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Emoji       string  `json:"emoji"`
	ColorToken  string  `json:"color_token"`
}

type AddToCollectionRequest struct {
	WishlistItemID uuid.UUID `json:"wishlist_item_id" binding:"required"`
}