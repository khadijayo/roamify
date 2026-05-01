package auth

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.RouterGroup, h *Handler) {
	authRoutes := r.Group("/auth")
	{
		authRoutes.POST("/register", h.Register)
		authRoutes.POST("/login", h.Login)
		authRoutes.POST("/social", h.SocialAuth)
		authRoutes.GET("/verify", h.VerifyEmailPage)
		authRoutes.GET("/verify-email", h.VerifyEmail)
		authRoutes.POST("/forgot-password", h.ForgotPassword)
		authRoutes.POST("/verify-reset-code", h.VerifyResetCode)
		authRoutes.POST("/reset-password", h.ResetPassword)
		authRoutes.POST("/resend-verification", h.ResendVerification)
	}
}
