package posts

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Visibility string

const (
	VisibilityPublic    Visibility = "public"
	VisibilityFollowers Visibility = "followers"
	VisibilityPrivate   Visibility = "private"
)

type Post struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	AuthorUserID uuid.UUID      `gorm:"type:uuid;not null;index"                       json:"author_user_id"`
	Content      string         `gorm:"type:text;not null"                             json:"content"`
	Location     string         `gorm:"type:varchar(255)"                              json:"location"`
	ImageURL     *string        `gorm:"type:text"                                      json:"image_url"`
	LikesCount   int            `gorm:"default:0"                                      json:"likes_count"`
	Visibility   Visibility     `gorm:"type:varchar(20);default:'public'"              json:"visibility"`
	CreatedAt    time.Time      `                                                      json:"created_at"`
	UpdatedAt    time.Time      `                                                      json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index"                                          json:"-"`

	// Associations
	Tags  []PostTag  `gorm:"foreignKey:PostID" json:"tags,omitempty"`
	Likes []PostLike `gorm:"foreignKey:PostID" json:"likes,omitempty"`
}

type PostTag struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PostID    uuid.UUID `gorm:"type:uuid;not null;index"                       json:"post_id"`
	Tag       string    `gorm:"type:varchar(100);not null"                     json:"tag"`
	CreatedAt time.Time `                                                      json:"created_at"`
}

func (PostTag) TableName() string { return "post_tags" }

type PostLike struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PostID    uuid.UUID `gorm:"type:uuid;not null;index"                       json:"post_id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index"                       json:"user_id"`
	CreatedAt time.Time `                                                      json:"created_at"`
}

func (PostLike) TableName() string { return "post_likes" }

// ---------------------------------------------------------------------------
// DTOs
// ---------------------------------------------------------------------------

type CreatePostRequest struct {
	Content    string     `json:"content"    binding:"required"`
	Location   string     `json:"location"`
	ImageURL   *string    `json:"image_url"`
	Tags       []string   `json:"tags"`
	Visibility Visibility `json:"visibility"`
}

type UpdatePostRequest struct {
	Content    string     `json:"content"`
	Location   string     `json:"location"`
	ImageURL   *string    `json:"image_url"`
	Visibility Visibility `json:"visibility"`
}