package posts

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	CreatePost(post *Post) error
	FindPostByID(id uuid.UUID) (*Post, error)
	FindFeed(limit, offset int) ([]Post, int64, error)
	FindFeedForUser(viewerID uuid.UUID, limit, offset int) ([]Post, int64, error) 
	FindByAuthor(authorID uuid.UUID, limit, offset int) ([]Post, int64, error)
	UpdatePost(post *Post) error
	DeletePost(id uuid.UUID) error

	AddTags(tags []PostTag) error
	DeleteTagsByPost(postID uuid.UUID) error

	AddLike(like *PostLike) error
	FindLike(postID, userID uuid.UUID) (*PostLike, error)
	RemoveLike(postID, userID uuid.UUID) error
	IncrementLikes(postID uuid.UUID) error
	DecrementLikes(postID uuid.UUID) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreatePost(post *Post) error {
	return r.db.Create(post).Error
}

func (r *repository) FindPostByID(id uuid.UUID) (*Post, error) {
	var post Post
	err := r.db.Preload("Tags").First(&post, "id = ?", id).Error
	return &post, err
}

func (r *repository) FindFeed(limit, offset int) ([]Post, int64, error) {
	var posts []Post
	var count int64
	r.db.Model(&Post{}).Where("visibility = ?", VisibilityPublic).Count(&count)
	err := r.db.Preload("Tags").
		Where("visibility = ?", VisibilityPublic).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&posts).Error
	return posts, count, err
}

func (r *repository) FindByAuthor(authorID uuid.UUID, limit, offset int) ([]Post, int64, error) {
	var posts []Post
	var count int64
	r.db.Model(&Post{}).Where("author_user_id = ?", authorID).Count(&count)
	err := r.db.Preload("Tags").
		Where("author_user_id = ?", authorID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&posts).Error
	return posts, count, err
}

func (r *repository) UpdatePost(post *Post) error {
	return r.db.Save(post).Error
}

func (r *repository) DeletePost(id uuid.UUID) error {
	return r.db.Delete(&Post{}, "id = ?", id).Error
}

func (r *repository) AddTags(tags []PostTag) error {
	if len(tags) == 0 {
		return nil
	}
	return r.db.Create(&tags).Error
}

func (r *repository) DeleteTagsByPost(postID uuid.UUID) error {
	return r.db.Where("post_id = ?", postID).Delete(&PostTag{}).Error
}

func (r *repository) AddLike(like *PostLike) error {
	return r.db.Create(like).Error
}

func (r *repository) FindLike(postID, userID uuid.UUID) (*PostLike, error) {
	var like PostLike
	err := r.db.Where("post_id = ? AND user_id = ?", postID, userID).First(&like).Error
	return &like, err
}

func (r *repository) RemoveLike(postID, userID uuid.UUID) error {
	return r.db.Where("post_id = ? AND user_id = ?", postID, userID).Delete(&PostLike{}).Error
}

func (r *repository) IncrementLikes(postID uuid.UUID) error {
	return r.db.Model(&Post{}).Where("id = ?", postID).UpdateColumn("likes_count", gorm.Expr("likes_count + 1")).Error
}

func (r *repository) DecrementLikes(postID uuid.UUID) error {
	return r.db.Model(&Post{}).Where("id = ?", postID).UpdateColumn("likes_count", gorm.Expr("GREATEST(likes_count - 1, 0)")).Error
}

func (r *repository) FindFeedForUser(viewerID uuid.UUID, limit, offset int) ([]Post, int64, error) {
	var posts []Post
	var count int64

	// Count query
	r.db.Model(&Post{}).
		Where(
			"visibility = ? OR (visibility = ? AND author_user_id IN (SELECT following_id FROM user_follows WHERE follower_id = ?))",
			VisibilityPublic, VisibilityFollowers, viewerID,
		).
		Count(&count)

	// Fetch query
	err := r.db.Preload("Tags").
		Where(
			"visibility = ? OR (visibility = ? AND author_user_id IN (SELECT following_id FROM user_follows WHERE follower_id = ?))",
			VisibilityPublic, VisibilityFollowers, viewerID,
		).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&posts).Error

	return posts, count, err
}


