package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	pkgjwt "github.com/khadijayo/roamify/pkg/jwt"
	"github.com/khadijayo/roamify/pkg/response"
)

const UserIDKey = "userID"
const UserEmailKey = "userEmail"

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			response.Unauthorized(c, "missing or malformed authorization header")
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")
		claims, err := pkgjwt.Parse(tokenStr)
		if err != nil {
			response.Unauthorized(c, "invalid or expired token")
			c.Abort()
			return
		}

		c.Set(UserIDKey, claims.UserID)
		c.Set(UserEmailKey, claims.Email)
		c.Next()
	}
}

// GetUserID is a helper for handlers to extract the authenticated user's ID.
func GetUserID(c *gin.Context) uuid.UUID {
	id, _ := c.Get(UserIDKey)
	return id.(uuid.UUID)
}