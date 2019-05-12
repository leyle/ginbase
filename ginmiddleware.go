package ginbase

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"os"
	"strings"
)

func DumpHandler(c *gin.Context) {

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
	return func(c *gin.Context) {
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
					fmt.Println("读取消息body 失败 ", err.Error())
				}
				bodyStr = string(body)
				c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
			}
			if strings.Contains(uri, "msg/recv") {
				_, _ = fmt.Fprintf(os.Stdout, "REQUEST: [%s][%s]\n", method, uri)
			} else {
				_, _ = fmt.Fprintf(os.Stdout, "REQUEST: [%s][%s]Body:\n%s\n", method, uri, bodyStr)
			}
		}

		c.Next()
	}
}

