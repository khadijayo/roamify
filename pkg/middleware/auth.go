package middleware

import (
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

		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			response.Unauthorized(c, "missing or malformed authorization header")
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")

		claims, err := pkgjwt.Parse(tokenStr, secret)
		if err != nil {
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
			response.Unauthorized(c, "invalid or expired token")
			c.Abort()
			return
		}

		if user.IsBanned {
			response.Forbidden(c, "your account has been banned")
			c.Abort()
			return
		}

		c.Set(UserIDKey, user.ID)
		c.Set(UserIDContextKey, user.ID)
		email := claims.Email
		if user.Email != nil && *user.Email != "" {
			email = *user.Email
		}
		c.Set(UserRoleKey, user.Role)
		c.Set(RoleContextKey, user.Role)
		c.Set(UserEmailKey, email)
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
	return value
}
