package ginbase

import (
	"github.com/gin-gonic/gin"
	. "github.com/leyle/gsimplelog"
	"net/http/httputil"
	"strings"
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
	const formData = "multipart/form-data"
	if strings.Contains(ctype, formData) {
		Logger.Debugf("REQUEST[%s][%s][%s]", reqId, uri, method)
	} else {
		if Debug {
			rawData, err := httputil.DumpRequest(c.Request, true)
			if err != nil {
				Logger.Errorf("dump request failed, %s", err.Error())
			} else {
				Logger.Debugf("REQUEST[%s]\n%s", reqId, string(rawData))
			}
		} else {
			Logger.Debugf("REQUEST[%s][%s][%s]", reqId, uri, method)
		}
	}

	c.Next()
}


