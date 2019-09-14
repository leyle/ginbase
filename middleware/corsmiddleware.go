package middleware

import (
	"github.com/gin-gonic/gin"
	"strings"
)

var allowHeaders = []string{
	"Content-Type",
	"TOKEN",
	"Content-Length",
	"Accept-Encoding",
	"X-CSRF-Token",
	"Authorization",
	"accept",
	"origin",
	"Cache-Control",
	"X-Requested-With",
}

func AddAllowHeaders(val string) {
	allowHeaders = append(allowHeaders, val)
}

func getAllowHeaders() string {
	return strings.Join(allowHeaders, ",")
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", getAllowHeaders())
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, HEAD, CONNECT, TRACE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
			return
		}
		c.Next()
	}
}
