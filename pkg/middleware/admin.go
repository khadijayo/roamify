package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/khadijayo/roamify/pkg/response"
)

func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		if GetUserRole(c) != "admin" {
			response.Forbidden(c, "admin access required")
			c.Abort()
			return
		}
		c.Next()
	}
}
