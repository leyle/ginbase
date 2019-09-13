package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/leyle/ginbase/consolelog"
	"testing"
	"time"
)

func TestLog(t *testing.T) {
	r := gin.New()
	r.Use(ReqIdMiddleware())
	r.Use(GinLogMiddleware())

	// IgnoreReadReqBodyPath = []string{"/api/hello"}

	router := r.Group("/api")
	router.Any("/hello", handler)

	addr := "0.0.0.0:9000"
	err := r.Run(addr)
	if err != nil {
		t.Error(err)
	}
}

func handler(c *gin.Context) {
	consolelog.Logger.Info(c, "shiyxiiazhege")
	time.Sleep(2 * time.Millisecond)
	c.JSON(200, gin.H{"hello": "world"})
}