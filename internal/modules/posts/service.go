package posts

import (
	"errors"
	"math"

	"github.com/google/uuid"
	"github.com/khadijayo/roamify/pkg/response"
	"gorm.io/gorm"
)

type Service interface {
	CreatePost(authorID uuid.UUID, req *CreatePostRequest) (*Post, error)
	GetPost(id uuid.UUID) (*Post, error)
	GetFeed(page, pageSize int) ([]Post, *response.Meta, error)
	GetUserPosts(authorID uuid.UUID, page, pageSize int) ([]Post, *response.Meta, error)
	UpdatePost(postID, authorID uuid.UUID, req *UpdatePostRequest) (*Post, error)
	DeletePost(postID, authorID uuid.UUID) error
	LikePost(postID, userID uuid.UUID) error
	UnlikePost(postID, userID uuid.UUID) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreatePost(authorID uuid.UUID, req *CreatePostRequest) (*Post, error) {
	vis := req.Visibility
	if vis == "" {
		vis = VisibilityPublic
	}
	post := &Post{
		AuthorUserID: authorID,
		Content:      req.Content,
		Location:     req.Location,
		ImageURL:     req.ImageURL,
		Visibility:   vis,
	}
	if err := s.repo.CreatePost(post); err != nil {
		return nil, err
	}
	if len(req.Tags) > 0 {
		tags := make([]PostTag, 0, len(req.Tags))
		for _, t := range req.Tags {
			tags = append(tags, PostTag{PostID: post.ID, Tag: t})
		}
		_ = s.repo.AddTags(tags)
	}
	return s.repo.FindPostByID(post.ID)
}

func (s *service) GetPost(id uuid.UUID) (*Post, error) {
	return s.repo.FindPostByID(id)
}

func (s *service) GetFeed(page, pageSize int) ([]Post, *response.Meta, error) {
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
	meta := &response.Meta{
		Page:       page,
		PageSize:   pageSize,
		TotalItems: total,
		TotalPages: int(math.Ceil(float64(total) / float64(pageSize))),
	}
	return posts, meta, nil
}

func (s *service) GetUserPosts(authorID uuid.UUID, page, pageSize int) ([]Post, *response.Meta, error) {
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
	meta := &response.Meta{
		Page:       page,
		PageSize:   pageSize,
		TotalItems: total,
		TotalPages: int(math.Ceil(float64(total) / float64(pageSize))),
	}
	return posts, meta, nil
}

func (s *service) UpdatePost(postID, authorID uuid.UUID, req *UpdatePostRequest) (*Post, error) {
	post, err := s.repo.FindPostByID(postID)
	if err != nil {
		return nil, err
	}
	if post.AuthorUserID != authorID {
		return nil, errors.New("not authorized to edit this post")
	}
	if req.Content != "" {
		post.Content = req.Content
	}
	if req.Location != "" {
		post.Location = req.Location
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
	return post, nil
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
	return s.repo.IncrementLikes(postID)
}

func (s *service) UnlikePost(postID, userID uuid.UUID) error {
	if err := s.repo.RemoveLike(postID, userID); err != nil {
		return err
	}
	return s.repo.DecrementLikes(postID)
}
