package middleware

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/leyle/ginbase/consolelog"
	"github.com/leyle/ginbase/returnfun"
	"runtime/debug"
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

func (c *CustomErrStruct) Append(msg string) *CustomErrStruct {
	t := &CustomErrStruct{
		Code: c.Code,
		Msg:  c.Msg + msg,
	}
	return t
}

var ErrDbExec = &CustomErrStruct{
	Code: 5000,
	Msg:  "Database execute failed: ",
}

var ErrNoIdData = &CustomErrStruct{
	Code: 40000,
	Msg:  "No data for this id: ",
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
func RecoveryMiddleware(f func(*gin.Context, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rval := recover(); rval != nil {
				// runtime error, such as nil pointer dereference, should print stack
				prval := fmt.Sprintf("%v", rval)
				if strings.Contains(prval, "runtime") {
					consolelog.Logger.Error(GetReqId(c), prval)
					debug.PrintStack()
				}
				err, ok := rval.(error)
				if ok {
					f(c, err)
				} else {
					err, ok := rval.(string)
					if ok {
						f(c, errors.New(err))
					} else {
						// 简单处理
						emsg := fmt.Sprintf("%v", rval)
						consolelog.Logger.Error(GetReqId(c), emsg)
						f(c, errors.New(emsg))
					}
				}
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