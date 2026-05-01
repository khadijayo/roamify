package auth

import (
	"errors"
	"net/http"

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

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "User registered successfully. Check your email to verify.",
		"data":    res,
	})
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

func (h *Handler) VerifyEmailPage(c *gin.Context) {
	if _, err := h.svc.VerifyEmail(c.Request.Context(), c.Query("token")); err != nil {
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(`
<html>
  <body style="font-family:sans-serif;text-align:center;margin-top:50px;">
    <h2>Email verification failed</h2>
    <p>The verification link is invalid or expired.</p>
  </body>
</html>
`))
		return
	}

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(`
<html>
  <body style="font-family:sans-serif;text-align:center;margin-top:50px;">
    <h2>Email verified successfully</h2>
    <p>You can now log in.</p>
  </body>
</html>
`))
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
