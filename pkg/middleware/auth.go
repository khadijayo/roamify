package middleware

import (
	"log"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	pkgjwt "github.com/khadijayo/roamify/pkg/jwt"
	"github.com/khadijayo/roamify/pkg/response"
	"gorm.io/gorm"
)

const UserIDKey = "userID"
const UserEmailKey = "userEmail"
const UserRoleKey = "userRole"
const UserIDContextKey = "user_id"
const RoleContextKey = "role"

type authUserRecord struct {
	ID       uuid.UUID `gorm:"column:id"`
	Email    *string   `gorm:"column:email"`
	Role     string    `gorm:"column:role"`
	IsBanned bool      `gorm:"column:is_banned"`
}

func Auth(secret string, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		tokenStr, ok := extractBearerToken(header)
		if !ok {
			debugAuth("denied malformed authorization header path=%s method=%s header_present=%t", c.FullPath(), c.Request.Method, header != "")
			response.Unauthorized(c, "missing or malformed authorization header")
			c.Abort()
			return
		}

		claims, err := pkgjwt.Parse(tokenStr, secret)
		if err != nil {
			debugAuth("denied invalid token path=%s method=%s error=%v", c.FullPath(), c.Request.Method, err)
			response.Unauthorized(c, "invalid or expired token")
			c.Abort()
			return
		}

		var user authUserRecord
		if err := db.WithContext(c.Request.Context()).
			Table("users").
			Select("id, email, role, is_banned").
			Where("id = ? AND deleted_at IS NULL", claims.UserID).
			First(&user).Error; err != nil {
			debugAuth("denied user lookup failed path=%s method=%s token_user_id=%s error=%v", c.FullPath(), c.Request.Method, claims.UserID, err)
			response.Unauthorized(c, "invalid or expired token")
			c.Abort()
			return
		}

		if user.IsBanned {
			debugAuth("denied banned user path=%s method=%s user_id=%s", c.FullPath(), c.Request.Method, user.ID)
			response.Forbidden(c, "your account has been banned")
			c.Abort()
			return
		}

		role := normalizeRole(user.Role)
		if role == "" {
			role = "user"
		}

		c.Set(UserIDKey, user.ID)
		c.Set(UserIDContextKey, user.ID)
		email := strings.ToLower(strings.TrimSpace(claims.Email))
		if user.Email != nil && *user.Email != "" {
			email = strings.ToLower(strings.TrimSpace(*user.Email))
		}
		c.Set(UserRoleKey, role)
		c.Set(RoleContextKey, role)
		c.Set(UserEmailKey, email)
		c.Set("token_role", normalizeRole(claims.Role))
		c.Set("token_email", strings.ToLower(strings.TrimSpace(claims.Email)))
		debugAuth("authorized path=%s method=%s user_id=%s db_role=%s token_role=%s", c.FullPath(), c.Request.Method, user.ID, role, normalizeRole(claims.Role))
		c.Next()
	}
}

func RequireAuth(secret string, db *gorm.DB) gin.HandlerFunc {
	return Auth(secret, db)
}

func GetUserID(c *gin.Context) uuid.UUID {
	id, _ := c.Get(UserIDKey)
	return id.(uuid.UUID)
}

func GetUserRole(c *gin.Context) string {
	role, ok := c.Get(UserRoleKey)
	if !ok {
		return ""
	}
	value, _ := role.(string)
	return normalizeRole(value)
}

func extractBearerToken(header string) (string, bool) {
	parts := strings.Fields(strings.TrimSpace(header))
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
		return "", false
	}
	return strings.TrimSpace(parts[1]), true
}

func normalizeRole(role string) string {
	return strings.ToLower(strings.TrimSpace(role))
}

func authDebugEnabled() bool {
	return strings.EqualFold(os.Getenv("AUTH_DEBUG"), "true")
}

func debugAuth(format string, args ...interface{}) {
	if authDebugEnabled() {
		log.Printf("[auth] "+format, args...)
	}
}
