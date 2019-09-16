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

// 是否打印请求头
var PrintHeader = false

// 不支持通配符，string equals 匹配模式
var ignoreReadReqBodyPath = []string{}

func AddIgnoreReadReqBodyPath(paths ...string) {
	ignoreReadReqBodyPath = append(ignoreReadReqBodyPath, paths ...)
}

func isIgnoreReadBodyPath(reqPath string) bool {
	for _, path := range ignoreReadReqBodyPath {
		if reqPath == path {
			return true
		}
	}
	return false
}

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

		// 判断是否打印 header
		if PrintHeader {
			hmsg := ""
			for k, v := range c.Request.Header {
				val := strings.Join(v, " ")
				hmsg += fmt.Sprintf("%s:%s\n", k, val)
			}
			if hmsg != "" {
				reqMsg += "\n" + hmsg[0: len(hmsg) - 1]
			}
		}

		// 判断是否有 request body，如果有，就转存读取
		if c.Request.ContentLength > 0 && !isIgnoreReadBodyPath(c.Request.URL.Path) {
			body, err := ioutil.ReadAll(c.Request.Body)
			if err != nil {
				// 忽略掉错误，不继续处理
				Logger.Errorf(GetReqId(c), "读取请求body失败, %s", err.Error())
			} else {
				// 还原回去
				c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
				if len(body) > 0 {
					reqMsg += "\n" + string(body)
				}
			}
		}

		Logger.Info(GetReqId(c), reqMsg)

		// rewrite writer，方便后续转存数据
		c.Writer = &respWriter{
			ResponseWriter: c.Writer,
			cache:          bytes.NewBufferString(""),
		}

		c.Next()

		// 下面的内容会在请求结束后执行
		statusCode := c.Writer.Status()
		respBody := ""

		respMsg := fmt.Sprintf("[Resp][%s][%s][%d]", method, path, statusCode)

		rw, ok := c.Writer.(*respWriter)
		if !ok {
			Logger.Warnf(GetReqId(c), "处理response数据，转回respwriter失败")
		} else {
			if rw.cache.Len() > 0 {
				respBody = "\n" + rw.cache.String()
			}
		}

		latency := time.Now().Sub(startT)
		respMsg += fmt.Sprintf("[%v]", latency)
		if respBody != "" {
			respMsg += respBody
		}

		Logger.Info(GetReqId(c), respMsg)
	}
}

// rewrite Write()
type respWriter struct {
	gin.ResponseWriter
	cache *bytes.Buffer
}

// 会导致内存增加，性能稍微降低，但是我觉得值得
func (r *respWriter) Write(b []byte) (int, error) {
	r.cache.Write(b)
	return r.ResponseWriter.Write(b)
}