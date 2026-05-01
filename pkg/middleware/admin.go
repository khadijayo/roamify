package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/khadijayo/roamify/pkg/response"
)

func AdminOnly() gin.HandlerFunc {
	return RequireAdmin()
}

func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get(RoleContextKey)
		if !exists || role != "admin" {
			response.Forbidden(c, "admin access required")
			c.Abort()
			return
		}
		c.Next()
	}
}
