package middleware

import "github.com/gin-gonic/gin"

func SetupGin() *gin.Engine {
	e := gin.New()
	e.Use(ReqIdMiddleware())
	e.Use(GinLogMiddleware())
	e.Use(CORSMiddleware())
	e.Use(RecoveryMiddleware(DefaultStopExecHandler))

	return e
}
