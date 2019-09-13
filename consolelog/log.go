package consolelog

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/leyle/ginbase"
)

const (
	LogLevelDebug = iota
	LoglevelInfo
	LogLevelWarn
	LogLevelError
	LogLevelOff
)

var (
	DebugColor = ""
	InfoColor = ""
	WarnColor = ""
	ErrorColor = ""
)

func init() {
	DebugColor = fmt.Sprintf("%c[1;0;34m[DEBUG]%c[0m", 0x1B, 0x1B)
	InfoColor = fmt.Sprintf("%c[1;0;32m[INFO]%c[0m", 0x1B, 0x1B)
	WarnColor = fmt.Sprintf("%c[1;0;33m[WARNING]%c[0m", 0x1B, 0x1B)
	ErrorColor = fmt.Sprintf("%c[1;0;31m[ERROR]%c[0m", 0x1B, 0x1B)
}


type ConsoleLog struct {
	Level int
}

var Logger = ConsoleLog{
	Level: LogLevelDebug,
}

func (l *ConsoleLog) SetLogLevel(level int) {
	l.Level = level
}

func (l *ConsoleLog) Debug(c *gin.Context, ps ...interface{}) {
	if l.Level <= LogLevelDebug {
		reqId := getReqId(c)
		fmt.Printf("%s[%s][%s]|%s\n", DebugColor, reqId, ginbase.CurHumanTime(), fmt.Sprint(ps ...))
	}
}

func (l *ConsoleLog) Debugf(c *gin.Context, format string, ps ...interface{}) {
	if l.Level <= LogLevelDebug {
		reqId := getReqId(c)
		fmt.Printf("%s[%s][%s]|%s\n", DebugColor, reqId, ginbase.CurHumanTime(), fmt.Sprintf(format, ps ...))
	}
}

func (l *ConsoleLog) Info(c *gin.Context, ps ...interface{}) {
	if l.Level <= LoglevelInfo {
		reqId := getReqId(c)
		fmt.Printf("%s[%s][%s]|%s\n", InfoColor, reqId, ginbase.CurHumanTime(), fmt.Sprint(ps ...))
	}
}

func (l *ConsoleLog) Infof(c *gin.Context, format string, ps ...interface{}) {
	if l.Level <= LoglevelInfo {
		reqId := getReqId(c)
		fmt.Printf("%s[%s][%s]|%s\n", InfoColor, reqId, ginbase.CurHumanTime(), fmt.Sprintf(format, ps ...))
	}
}

func (l *ConsoleLog) Warn(c *gin.Context, ps ...interface{}) {
	if l.Level <= LogLevelWarn {
		reqId := getReqId(c)
		fmt.Printf("%s[%s][%s]|%s\n", WarnColor, reqId, ginbase.CurHumanTime(), fmt.Sprint(ps ...))
	}
}

func (l *ConsoleLog) Warnf(c *gin.Context, format string, ps ...interface{}) {
	if l.Level <= LogLevelWarn {
		reqId := getReqId(c)
		fmt.Printf("%s[%s][%s]|%s\n", WarnColor, reqId, ginbase.CurHumanTime(), fmt.Sprintf(format, ps ...))
	}
}

func (l *ConsoleLog) Error(c *gin.Context, ps ...interface{}) {
	if l.Level <= LogLevelError {
		reqId := getReqId(c)
		fmt.Printf("%s[%s][%s]|%s\n", ErrorColor, reqId, ginbase.CurHumanTime(), fmt.Sprint(ps ...))
	}
}

func (l *ConsoleLog) Errorf(c *gin.Context, format string, ps ...interface{}) {
	if l.Level <= LogLevelError {
		reqId := getReqId(c)
		fmt.Printf("%s[%s][%s]|%s\n", ErrorColor, reqId, ginbase.CurHumanTime(), fmt.Sprintf(format, ps ...))
	}
}

func getReqId(c *gin.Context) string {
	reqId, ok := c.Get(ginbase.ReqIdKey)
	if !ok {
		// 必须panic，因为属于程序错误，忘记配置 reqid 了
		panic("忘记配置reqid，请检查程序中间件")
	}
	return reqId.(string)
}