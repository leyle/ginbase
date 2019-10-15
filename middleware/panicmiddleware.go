package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/leyle/ginbase/returnfun"
)

// panic 后可以给客户端返回一个期望的数据格式

// 抛出错误
func StopExec(err error) {
	if err == nil {
		return
	}
	panic(err)
}

// 恢复回来
func RecoveryMiddleware(f func(c *gin.Context, err error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				f(c, err.(error))
			}
		}()

		c.Next()
	}
}

// 提供一个默认的 recoveryhandler
func DefaultStopExecHandler(c *gin.Context, err error) {
	returnfun.ReturnJson(c, 400, 400, err.Error(), "")
}