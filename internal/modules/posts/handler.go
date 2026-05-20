package posts

import (
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/khadijayo/roamify/internal/services"
	"github.com/khadijayo/roamify/pkg/middleware"
	"github.com/khadijayo/roamify/pkg/response"
	"gorm.io/gorm"
)

type Handler struct {
	svc   Service
	cloud *services.CloudinaryService
}

func NewHandler(svc Service, cloud *services.CloudinaryService) *Handler {
	return &Handler{svc: svc, cloud: cloud}
}

func (h *Handler) CreatePost(c *gin.Context) {
	userID := middleware.GetUserID(c)
	req, err := h.bindCreatePostRequest(c)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	post, err := h.svc.CreatePost(c.Request.Context(), userID, req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Created(c, "post created", post)
}

func (h *Handler) GetFeedV2(c *gin.Context) {
	viewerID := middleware.GetUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	posts, meta, err := h.svc.GetFeedForUser(c.Request.Context(), viewerID, page, pageSize)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OKPaginated(c, "feed fetched", posts, meta)
}

func (h *Handler) GetPost(c *gin.Context) {
	viewerID := middleware.GetUserID(c)
	viewerRole := middleware.GetUserRole(c)
	postID, err := uuid.Parse(c.Param("postId"))
	if err != nil {
		response.BadRequest(c, "invalid post id")
		return
	}
	post, err := h.svc.GetPost(c.Request.Context(), postID, viewerID, viewerRole)
	if err != nil {
		response.NotFound(c, "post not found")
		return
	}
	response.OK(c, "post fetched", post)
}

func (h *Handler) GetUserPosts(c *gin.Context) {
	viewerID := middleware.GetUserID(c)
	viewerRole := middleware.GetUserRole(c)
	authorID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	posts, meta, err := h.svc.GetUserPosts(c.Request.Context(), authorID, viewerID, viewerRole, page, pageSize)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OKPaginated(c, "user posts fetched", posts, meta)
}

func (h *Handler) UpdatePost(c *gin.Context) {
	userID := middleware.GetUserID(c)
	postID, err := uuid.Parse(c.Param("postId"))
	if err != nil {
		response.BadRequest(c, "invalid post id")
		return
	}
	var req UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	post, err := h.svc.UpdatePost(c.Request.Context(), postID, userID, &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, "post updated", post)
}

func (h *Handler) DeletePost(c *gin.Context) {
	userID := middleware.GetUserID(c)
	postID, err := uuid.Parse(c.Param("postId"))
	if err != nil {
		response.BadRequest(c, "invalid post id")
		return
	}
	if err := h.svc.DeletePost(c.Request.Context(), postID, userID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, "post deleted", nil)
}

func (h *Handler) LikePost(c *gin.Context) {
	userID := middleware.GetUserID(c)
	postID, err := uuid.Parse(c.Param("postId"))
	if err != nil {
		response.BadRequest(c, "invalid post id")
		return
	}
	if err := h.svc.LikePost(c.Request.Context(), postID, userID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, "post liked", nil)
}

func (h *Handler) UnlikePost(c *gin.Context) {
	userID := middleware.GetUserID(c)
	postID, err := uuid.Parse(c.Param("postId"))
	if err != nil {
		response.BadRequest(c, "invalid post id")
		return
	}
	if err := h.svc.UnlikePost(c.Request.Context(), postID, userID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, "post unliked", nil)
}

func (h *Handler) GetComments(c *gin.Context) {
	viewerID := middleware.GetUserID(c)
	viewerRole := middleware.GetUserRole(c)
	postID, err := uuid.Parse(c.Param("postId"))
	if err != nil {
		response.BadRequest(c, "invalid post id")
		return
	}
	comments, err := h.svc.GetComments(c.Request.Context(), postID, viewerID, viewerRole)
	if err != nil {
		response.NotFound(c, "post not found")
		return
	}
	response.OK(c, "comments fetched", comments)
}

func (h *Handler) AddComment(c *gin.Context) {
	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)
	postID, err := uuid.Parse(c.Param("postId"))
	if err != nil {
		response.BadRequest(c, "invalid post id")
		return
	}

	var req CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	comment, err := h.svc.AddComment(c.Request.Context(), postID, userID, userRole, &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, "comment created", comment)
}

func (h *Handler) DeleteComment(c *gin.Context) {
	postID, err := uuid.Parse(c.Param("postId"))
	if err != nil {
		response.BadRequest(c, "invalid post id")
		return
	}

	commentID, err := uuid.Parse(c.Query("comment_id"))
	if err != nil {
		response.BadRequest(c, "invalid comment_id")
		return
	}

	err = h.svc.DeleteComment(
		c.Request.Context(),
		postID,
		commentID,
		middleware.GetUserID(c),
		middleware.GetUserRole(c),
	)
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			response.NotFound(c, "comment not found")
		case err.Error() == "forbidden":
			response.Forbidden(c, err.Error())
		default:
			response.BadRequest(c, err.Error())
		}
		return
	}

	response.OK(c, "comment deleted", gin.H{"comment_id": commentID})
}

func (h *Handler) bindCreatePostRequest(c *gin.Context) (*CreatePostRequest, error) {
	if strings.HasPrefix(c.ContentType(), "multipart/form-data") {
		return h.bindMultipartCreatePostRequest(c)
	}

	var req CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, err
	}
	return &req, nil
}

func (h *Handler) bindMultipartCreatePostRequest(c *gin.Context) (*CreatePostRequest, error) {
	req := &CreatePostRequest{
		Content:    strings.TrimSpace(c.PostForm("content")),
		Location:   strings.TrimSpace(c.PostForm("location")),
		Visibility: Visibility(strings.TrimSpace(c.PostForm("visibility"))),
	}

	if req.Content == "" {
		return nil, errors.New("content is required")
	}

	if imageURL := strings.TrimSpace(c.PostForm("image_url")); imageURL != "" {
		req.ImageURL = &imageURL
	}

	tags, err := parsePostTags(c)
	if err != nil {
		return nil, err
	}
	req.Tags = tags

	fileHeader, err := firstUploadedFile(c, "image", "image_file", "file", "post_image")
	if err != nil {
		return nil, err
	}
	if fileHeader != nil {
		storedURL, err := h.uploadPostImage(c, fileHeader)
		if err != nil {
			return nil, err
		}
		req.ImageURL = &storedURL
	}

	return req, nil
}

func (h *Handler) uploadPostImage(c *gin.Context, file *multipart.FileHeader) (string, error) {
	if h.cloud == nil {
		return "", errors.New("cloudinary service is not configured")
	}

	uploadedFile, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("open uploaded file: %w", err)
	}
	defer uploadedFile.Close()

	result, err := h.cloud.Upload(c.Request.Context(), uploadedFile, file.Filename)
	if err != nil {
		return "", err
	}

	return result.SecureURL, nil
}

func parsePostTags(c *gin.Context) ([]string, error) {
	tags := c.PostFormArray("tags")
	if len(tags) == 0 {
		tags = c.PostFormArray("tags[]")
	}
	if len(tags) > 0 {
		return cleanTags(tags), nil
	}

	raw := strings.TrimSpace(c.PostForm("tags"))
	if raw == "" {
		return nil, nil
	}

	if strings.HasPrefix(raw, "[") {
		var parsed []string
		if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
			return nil, errors.New("invalid tags format")
		}
		return cleanTags(parsed), nil
	}

	return cleanTags(strings.Split(raw, ",")), nil
}

func cleanTags(tags []string) []string {
	cleaned := make([]string, 0, len(tags))
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		cleaned = append(cleaned, tag)
	}
	return cleaned
}

func firstUploadedFile(c *gin.Context, fieldNames ...string) (*multipart.FileHeader, error) {
	for _, fieldName := range fieldNames {
		file, err := c.FormFile(fieldName)
		if err == nil {
			return file, nil
		}
		if !errors.Is(err, http.ErrMissingFile) {
			return nil, err
		}
	}
	return nil, nil
}
