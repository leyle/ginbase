package ginbase

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"strings"
	. "github.com/leyle/gsimplelog"
)

const REQUEST_ID_HEADER_KEY = "X-Request-Id"

var Debug = false

func DummyHandler(c *gin.Context) {
	ReturnJson(c, 501, 501, "暂未实现", "")
}

// 给每一个 request 设置一个 request id
func SetRequestIdMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqId := GenerateDataId()
		c.Request.Header.Set(REQUEST_ID_HEADER_KEY, reqId)
		c.Next()
	}
}

func SetResponseIdMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqId := c.Request.Header.Get(REQUEST_ID_HEADER_KEY)
		c.Writer.Header().Set(REQUEST_ID_HEADER_KEY, reqId)
		c.Next()
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, TOKEN, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
			return
		}
		c.Next()
	}
}

func LogMiddleware() gin.HandlerFunc {
	if !Debug {
		return func(c *gin.Context) {
			c.Next()
		}
	} else {
		return func(c *gin.Context) {
			logFunc(c)
		}
	}
}

func logFunc(c *gin.Context) {
	reqId := c.Request.Header.Get(REQUEST_ID_HEADER_KEY)

	uri := c.Request.RequestURI
	method := strings.ToUpper(c.Request.Method)
	ctype := strings.ToLower(c.Request.Header.Get("Content-Type"))
	if strings.Contains(ctype, "application/json") {
		var err error
		var body []byte
		var bodyStr string
		if c.Request.Body != nil {
			body, err = ioutil.ReadAll(c.Request.Body)
			if err != nil {
				Logger.Errorf("读取 requestbody 失败,%s", err.Error())
			}
			bodyStr = string(body)
			c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		}

		Logger.Debugf("REQUEST[%s]:[%s][%s]\n%s", reqId, method, uri, bodyStr)
	}

	c.Next()
}


