package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		method := c.Request.Method

		if query != "" {
			path = path + "?" + query
		}

		fmt.Printf("[ROAMIFY] %v | %3d | %13v | %-7s %s\n",
			start.Format("2006/01/02 - 15:04:05"),
			status,
			latency,
			method,
			path,
		)
	}
}