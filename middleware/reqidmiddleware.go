package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/leyle/ginbase/constant"
	"github.com/leyle/ginbase/util"
)

const DefaultReqId = "NoReqId"

func ReqIdMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqId := util.GenerateDataId()
		c.Set(constant.ReqIdKey, reqId)
		c.Writer.Header().Set(constant.XRequestIdHeaderKey, reqId)
		c.Next()
	}
}

func GetReqId(c *gin.Context) string {
	reqId, ok := c.Get(constant.ReqIdKey)
	if !ok {
		// 必须panic，因为属于程序错误，忘记配置 reqid 了
		panic("忘记配置reqid，请检查程序中间件")
	}
	return reqId.(string)
}
