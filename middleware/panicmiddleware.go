package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/leyle/ginbase/consolelog"
	"github.com/leyle/ginbase/returnfun"
	"strconv"
	"strings"
)

const DefaultCustomErrCode = 4000 // 默认的错误码，通用的错误码
const DefaultSep = "|"

type CustomErrStruct struct {
	Code int
	Msg string
}

func (c *CustomErrStruct) Error() string {
	return fmt.Sprintf("%d|%s", c.Code, c.Msg)
}

// 反向解析出来 code 和 msg
func ParseCustomErr(err error) *CustomErrStruct {
	msg := err.Error()
	if !strings.Contains(msg, DefaultSep)  {
		return &CustomErrStruct{
			Code: DefaultCustomErrCode,
			Msg:  msg,
		}
	}

	ret := strings.SplitN(msg, DefaultSep, 2)
	scode := ret[0]
	emsg := ret[1]

	return &CustomErrStruct{
		Code: parseStrCode(scode),
		Msg:  strings.TrimSpace(emsg),
	}
}

// 解析字符串格式的 code， 如果解析失败，就返回默认 code
func parseStrCode(code string) int {
	code = strings.TrimSpace(code)
	c, err := strconv.ParseInt(code, 10, 64)
	if err != nil {
		consolelog.Logger.Errorf("", "parse err code failed, %s", err.Error())
		return DefaultCustomErrCode
	}
	return int(c)
}

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
	cerr := ParseCustomErr(err)
	returnfun.ReturnJson(c, 400, cerr.Code, cerr.Msg, "")
}