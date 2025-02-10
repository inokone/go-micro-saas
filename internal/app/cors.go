package app

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func SubdomainAllowingCORS() gin.HandlerFunc {

	const domainName = "example.com"

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		isHttp := strings.HasPrefix(origin, "http://")
		isHttps := strings.HasPrefix(origin, "https://")

		// Allow requests from localhost:3000 (DEVELOPMENT) and example.com (PRODUCTION)
		if origin != "" &&
			((isHttp && (strings.HasSuffix(origin, ".localhost:3000") || strings.HasSuffix(origin, "/localhost:3000"))) ||

				(isHttps && (strings.HasSuffix(origin, "."+domainName) || strings.HasSuffix(origin, "/"+domainName)))) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Length, Content-Type, Authorization")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

			// Handle preflight requests
			if c.Request.Method == http.MethodOptions {
				c.AbortWithStatus(http.StatusOK)
				return
			}
		}

		c.Next()
	}
}
