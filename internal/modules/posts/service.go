package posts

import (
	"errors"
	"math"
	"strings"

	"github.com/google/uuid"
	"github.com/khadijayo/roamify/pkg/response"
	"gorm.io/gorm"
)

type Service interface {
	CreatePost(authorID uuid.UUID, req *CreatePostRequest) (*PostResponse, error)
	GetPost(id, viewerID uuid.UUID) (*PostResponse, error)
	GetFeed(page, pageSize int) ([]PostResponse, *response.Meta, error)
	GetFeedForUser(viewerID uuid.UUID, page, pageSize int) ([]PostResponse, *response.Meta, error)
	GetUserPosts(authorID, viewerID uuid.UUID, page, pageSize int) ([]PostResponse, *response.Meta, error)
	UpdatePost(postID, authorID uuid.UUID, req *UpdatePostRequest) (*PostResponse, error)
	DeletePost(postID, authorID uuid.UUID) error
	LikePost(postID, userID uuid.UUID) error
	UnlikePost(postID, userID uuid.UUID) error
	GetComments(postID uuid.UUID) ([]PostCommentResponse, error)
	AddComment(postID, userID uuid.UUID, req *CreateCommentRequest) (*PostCommentResponse, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreatePost(authorID uuid.UUID, req *CreatePostRequest) (*PostResponse, error) {
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
	if err := s.repo.CreatePost(post); err != nil {
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
			_ = s.repo.AddTags(tags)
		}
	}
	return s.GetPost(post.ID, authorID)
}

func (s *service) GetPost(id, viewerID uuid.UUID) (*PostResponse, error) {
	post, err := s.repo.FindPostByID(id)
	if err != nil {
		return nil, err
	}
	return s.buildPostResponse(post, viewerID, true)
}

func (s *service) GetFeed(page, pageSize int) ([]PostResponse, *response.Meta, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	posts, total, err := s.repo.FindFeed(pageSize, offset)
	if err != nil {
		return nil, nil, err
	}
	items, err := s.buildPostResponses(posts, uuid.Nil, false)
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

func (s *service) GetFeedForUser(viewerID uuid.UUID, page, pageSize int) ([]PostResponse, *response.Meta, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	posts, total, err := s.repo.FindFeedForUser(viewerID, pageSize, offset)
	if err != nil {
		return nil, nil, err
	}
	items, err := s.buildPostResponses(posts, viewerID, false)
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

func (s *service) GetUserPosts(authorID, viewerID uuid.UUID, page, pageSize int) ([]PostResponse, *response.Meta, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	posts, total, err := s.repo.FindByAuthor(authorID, pageSize, offset)
	if err != nil {
		return nil, nil, err
	}
	items, err := s.buildPostResponses(posts, viewerID, false)
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

func (s *service) UpdatePost(postID, authorID uuid.UUID, req *UpdatePostRequest) (*PostResponse, error) {
	post, err := s.repo.FindPostByID(postID)
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
	if err := s.repo.UpdatePost(post); err != nil {
		return nil, err
	}
	return s.GetPost(post.ID, authorID)
}

func (s *service) DeletePost(postID, authorID uuid.UUID) error {
	post, err := s.repo.FindPostByID(postID)
	if err != nil {
		return err
	}
	if post.AuthorUserID != authorID {
		return errors.New("not authorized to delete this post")
	}
	return s.repo.DeletePost(postID)
}

func (s *service) LikePost(postID, userID uuid.UUID) error {
	if _, err := s.repo.FindPostByID(postID); err != nil {
		return err
	}
	_, err := s.repo.FindLike(postID, userID)
	if err == nil {
		return errors.New("already liked")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	like := &PostLike{PostID: postID, UserID: userID}
	if err := s.repo.AddLike(like); err != nil {
		return err
	}
	if err := s.repo.IncrementLikes(postID); err != nil {
		_ = s.repo.RefreshLikesCount(postID)
		return err
	}
	return s.repo.RefreshLikesCount(postID)
}

func (s *service) UnlikePost(postID, userID uuid.UUID) error {
	if _, err := s.repo.FindPostByID(postID); err != nil {
		return err
	}
	if _, err := s.repo.FindLike(postID, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("post not liked yet")
		}
		return err
	}
	if err := s.repo.RemoveLike(postID, userID); err != nil {
		return err
	}
	if err := s.repo.DecrementLikes(postID); err != nil {
		_ = s.repo.RefreshLikesCount(postID)
		return err
	}
	return s.repo.RefreshLikesCount(postID)
}

func (s *service) GetComments(postID uuid.UUID) ([]PostCommentResponse, error) {
	if _, err := s.repo.FindPostByID(postID); err != nil {
		return nil, err
	}
	comments, err := s.repo.FindCommentsByPost(postID)
	if err != nil {
		return nil, err
	}
	return s.buildCommentResponses(comments)
}

func (s *service) AddComment(postID, userID uuid.UUID, req *CreateCommentRequest) (*PostCommentResponse, error) {
	if _, err := s.repo.FindPostByID(postID); err != nil {
		return nil, err
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
	if err := s.repo.AddComment(comment); err != nil {
		return nil, err
	}

	comments, err := s.buildCommentResponses([]PostComment{*comment})
	if err != nil {
		return nil, err
	}
	if len(comments) == 0 {
		return nil, errors.New("comment could not be loaded")
	}
	return &comments[0], nil
}

func (s *service) buildPostResponse(post *Post, viewerID uuid.UUID, includeComments bool) (*PostResponse, error) {
	posts, err := s.buildPostResponses([]Post{*post}, viewerID, includeComments)
	if err != nil {
		return nil, err
	}
	if len(posts) == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &posts[0], nil
}

func (s *service) buildPostResponses(posts []Post, viewerID uuid.UUID, includeComments bool) ([]PostResponse, error) {
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

	authors, err := s.repo.FindAuthorSummaries(authorIDs)
	if err != nil {
		return nil, err
	}
	likeCounts, err := s.repo.FindLikeCounts(postIDs)
	if err != nil {
		return nil, err
	}
	commentCounts, err := s.repo.FindCommentCounts(postIDs)
	if err != nil {
		return nil, err
	}
	likedPostIDs, err := s.repo.FindLikedPostIDs(postIDs, viewerID)
	if err != nil {
		return nil, err
	}

	commentsByPost := make(map[uuid.UUID][]PostCommentResponse)
	if includeComments {
		for _, post := range posts {
			comments, err := s.GetComments(post.ID)
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

func (s *service) buildCommentResponses(comments []PostComment) ([]PostCommentResponse, error) {
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

	authors, err := s.repo.FindAuthorSummaries(authorIDs)
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
