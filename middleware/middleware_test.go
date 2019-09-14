package middleware

import (
	"errors"
	"github.com/gin-gonic/gin"
	. "github.com/leyle/ginbase/consolelog"
	"testing"
	"time"
)

func TestLog(t *testing.T) {
	/*
	r := gin.New()
	r.Use(ReqIdMiddleware())
	r.Use(GinLogMiddleware())
	r.Use(CORSMiddleware())

	r.Use(RecoveryMiddleware(DefaultStopExecHandler))
	 */

	r := SetupGin()

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
	Logger.Info(GetReqId(c), "shiyxiiazhege")
	time.Sleep(2 * time.Millisecond)
	StopExec(errors.New("one error"))
	c.JSON(200, gin.H{"hello": "world"})
}