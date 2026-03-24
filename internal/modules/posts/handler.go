package posts

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/khadijayo/roamify/pkg/middleware"
	"github.com/khadijayo/roamify/pkg/response"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// POST /posts
func (h *Handler) CreatePost(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	post, err := h.svc.CreatePost(userID, &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Created(c, "post created", post)
}

// GET /posts  (feed)
func (h *Handler) GetFeed(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	posts, meta, err := h.svc.GetFeed(page, pageSize)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OKPaginated(c, "feed fetched", posts, meta)
}

// GET /posts/:postId
func (h *Handler) GetPost(c *gin.Context) {
	postID, err := uuid.Parse(c.Param("postId"))
	if err != nil {
		response.BadRequest(c, "invalid post id")
		return
	}
	post, err := h.svc.GetPost(postID)
	if err != nil {
		response.NotFound(c, "post not found")
		return
	}
	response.OK(c, "post fetched", post)
}

// GET /users/:userId/posts
func (h *Handler) GetUserPosts(c *gin.Context) {
	authorID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	posts, meta, err := h.svc.GetUserPosts(authorID, page, pageSize)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OKPaginated(c, "user posts fetched", posts, meta)
}

// PATCH /posts/:postId
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
	post, err := h.svc.UpdatePost(postID, userID, &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, "post updated", post)
}

// DELETE /posts/:postId
func (h *Handler) DeletePost(c *gin.Context) {
	userID := middleware.GetUserID(c)
	postID, err := uuid.Parse(c.Param("postId"))
	if err != nil {
		response.BadRequest(c, "invalid post id")
		return
	}
	if err := h.svc.DeletePost(postID, userID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, "post deleted", nil)
}

// POST /posts/:postId/like
func (h *Handler) LikePost(c *gin.Context) {
	userID := middleware.GetUserID(c)
	postID, err := uuid.Parse(c.Param("postId"))
	if err != nil {
		response.BadRequest(c, "invalid post id")
		return
	}
	if err := h.svc.LikePost(postID, userID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, "post liked", nil)
}

// DELETE /posts/:postId/like
func (h *Handler) UnlikePost(c *gin.Context) {
	userID := middleware.GetUserID(c)
	postID, err := uuid.Parse(c.Param("postId"))
	if err != nil {
		response.BadRequest(c, "invalid post id")
		return
	}
	if err := h.svc.UnlikePost(postID, userID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, "post unliked", nil)
}