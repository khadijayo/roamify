package posts

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	CreatePost(ctx context.Context, post *Post) error
	FindPostByID(ctx context.Context, id uuid.UUID) (*Post, error)
	FindFeed(ctx context.Context, limit, offset int) ([]Post, int64, error)
	FindFeedForUser(ctx context.Context, viewerID uuid.UUID, limit, offset int) ([]Post, int64, error)
	FindByAuthor(ctx context.Context, authorID uuid.UUID, limit, offset int, includeHidden bool) ([]Post, int64, error)
	UpdatePost(ctx context.Context, post *Post) error
	DeletePost(ctx context.Context, id uuid.UUID) error

	AddTags(ctx context.Context, tags []PostTag) error
	DeleteTagsByPost(ctx context.Context, postID uuid.UUID) error

	AddLike(ctx context.Context, like *PostLike) error
	FindLike(ctx context.Context, postID, userID uuid.UUID) (*PostLike, error)
	RemoveLike(ctx context.Context, postID, userID uuid.UUID) error
	IncrementLikes(ctx context.Context, postID uuid.UUID) error
	DecrementLikes(ctx context.Context, postID uuid.UUID) error
	RefreshLikesCount(ctx context.Context, postID uuid.UUID) error

	AddComment(ctx context.Context, comment *PostComment) error
	FindCommentsByPost(ctx context.Context, postID uuid.UUID) ([]PostComment, error)
	FindCommentByID(ctx context.Context, commentID uuid.UUID) (*PostComment, error)
	DeleteComment(ctx context.Context, commentID uuid.UUID) error
	FindAuthorSummaries(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]postAuthorSummary, error)
	FindLikeCounts(ctx context.Context, postIDs []uuid.UUID) (map[uuid.UUID]int64, error)
	FindCommentCounts(ctx context.Context, postIDs []uuid.UUID) (map[uuid.UUID]int64, error)
	FindLikedPostIDs(ctx context.Context, postIDs []uuid.UUID, viewerID uuid.UUID) (map[uuid.UUID]bool, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreatePost(ctx context.Context, post *Post) error {
	return r.db.WithContext(ctx).Create(post).Error
}

func (r *repository) FindPostByID(ctx context.Context, id uuid.UUID) (*Post, error) {
	var post Post
	err := r.db.WithContext(ctx).Preload("Tags").First(&post, "id = ?", id).Error
	return &post, err
}

func (r *repository) FindFeed(ctx context.Context, limit, offset int) ([]Post, int64, error) {
	var posts []Post
	var count int64
	query := r.db.WithContext(ctx).Model(&Post{}).Where("visibility = ? AND is_hidden = false", VisibilityPublic)
	query.Count(&count)
	err := r.db.WithContext(ctx).Preload("Tags").
		Where("visibility = ? AND is_hidden = false", VisibilityPublic).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&posts).Error
	return posts, count, err
}

func (r *repository) FindByAuthor(ctx context.Context, authorID uuid.UUID, limit, offset int, includeHidden bool) ([]Post, int64, error) {
	var posts []Post
	var count int64
	query := r.db.WithContext(ctx).Model(&Post{}).Where("author_user_id = ?", authorID)
	findQuery := r.db.WithContext(ctx).Preload("Tags").Where("author_user_id = ?", authorID)
	if !includeHidden {
		query = query.Where("is_hidden = false")
		findQuery = findQuery.Where("is_hidden = false")
	}
	query.Count(&count)
	err := findQuery.
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&posts).Error
	return posts, count, err
}

func (r *repository) UpdatePost(ctx context.Context, post *Post) error {
	return r.db.WithContext(ctx).Save(post).Error
}

func (r *repository) DeletePost(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&Post{}, "id = ?", id).Error
}

func (r *repository) AddTags(ctx context.Context, tags []PostTag) error {
	if len(tags) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&tags).Error
}

func (r *repository) DeleteTagsByPost(ctx context.Context, postID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("post_id = ?", postID).Delete(&PostTag{}).Error
}

func (r *repository) AddLike(ctx context.Context, like *PostLike) error {
	return r.db.WithContext(ctx).Create(like).Error
}

func (r *repository) FindLike(ctx context.Context, postID, userID uuid.UUID) (*PostLike, error) {
	var like PostLike
	err := r.db.WithContext(ctx).Where("post_id = ? AND user_id = ?", postID, userID).First(&like).Error
	return &like, err
}

func (r *repository) RemoveLike(ctx context.Context, postID, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("post_id = ? AND user_id = ?", postID, userID).Delete(&PostLike{}).Error
}

func (r *repository) IncrementLikes(ctx context.Context, postID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&Post{}).Where("id = ?", postID).UpdateColumn("likes_count", gorm.Expr("likes_count + 1")).Error
}

func (r *repository) DecrementLikes(ctx context.Context, postID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&Post{}).Where("id = ?", postID).UpdateColumn("likes_count", gorm.Expr("GREATEST(likes_count - 1, 0)")).Error
}

func (r *repository) RefreshLikesCount(ctx context.Context, postID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&Post{}).
		Where("id = ?", postID).
		Update("likes_count", gorm.Expr("(SELECT COUNT(*) FROM post_likes WHERE post_id = ?)", postID)).
		Error
}

func (r *repository) FindFeedForUser(ctx context.Context, viewerID uuid.UUID, limit, offset int) ([]Post, int64, error) {
	var posts []Post
	var count int64

	query := r.db.WithContext(ctx).Model(&Post{}).
		Where(
			"(visibility = ? OR (visibility = ? AND author_user_id IN (SELECT following_id FROM user_follows WHERE follower_id = ?))) AND is_hidden = false",
			VisibilityPublic, VisibilityFollowers, viewerID,
		)
	query.Count(&count)

	err := r.db.WithContext(ctx).Preload("Tags").
		Where(
			"(visibility = ? OR (visibility = ? AND author_user_id IN (SELECT following_id FROM user_follows WHERE follower_id = ?))) AND is_hidden = false",
			VisibilityPublic, VisibilityFollowers, viewerID,
		).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&posts).Error

	return posts, count, err
}

func (r *repository) AddComment(ctx context.Context, comment *PostComment) error {
	return r.db.WithContext(ctx).Create(comment).Error
}

func (r *repository) FindCommentsByPost(ctx context.Context, postID uuid.UUID) ([]PostComment, error) {
	var comments []PostComment
	err := r.db.WithContext(ctx).
		Where("post_id = ?", postID).
		Order("created_at ASC").
		Find(&comments).Error
	return comments, err
}

func (r *repository) FindCommentByID(ctx context.Context, commentID uuid.UUID) (*PostComment, error) {
	var comment PostComment
	err := r.db.WithContext(ctx).First(&comment, "id = ?", commentID).Error
	return &comment, err
}

func (r *repository) DeleteComment(ctx context.Context, commentID uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&PostComment{}, "id = ?", commentID).Error
}

func (r *repository) FindAuthorSummaries(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]postAuthorSummary, error) {
	summaries := make(map[uuid.UUID]postAuthorSummary)
	if len(userIDs) == 0 {
		return summaries, nil
	}

	var rows []struct {
		ID        uuid.UUID `gorm:"column:id"`
		FullName  string    `gorm:"column:full_name"`
		AvatarURL *string   `gorm:"column:avatar_url"`
	}

	if err := r.db.WithContext(ctx).Table("users").
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

func (r *repository) FindLikeCounts(ctx context.Context, postIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
	return r.aggregateLikeCounts(ctx, postIDs)
}

func (r *repository) FindCommentCounts(ctx context.Context, postIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
	return r.aggregateCommentCounts(ctx, postIDs)
}

func (r *repository) FindLikedPostIDs(ctx context.Context, postIDs []uuid.UUID, viewerID uuid.UUID) (map[uuid.UUID]bool, error) {
	liked := make(map[uuid.UUID]bool)
	if len(postIDs) == 0 || viewerID == uuid.Nil {
		return liked, nil
	}

	var rows []struct {
		PostID uuid.UUID `gorm:"column:post_id"`
	}

	if err := r.db.WithContext(ctx).Table("post_likes").
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

func (r *repository) aggregateLikeCounts(ctx context.Context, postIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
	counts := make(map[uuid.UUID]int64)
	if len(postIDs) == 0 {
		return counts, nil
	}

	var rows []struct {
		PostID uuid.UUID `gorm:"column:post_id"`
		Count  int64     `gorm:"column:count"`
	}

	if err := r.db.WithContext(ctx).Table("post_likes").
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

func (r *repository) aggregateCommentCounts(ctx context.Context, postIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
	counts := make(map[uuid.UUID]int64)
	if len(postIDs) == 0 {
		return counts, nil
	}

	var rows []struct {
		PostID uuid.UUID `gorm:"column:post_id"`
		Count  int64     `gorm:"column:count"`
	}

	if err := r.db.WithContext(ctx).Table("post_comments").
		Select("post_id, COUNT(*) AS count").
		Where("post_id IN ? AND deleted_at IS NULL", postIDs).
		Group("post_id").
		Find(&rows).Error; err != nil {
		return nil, err
	}

	for _, row := range rows {
		counts[row.PostID] = row.Count
	}

	return counts, nil
}
