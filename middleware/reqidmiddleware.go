package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/leyle/ginbase"
)

const DefaultReqId = "NoReqId"

func ReqIdMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqId := ginbase.GenerateDataId()
		c.Set(ginbase.ReqIdKey, reqId)
		c.Writer.Header().Set(ginbase.XRequestIdHeaderKey, reqId)
		c.Next()
	}
}

func GetReqId(c *gin.Context) string {
	reqId, ok := c.Get(ginbase.ReqIdKey)
	if !ok {
		// 必须panic，因为属于程序错误，忘记配置 reqid 了
		panic("忘记配置reqid，请检查程序中间件")
	}
	return reqId.(string)
}
