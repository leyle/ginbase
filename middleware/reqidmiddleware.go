package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/leyle/ginbase"
)

func ReqIdMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(ginbase.ReqIdKey, ginbase.GenerateDataId())
		c.Next()
	}
}
