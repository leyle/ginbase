package ginbase

import "github.com/gin-gonic/gin"

func DummyHandler(c *gin.Context) {
	ReturnJson(c, 501, 501, "暂未实现", "")
}