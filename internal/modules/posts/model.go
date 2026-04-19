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

	Tags     []PostTag     `gorm:"foreignKey:PostID" json:"tags,omitempty"`
	Likes    []PostLike    `gorm:"foreignKey:PostID" json:"likes,omitempty"`
	Comments []PostComment `gorm:"foreignKey:PostID" json:"comments,omitempty"`
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

type PostComment struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PostID    uuid.UUID      `gorm:"type:uuid;not null;index"                       json:"post_id"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null;index"                       json:"user_id"`
	Content   string         `gorm:"type:text;not null"                             json:"content"`
	CreatedAt time.Time      `                                                      json:"created_at"`
	UpdatedAt time.Time      `                                                      json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index"                                          json:"-"`
}

func (PostComment) TableName() string { return "post_comments" }

type CreatePostRequest struct {
	Content    string     `json:"content" form:"content" binding:"required"`
	Location   string     `json:"location" form:"location"`
	ImageURL   *string    `json:"image_url" form:"image_url"`
	Tags       []string   `json:"tags" form:"tags"`
	Visibility Visibility `json:"visibility" form:"visibility"`
}

type UpdatePostRequest struct {
	Content    string     `json:"content"`
	Location   string     `json:"location"`
	ImageURL   *string    `json:"image_url"`
	Visibility Visibility `json:"visibility"`
}

type CreateCommentRequest struct {
	Content string `json:"content" binding:"required"`
}

type PostCommentResponse struct {
	ID              uuid.UUID `json:"id"`
	PostID          uuid.UUID `json:"post_id"`
	UserID          uuid.UUID `json:"user_id"`
	AuthorName      string    `json:"author_name"`
	FullName        string    `json:"full_name"`
	AuthorAvatarURL *string   `json:"author_avatar_url,omitempty"`
	Content         string    `json:"content"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type PostResponse struct {
	ID              uuid.UUID             `json:"id"`
	AuthorUserID    uuid.UUID             `json:"author_user_id"`
	AuthorName      string                `json:"author_name"`
	FullName        string                `json:"full_name"`
	AuthorAvatarURL *string               `json:"author_avatar_url,omitempty"`
	Content         string                `json:"content"`
	Location        string                `json:"location"`
	ImageURL        *string               `json:"image_url"`
	LikesCount      int64                 `json:"likes_count"`
	CommentsCount   int64                 `json:"comments_count"`
	IsLiked         bool                  `json:"is_liked"`
	Visibility      Visibility            `json:"visibility"`
	CreatedAt       time.Time             `json:"created_at"`
	UpdatedAt       time.Time             `json:"updated_at"`
	Tags            []PostTag             `json:"tags,omitempty"`
	Comments        []PostCommentResponse `json:"comments,omitempty"`
}

type postAuthorSummary struct {
	UserID    uuid.UUID
	FullName  string
	AvatarURL *string
}
