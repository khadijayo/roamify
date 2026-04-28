package auth

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/khadijayo/roamify/pkg/response"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	res, err := h.svc.Register(c.Request.Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, ErrEmailAlreadyRegistered):
			response.Conflict(c, err.Error())
		case errors.Is(err, ErrVerificationEmailFailed):
			response.InternalError(c, err.Error())
		default:
			response.BadRequest(c, err.Error())
		}
		return
	}

	response.Created(c, "account created", res)
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	res, err := h.svc.Login(c.Request.Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, ErrEmailNotVerified), errors.Is(err, ErrInvalidCredentials):
			response.Unauthorized(c, err.Error())
		case errors.Is(err, ErrAccountBanned):
			response.Forbidden(c, err.Error())
		default:
			response.BadRequest(c, err.Error())
		}
		return
	}

	response.OK(c, "login successful", res)
}

func (h *Handler) SocialAuth(c *gin.Context) {
	var req SocialAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	res, err := h.svc.SocialAuth(c.Request.Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, ErrAccountBanned):
			response.Forbidden(c, err.Error())
		default:
			response.BadRequest(c, err.Error())
		}
		return
	}

	response.OK(c, "social login successful", res)
}

func (h *Handler) VerifyEmail(c *gin.Context) {
	res, err := h.svc.VerifyEmail(c.Request.Context(), c.Query("token"))
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidVerificationToken):
			response.BadRequest(c, err.Error())
		default:
			response.InternalError(c, err.Error())
		}
		return
	}

	response.OK(c, "email verified", res)
}

func (h *Handler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	res, err := h.svc.ForgotPassword(c.Request.Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, ErrResetCodeRecentlySent):
			response.TooManyRequests(c, err.Error())
		case errors.Is(err, ErrResetCodeSendFailed):
			response.InternalError(c, err.Error())
		default:
			response.BadRequest(c, err.Error())
		}
		return
	}

	response.OK(c, "password reset code requested", res)
}

func (h *Handler) VerifyResetCode(c *gin.Context) {
	var req VerifyResetCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	res, err := h.svc.VerifyResetCode(c.Request.Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidResetCode), errors.Is(err, ErrTooManyResetAttempts):
			response.BadRequest(c, err.Error())
		default:
			response.InternalError(c, err.Error())
		}
		return
	}

	response.OK(c, "reset code verified", res)
}

func (h *Handler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	res, err := h.svc.ResetPassword(c.Request.Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidResetCode), errors.Is(err, ErrTooManyResetAttempts):
			response.BadRequest(c, err.Error())
		default:
			response.InternalError(c, err.Error())
		}
		return
	}

	response.OK(c, "password reset successful", res)
}

func (h *Handler) ResendVerification(c *gin.Context) {
	var req ResendVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	res, err := h.svc.ResendVerification(c.Request.Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, ErrVerificationResendRateLimited):
			response.TooManyRequests(c, err.Error())
		case errors.Is(err, ErrVerificationResendFailed):
			response.InternalError(c, err.Error())
		default:
			response.BadRequest(c, err.Error())
		}
		return
	}

	response.OK(c, "verification email resent", res)
}
