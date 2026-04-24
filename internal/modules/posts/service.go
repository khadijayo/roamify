package posts

import (
	"context"
	"errors"
	"math"
	"strings"

	"github.com/google/uuid"
	"github.com/khadijayo/roamify/internal/modules/users"
	"github.com/khadijayo/roamify/pkg/response"
	"gorm.io/gorm"
)

type Service interface {
	CreatePost(ctx context.Context, authorID uuid.UUID, req *CreatePostRequest) (*PostResponse, error)
	GetPost(ctx context.Context, id, viewerID uuid.UUID, viewerRole string) (*PostResponse, error)
	GetFeed(ctx context.Context, page, pageSize int) ([]PostResponse, *response.Meta, error)
	GetFeedForUser(ctx context.Context, viewerID uuid.UUID, page, pageSize int) ([]PostResponse, *response.Meta, error)
	GetUserPosts(ctx context.Context, authorID, viewerID uuid.UUID, viewerRole string, page, pageSize int) ([]PostResponse, *response.Meta, error)
	UpdatePost(ctx context.Context, postID, authorID uuid.UUID, req *UpdatePostRequest) (*PostResponse, error)
	DeletePost(ctx context.Context, postID, authorID uuid.UUID) error
	LikePost(ctx context.Context, postID, userID uuid.UUID) error
	UnlikePost(ctx context.Context, postID, userID uuid.UUID) error
	GetComments(ctx context.Context, postID, viewerID uuid.UUID, viewerRole string) ([]PostCommentResponse, error)
	AddComment(ctx context.Context, postID, userID uuid.UUID, userRole string, req *CreateCommentRequest) (*PostCommentResponse, error)
	DeleteComment(ctx context.Context, postID, commentID, actorID uuid.UUID, actorRole string) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreatePost(ctx context.Context, authorID uuid.UUID, req *CreatePostRequest) (*PostResponse, error) {
	vis := req.Visibility
	if vis == "" {
		vis = VisibilityPublic
	}
	content := strings.TrimSpace(req.Content)
	if content == "" {
		return nil, errors.New("content is required")
	}
	post := &Post{
		AuthorUserID: authorID,
		Content:      content,
		Location:     strings.TrimSpace(req.Location),
		ImageURL:     req.ImageURL,
		Visibility:   vis,
	}
	if err := s.repo.CreatePost(ctx, post); err != nil {
		return nil, err
	}
	if len(req.Tags) > 0 {
		tags := make([]PostTag, 0, len(req.Tags))
		for _, t := range req.Tags {
			tag := strings.TrimSpace(t)
			if tag == "" {
				continue
			}
			tags = append(tags, PostTag{PostID: post.ID, Tag: tag})
		}
		if len(tags) > 0 {
			_ = s.repo.AddTags(ctx, tags)
		}
	}
	return s.GetPost(ctx, post.ID, authorID, string(users.RoleUser))
}

func (s *service) GetPost(ctx context.Context, id, viewerID uuid.UUID, viewerRole string) (*PostResponse, error) {
	post, err := s.repo.FindPostByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if post.IsHidden && viewerRole != string(users.RoleAdmin) && post.AuthorUserID != viewerID {
		return nil, gorm.ErrRecordNotFound
	}
	return s.buildPostResponse(ctx, post, viewerID, viewerRole, true)
}

func (s *service) GetFeed(ctx context.Context, page, pageSize int) ([]PostResponse, *response.Meta, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	posts, total, err := s.repo.FindFeed(ctx, pageSize, offset)
	if err != nil {
		return nil, nil, err
	}
	items, err := s.buildPostResponses(ctx, posts, uuid.Nil, "", false)
	if err != nil {
		return nil, nil, err
	}
	meta := &response.Meta{
		Page:       page,
		PageSize:   pageSize,
		TotalItems: total,
		TotalPages: int(math.Ceil(float64(total) / float64(pageSize))),
	}
	return items, meta, nil
}

func (s *service) GetFeedForUser(ctx context.Context, viewerID uuid.UUID, page, pageSize int) ([]PostResponse, *response.Meta, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	posts, total, err := s.repo.FindFeedForUser(ctx, viewerID, pageSize, offset)
	if err != nil {
		return nil, nil, err
	}
	items, err := s.buildPostResponses(ctx, posts, viewerID, string(users.RoleUser), false)
	if err != nil {
		return nil, nil, err
	}
	meta := &response.Meta{
		Page:       page,
		PageSize:   pageSize,
		TotalItems: total,
		TotalPages: int(math.Ceil(float64(total) / float64(pageSize))),
	}
	return items, meta, nil
}

func (s *service) GetUserPosts(ctx context.Context, authorID, viewerID uuid.UUID, viewerRole string, page, pageSize int) ([]PostResponse, *response.Meta, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	includeHidden := viewerRole == string(users.RoleAdmin) || authorID == viewerID
	posts, total, err := s.repo.FindByAuthor(ctx, authorID, pageSize, offset, includeHidden)
	if err != nil {
		return nil, nil, err
	}
	items, err := s.buildPostResponses(ctx, posts, viewerID, viewerRole, false)
	if err != nil {
		return nil, nil, err
	}
	meta := &response.Meta{
		Page:       page,
		PageSize:   pageSize,
		TotalItems: total,
		TotalPages: int(math.Ceil(float64(total) / float64(pageSize))),
	}
	return items, meta, nil
}

func (s *service) UpdatePost(ctx context.Context, postID, authorID uuid.UUID, req *UpdatePostRequest) (*PostResponse, error) {
	post, err := s.repo.FindPostByID(ctx, postID)
	if err != nil {
		return nil, err
	}
	if post.AuthorUserID != authorID {
		return nil, errors.New("not authorized to edit this post")
	}
	if req.Content != "" {
		post.Content = strings.TrimSpace(req.Content)
	}
	if req.Location != "" {
		post.Location = strings.TrimSpace(req.Location)
	}
	if req.ImageURL != nil {
		post.ImageURL = req.ImageURL
	}
	if req.Visibility != "" {
		post.Visibility = req.Visibility
	}
	if err := s.repo.UpdatePost(ctx, post); err != nil {
		return nil, err
	}
	return s.GetPost(ctx, post.ID, authorID, string(users.RoleUser))
}

func (s *service) DeletePost(ctx context.Context, postID, authorID uuid.UUID) error {
	post, err := s.repo.FindPostByID(ctx, postID)
	if err != nil {
		return err
	}
	if post.AuthorUserID != authorID {
		return errors.New("not authorized to delete this post")
	}
	return s.repo.DeletePost(ctx, postID)
}

func (s *service) LikePost(ctx context.Context, postID, userID uuid.UUID) error {
	if _, err := s.repo.FindPostByID(ctx, postID); err != nil {
		return err
	}
	_, err := s.repo.FindLike(ctx, postID, userID)
	if err == nil {
		return errors.New("already liked")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	like := &PostLike{PostID: postID, UserID: userID}
	if err := s.repo.AddLike(ctx, like); err != nil {
		return err
	}
	if err := s.repo.IncrementLikes(ctx, postID); err != nil {
		_ = s.repo.RefreshLikesCount(ctx, postID)
		return err
	}
	return s.repo.RefreshLikesCount(ctx, postID)
}

func (s *service) UnlikePost(ctx context.Context, postID, userID uuid.UUID) error {
	if _, err := s.repo.FindPostByID(ctx, postID); err != nil {
		return err
	}
	if _, err := s.repo.FindLike(ctx, postID, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("post not liked yet")
		}
		return err
	}
	if err := s.repo.RemoveLike(ctx, postID, userID); err != nil {
		return err
	}
	if err := s.repo.DecrementLikes(ctx, postID); err != nil {
		_ = s.repo.RefreshLikesCount(ctx, postID)
		return err
	}
	return s.repo.RefreshLikesCount(ctx, postID)
}

func (s *service) GetComments(ctx context.Context, postID, viewerID uuid.UUID, viewerRole string) ([]PostCommentResponse, error) {
	post, err := s.repo.FindPostByID(ctx, postID)
	if err != nil {
		return nil, err
	}
	if post.IsHidden && viewerRole != string(users.RoleAdmin) && post.AuthorUserID != viewerID {
		return nil, gorm.ErrRecordNotFound
	}
	comments, err := s.repo.FindCommentsByPost(ctx, postID)
	if err != nil {
		return nil, err
	}
	return s.buildCommentResponses(ctx, comments)
}

func (s *service) AddComment(ctx context.Context, postID, userID uuid.UUID, userRole string, req *CreateCommentRequest) (*PostCommentResponse, error) {
	post, err := s.repo.FindPostByID(ctx, postID)
	if err != nil {
		return nil, err
	}
	if post.IsHidden && userRole != string(users.RoleAdmin) && post.AuthorUserID != userID {
		return nil, gorm.ErrRecordNotFound
	}

	content := strings.TrimSpace(req.Content)
	if content == "" {
		return nil, errors.New("content is required")
	}

	comment := &PostComment{
		PostID:  postID,
		UserID:  userID,
		Content: content,
	}
	if err := s.repo.AddComment(ctx, comment); err != nil {
		return nil, err
	}

	comments, err := s.buildCommentResponses(ctx, []PostComment{*comment})
	if err != nil {
		return nil, err
	}
	if len(comments) == 0 {
		return nil, errors.New("comment could not be loaded")
	}
	return &comments[0], nil
}

func (s *service) DeleteComment(ctx context.Context, postID, commentID, actorID uuid.UUID, actorRole string) error {
	if _, err := s.repo.FindPostByID(ctx, postID); err != nil {
		return err
	}

	comment, err := s.repo.FindCommentByID(ctx, commentID)
	if err != nil {
		return err
	}
	if comment.PostID != postID {
		return gorm.ErrRecordNotFound
	}
	if comment.UserID != actorID && actorRole != string(users.RoleAdmin) {
		return errors.New("forbidden")
	}

	return s.repo.DeleteComment(ctx, commentID)
}

func (s *service) buildPostResponse(ctx context.Context, post *Post, viewerID uuid.UUID, viewerRole string, includeComments bool) (*PostResponse, error) {
	posts, err := s.buildPostResponses(ctx, []Post{*post}, viewerID, viewerRole, includeComments)
	if err != nil {
		return nil, err
	}
	if len(posts) == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &posts[0], nil
}

func (s *service) buildPostResponses(ctx context.Context, posts []Post, viewerID uuid.UUID, viewerRole string, includeComments bool) ([]PostResponse, error) {
	responses := make([]PostResponse, 0, len(posts))
	if len(posts) == 0 {
		return responses, nil
	}

	postIDs := make([]uuid.UUID, 0, len(posts))
	authorIDs := make([]uuid.UUID, 0, len(posts))
	seenAuthors := make(map[uuid.UUID]bool)
	for _, post := range posts {
		postIDs = append(postIDs, post.ID)
		if !seenAuthors[post.AuthorUserID] {
			seenAuthors[post.AuthorUserID] = true
			authorIDs = append(authorIDs, post.AuthorUserID)
		}
	}

	authors, err := s.repo.FindAuthorSummaries(ctx, authorIDs)
	if err != nil {
		return nil, err
	}
	likeCounts, err := s.repo.FindLikeCounts(ctx, postIDs)
	if err != nil {
		return nil, err
	}
	commentCounts, err := s.repo.FindCommentCounts(ctx, postIDs)
	if err != nil {
		return nil, err
	}
	likedPostIDs, err := s.repo.FindLikedPostIDs(ctx, postIDs, viewerID)
	if err != nil {
		return nil, err
	}

	commentsByPost := make(map[uuid.UUID][]PostCommentResponse)
	if includeComments {
		for _, post := range posts {
			comments, err := s.GetComments(ctx, post.ID, viewerID, viewerRole)
			if err != nil {
				return nil, err
			}
			commentsByPost[post.ID] = comments
		}
	}

	for _, post := range posts {
		author := authors[post.AuthorUserID]
		item := PostResponse{
			ID:              post.ID,
			AuthorUserID:    post.AuthorUserID,
			AuthorName:      author.FullName,
			FullName:        author.FullName,
			AuthorAvatarURL: author.AvatarURL,
			Content:         post.Content,
			Location:        post.Location,
			ImageURL:        post.ImageURL,
			LikesCount:      likeCounts[post.ID],
			CommentsCount:   commentCounts[post.ID],
			IsHidden:        post.IsHidden,
			ReportsCount:    post.ReportsCount,
			IsLiked:         likedPostIDs[post.ID],
			Visibility:      post.Visibility,
			CreatedAt:       post.CreatedAt,
			UpdatedAt:       post.UpdatedAt,
			Tags:            post.Tags,
			Comments:        commentsByPost[post.ID],
		}
		responses = append(responses, item)
	}

	return responses, nil
}

func (s *service) buildCommentResponses(ctx context.Context, comments []PostComment) ([]PostCommentResponse, error) {
	responses := make([]PostCommentResponse, 0, len(comments))
	if len(comments) == 0 {
		return responses, nil
	}

	authorIDs := make([]uuid.UUID, 0, len(comments))
	seenAuthors := make(map[uuid.UUID]bool)
	for _, comment := range comments {
		if !seenAuthors[comment.UserID] {
			seenAuthors[comment.UserID] = true
			authorIDs = append(authorIDs, comment.UserID)
		}
	}

	authors, err := s.repo.FindAuthorSummaries(ctx, authorIDs)
	if err != nil {
		return nil, err
	}

	for _, comment := range comments {
		author := authors[comment.UserID]
		responses = append(responses, PostCommentResponse{
			ID:              comment.ID,
			PostID:          comment.PostID,
			UserID:          comment.UserID,
			AuthorName:      author.FullName,
			FullName:        author.FullName,
			AuthorAvatarURL: author.AvatarURL,
			Content:         comment.Content,
			CreatedAt:       comment.CreatedAt,
			UpdatedAt:       comment.UpdatedAt,
		})
	}

	return responses, nil
}
