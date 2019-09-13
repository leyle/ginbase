package middleware

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	. "github.com/leyle/ginbase/consolelog"
	"io/ioutil"
	"strings"
	"time"
)

var IgnoreReadReqBodyPath = []string{}

func GinLogMiddleware() gin.HandlerFunc {
	// 一个日志地方，处理两个信息
	// 一个是记录输入
	// 一个是记录最后的输出
	// 一个请求的完整的生命周期都可以看得到
	return func(c *gin.Context) {
		startT := time.Now()
		path := c.Request.RequestURI
		method := c.Request.Method
		ctype := strings.ToLower(c.Request.Header.Get("Content-Type"))
		clientIp := c.ClientIP()
		reqMsg := fmt.Sprintf("[Req][%s][%s][%s][%s]", method, path, clientIp, ctype)
		// 判断是否有 request body，如果有，就转存读取

		if c.Request.ContentLength > 0 && !ignoreReadBody(c.Request.URL.Path) {
			body, err := ioutil.ReadAll(c.Request.Body)
			if err != nil {
				// 忽略掉错误，不继续处理
				Logger.Errorf(c, "读取请求body失败, %s", err.Error())
			} else {
				// 还原回去
				c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
				if len(body) > 0 {
					reqMsg += "\n" + string(body)
				}
			}
		}

		Logger.Info(c, reqMsg)

		// rewrite writer，方便后续转存数据
		c.Writer = &respWriter{
			ResponseWriter: c.Writer,
			cache:          bytes.NewBufferString(""),
		}

		c.Next()

		// 下面的内容会在请求结束后执行
		latency := time.Now().Sub(startT)
		statusCode := c.Writer.Status()

		respMsg := fmt.Sprintf("[Resp][%s][%s][%d][%v]", method, path, statusCode, latency)

		rw, ok := c.Writer.(*respWriter)
		if !ok {
			Logger.Warnf(c, "处理response数据，转回respwriter失败")
		} else {
			if rw.cache.Len() > 0 {
				respMsg += "\n" + rw.cache.String()
			}
		}

		Logger.Info(c, respMsg)
	}
}

func ignoreReadBody(reqPath string) bool {
	for _, path := range IgnoreReadReqBodyPath {
		if reqPath == path {
			return true
		}
	}
	return false
}

// rewrite Write()
type respWriter struct {
	gin.ResponseWriter
	cache *bytes.Buffer
}

func (r *respWriter) Write(b []byte) (int, error) {
	r.cache.Write(b)
	return r.ResponseWriter.Write(b)
}