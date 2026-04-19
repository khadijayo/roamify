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
	RefreshLikesCount(postID uuid.UUID) error

	AddComment(comment *PostComment) error
	FindCommentsByPost(postID uuid.UUID) ([]PostComment, error)
	FindAuthorSummaries(userIDs []uuid.UUID) (map[uuid.UUID]postAuthorSummary, error)
	FindLikeCounts(postIDs []uuid.UUID) (map[uuid.UUID]int64, error)
	FindCommentCounts(postIDs []uuid.UUID) (map[uuid.UUID]int64, error)
	FindLikedPostIDs(postIDs []uuid.UUID, viewerID uuid.UUID) (map[uuid.UUID]bool, error)
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

func (r *repository) RefreshLikesCount(postID uuid.UUID) error {
	return r.db.Model(&Post{}).
		Where("id = ?", postID).
		Update("likes_count", gorm.Expr("(SELECT COUNT(*) FROM post_likes WHERE post_id = ?)", postID)).
		Error
}

func (r *repository) FindFeedForUser(viewerID uuid.UUID, limit, offset int) ([]Post, int64, error) {
	var posts []Post
	var count int64

	r.db.Model(&Post{}).
		Where(
			"visibility = ? OR (visibility = ? AND author_user_id IN (SELECT following_id FROM user_follows WHERE follower_id = ?))",
			VisibilityPublic, VisibilityFollowers, viewerID,
		).
		Count(&count)

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

func (r *repository) AddComment(comment *PostComment) error {
	return r.db.Create(comment).Error
}

func (r *repository) FindCommentsByPost(postID uuid.UUID) ([]PostComment, error) {
	var comments []PostComment
	err := r.db.
		Where("post_id = ?", postID).
		Order("created_at ASC").
		Find(&comments).Error
	return comments, err
}

func (r *repository) FindAuthorSummaries(userIDs []uuid.UUID) (map[uuid.UUID]postAuthorSummary, error) {
	summaries := make(map[uuid.UUID]postAuthorSummary)
	if len(userIDs) == 0 {
		return summaries, nil
	}

	var rows []struct {
		ID        uuid.UUID `gorm:"column:id"`
		FullName  string    `gorm:"column:full_name"`
		AvatarURL *string   `gorm:"column:avatar_url"`
	}

	if err := r.db.Table("users").
		Select("id, full_name, avatar_url").
		Where("id IN ?", userIDs).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	for _, row := range rows {
		summaries[row.ID] = postAuthorSummary{
			UserID:    row.ID,
			FullName:  row.FullName,
			AvatarURL: row.AvatarURL,
		}
	}

	return summaries, nil
}

func (r *repository) FindLikeCounts(postIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
	return r.aggregateCounts("post_likes", postIDs)
}

func (r *repository) FindCommentCounts(postIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
	return r.aggregateCounts("post_comments", postIDs)
}

func (r *repository) FindLikedPostIDs(postIDs []uuid.UUID, viewerID uuid.UUID) (map[uuid.UUID]bool, error) {
	liked := make(map[uuid.UUID]bool)
	if len(postIDs) == 0 || viewerID == uuid.Nil {
		return liked, nil
	}

	var rows []struct {
		PostID uuid.UUID `gorm:"column:post_id"`
	}

	if err := r.db.Table("post_likes").
		Select("post_id").
		Where("user_id = ? AND post_id IN ?", viewerID, postIDs).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	for _, row := range rows {
		liked[row.PostID] = true
	}

	return liked, nil
}

func (r *repository) aggregateCounts(table string, postIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
	counts := make(map[uuid.UUID]int64)
	if len(postIDs) == 0 {
		return counts, nil
	}

	var rows []struct {
		PostID uuid.UUID `gorm:"column:post_id"`
		Count  int64     `gorm:"column:count"`
	}

	if err := r.db.Table(table).
		Select("post_id, COUNT(*) AS count").
		Where("post_id IN ?", postIDs).
		Group("post_id").
		Find(&rows).Error; err != nil {
		return nil, err
	}

	for _, row := range rows {
		counts[row.PostID] = row.Count
	}

	return counts, nil
}
