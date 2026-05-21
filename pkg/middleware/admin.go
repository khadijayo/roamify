package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/khadijayo/roamify/pkg/response"
)

func AdminOnly() gin.HandlerFunc {
	return RequireAdmin()
}

func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		roleValue, exists := c.Get(RoleContextKey)
		role := normalizeRole(fmt.Sprint(roleValue))
		if !exists || role != "admin" {
			userID, _ := c.Get(UserIDContextKey)
			debugAuth("admin denied path=%s method=%s user_id=%v role=%q", c.FullPath(), c.Request.Method, userID, role)
			response.Forbidden(c, "admin access required")
			c.Abort()
			return
		}
		c.Next()
	}
}
