package auth

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/khadijayo/roamify/config"
	"github.com/khadijayo/roamify/pkg/response"
	"golang.org/x/oauth2"
)

type Handler struct {
	svc         Service
	googleOAuth *oauth2.Config
	frontendURL string
}

func NewHandler(svc Service, cfg *config.Config) *Handler {
	return &Handler{
		svc:         svc,
		googleOAuth: NewGoogleOAuthConfig(cfg),
		frontendURL: cfg.FrontendURL,
	}
}

// ── Existing handlers (unchanged) ────────────────────────────────────────────

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

// ── Google OAuth handlers ─────────────────────────────────────────────────────

// GoogleLogin redirects the user to Google's consent screen.
// GET /api/v1/auth/google/login
func (h *Handler) GoogleLogin(c *gin.Context) {
	if h.googleOAuth.ClientID == "" {
		response.InternalError(c, "Google OAuth is not configured on this server")
		return
	}

	// Use a random state token to prevent CSRF.
	// In production you should store this in a short-lived cookie or Redis.
	state := generateStateToken()
	c.SetCookie("oauth_state", state, 300, "/", "", true, true)

	url := h.googleOAuth.AuthCodeURL(state, oauth2.AccessTypeOffline)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// GoogleCallback handles the redirect back from Google after user consent.
// GET /api/v1/auth/google/callback
func (h *Handler) GoogleCallback(c *gin.Context) {
	// ── 1. Validate state (CSRF protection) ──────────────────────────────────
	cookieState, err := c.Cookie("oauth_state")
	queryState := c.Query("state")

	// If cookie is missing (e.g. browser blocked it), fall through gracefully.
	// If both are present, they must match.
	if err == nil && cookieState != "" && cookieState != queryState {
		h.redirectError(c, "invalid_state", "OAuth state mismatch – possible CSRF attempt")
		return
	}

	// Clear the state cookie immediately
	c.SetCookie("oauth_state", "", -1, "/", "", true, true)

	// ── 2. Check for error from Google ───────────────────────────────────────
	if errParam := c.Query("error"); errParam != "" {
		h.redirectError(c, errParam, c.Query("error_description"))
		return
	}

	// ── 3. Get the authorization code ────────────────────────────────────────
	code := c.Query("code")
	if code == "" {
		h.redirectError(c, "missing_code", "No authorization code received from Google")
		return
	}

	// ── 4. Exchange code for Google tokens ───────────────────────────────────
	res, err := h.svc.GoogleCallback(c.Request.Context(), code, h.googleOAuth)
	if err != nil {
		switch {
		case errors.Is(err, ErrGoogleTokenExchange):
			h.redirectError(c, "token_exchange_failed", "Failed to exchange code with Google")
		case errors.Is(err, ErrGoogleUserInfo):
			h.redirectError(c, "userinfo_failed", "Failed to fetch user info from Google")
		case errors.Is(err, ErrGoogleMissingEmail):
			h.redirectError(c, "missing_email", "Google account has no email address")
		case errors.Is(err, ErrAccountBanned):
			h.redirectError(c, "account_banned", "Your account has been banned")
		default:
			h.redirectError(c, "server_error", "An unexpected error occurred")
		}
		return
	}

	// ── 5. Redirect to frontend with JWT ─────────────────────────────────────
	redirectURL := h.frontendURL + "/oauth-success?token=" + res.Token
	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

// redirectError sends the user to the frontend with a readable error.
func (h *Handler) redirectError(c *gin.Context, code, message string) {
	redirectURL := h.frontendURL + "/oauth-error?error=" + code + "&message=" + message
	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}
