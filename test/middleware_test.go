package test

import (
	"github.com/gin-gonic/gin"
	. "github.com/leyle/ginbase/consolelog"
	"github.com/leyle/ginbase/middleware"
	"testing"
)

func TestLog(t *testing.T) {
	/*
	r := gin.New()
	r.Use(ReqIdMiddleware())
	r.Use(GinLogMiddleware())
	r.Use(CORSMiddleware())

	r.Use(RecoveryMiddleware(DefaultStopExecHandler))
	 */

	r := middleware.SetupGin()

	middleware.PrintHeader = false

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
	Logger.Info(middleware.GetReqId(c), "shiyxiiazhege")
	// time.Sleep(2 * time.Millisecond)
	// e := returnErr()
	// e := errors.New("some error")
	// middleware.StopExec(ErrInvalidMonth.Append(" xx yy zz"))
	// panic(212)
	// c.JSON(200, gin.H{"hello": "world"})
	middleware.StopExec(ErrInvalidMonth)

	m := returnNil()
	println(m.A)
}

var ErrInvalidMonth = &middleware.CustomErrStruct{
	Code: 4002,
	Msg:  "Invalid month format",
}

func returnErr() error {
	return ErrInvalidMonth
}

type Return struct {
	A string
}

func returnNil() *Return {
	return nil
}