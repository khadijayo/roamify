package users

import (
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

// POST /auth/register
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	res, err := h.svc.Register(&req)
	if err != nil {
		response.Conflict(c, err.Error())
		return
	}
	response.Created(c, "account created successfully", res)
}

// POST /auth/login
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	res, err := h.svc.Login(&req)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}
	response.OK(c, "login successful", res)
}

// GET /users/me
func (h *Handler) GetMe(c *gin.Context) {
	userID := middleware.GetUserID(c)
	user, err := h.svc.GetProfile(userID)
	if err != nil {
		response.NotFound(c, "user not found")
		return
	}
	response.OK(c, "profile fetched", user)
}

// PATCH /users/me
func (h *Handler) UpdateMe(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	user, err := h.svc.UpdateProfile(userID, &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, "profile updated", user)
}

// GET /users/me/vibe
func (h *Handler) GetVibeProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	vp, err := h.svc.GetVibeProfile(userID)
	if err != nil {
		response.NotFound(c, "vibe profile not found")
		return
	}
	response.OK(c, "vibe profile fetched", vp)
}

// PUT /users/me/vibe
func (h *Handler) UpsertVibeProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req UpdateVibeProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	vp, err := h.svc.UpsertVibeProfile(userID, &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, "vibe profile saved", vp)
}

// GET /users/:userId
// Returns another user's public profile — used when tapping a profile in feed/squad.
func (h *Handler) GetPublicProfile(c *gin.Context) {
	targetID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}
	user, err := h.svc.GetPublicProfile(targetID)
	if err != nil {
		response.NotFound(c, "user not found")
		return
	}
	response.OK(c, "profile fetched", user)
}

// GET /users/search?q=name
// Used by the squad invite flow — search users by name or email.
func (h *Handler) SearchUsers(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		response.BadRequest(c, "search query 'q' is required")
		return
	}
	users, err := h.svc.SearchUsers(q)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, "search results", users)
}
func (h *Handler) SocialAuth(c *gin.Context) {
    var req SocialAuthRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.BadRequest(c, err.Error())
        return
    }
    res, err := h.svc.SocialAuth(&req)
    if err != nil {
        response.BadRequest(c, err.Error())
        return
    }
    response.OK(c, "social login successful", res)
}

func (h *Handler) FollowUser(c *gin.Context) {
    followerID := middleware.GetUserID(c)
    var req FollowUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.BadRequest(c, err.Error())
        return
    }
    if err := h.svc.FollowUser(followerID, &req); err != nil {
        response.BadRequest(c, err.Error())
        return
    }
    response.OK(c, "user followed", nil)
}

func (h *Handler) UnfollowUser(c *gin.Context) {
    followerID := middleware.GetUserID(c)
    targetID, err := uuid.Parse(c.Param("userId"))
    if err != nil {
        response.BadRequest(c, "invalid user id")
        return
    }
    if err := h.svc.UnfollowUser(followerID, targetID); err != nil {
        response.BadRequest(c, err.Error())
        return
    }
    response.OK(c, "user unfollowed", nil)
}

func (h *Handler) GetFollowers(c *gin.Context) {
    targetID, err := uuid.Parse(c.Param("userId"))
    if err != nil {
        response.BadRequest(c, "invalid user id")
        return
    }
    data, err := h.svc.GetFollowers(targetID)
    if err != nil {
        response.InternalError(c, err.Error())
        return
    }
    response.OK(c, "followers fetched", data)
}

func (h *Handler) GetFollowing(c *gin.Context) {
    targetID, err := uuid.Parse(c.Param("userId"))
    if err != nil {
        response.BadRequest(c, "invalid user id")
        return
    }
    data, err := h.svc.GetFollowing(targetID)
    if err != nil {
        response.InternalError(c, err.Error())
        return
    }
    response.OK(c, "following fetched", data)
}

func (h *Handler) GetPrivacySettings(c *gin.Context) {
    userID := middleware.GetUserID(c)
    settings, err := h.svc.GetPrivacySettings(userID)
    if err != nil {
        response.InternalError(c, err.Error())
        return
    }
    response.OK(c, "privacy settings fetched", settings)
}

func (h *Handler) UpdatePrivacySettings(c *gin.Context) {
    userID := middleware.GetUserID(c)
    var req UpdatePrivacySettingsRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.BadRequest(c, err.Error())
        return
    }
    settings, err := h.svc.UpdatePrivacySettings(userID, &req)
    if err != nil {
        response.InternalError(c, err.Error())
        return
    }
    response.OK(c, "privacy settings updated", settings)
}