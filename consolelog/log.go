package consolelog

import (
	"fmt"
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

func (l *ConsoleLog) Debug(reqId string, ps ...interface{}) {
	if l.Level <= LogLevelDebug {
		fmt.Printf("%s[%s][%s]|%s\n", DebugColor, reqId, ginbase.CurHumanTime(), fmt.Sprint(ps ...))
	}
}

func (l *ConsoleLog) Debugf(reqId string, format string, ps ...interface{}) {
	if l.Level <= LogLevelDebug {
		fmt.Printf("%s[%s][%s]|%s\n", DebugColor, reqId, ginbase.CurHumanTime(), fmt.Sprintf(format, ps ...))
	}
}

func (l *ConsoleLog) Info(reqId string, ps ...interface{}) {
	if l.Level <= LoglevelInfo {
		fmt.Printf("%s[%s][%s]|%s\n", InfoColor, reqId, ginbase.CurHumanTime(), fmt.Sprint(ps ...))
	}
}

func (l *ConsoleLog) Infof(reqId string, format string, ps ...interface{}) {
	if l.Level <= LoglevelInfo {
		fmt.Printf("%s[%s][%s]|%s\n", InfoColor, reqId, ginbase.CurHumanTime(), fmt.Sprintf(format, ps ...))
	}
}

func (l *ConsoleLog) Warn(reqId string, ps ...interface{}) {
	if l.Level <= LogLevelWarn {
		fmt.Printf("%s[%s][%s]|%s\n", WarnColor, reqId, ginbase.CurHumanTime(), fmt.Sprint(ps ...))
	}
}

func (l *ConsoleLog) Warnf(reqId string, format string, ps ...interface{}) {
	if l.Level <= LogLevelWarn {
		fmt.Printf("%s[%s][%s]|%s\n", WarnColor, reqId, ginbase.CurHumanTime(), fmt.Sprintf(format, ps ...))
	}
}

func (l *ConsoleLog) Error(reqId string, ps ...interface{}) {
	if l.Level <= LogLevelError {
		fmt.Printf("%s[%s][%s]|%s\n", ErrorColor, reqId, ginbase.CurHumanTime(), fmt.Sprint(ps ...))
	}
}

func (l *ConsoleLog) Errorf(reqId string, format string, ps ...interface{}) {
	if l.Level <= LogLevelError {
		fmt.Printf("%s[%s][%s]|%s\n", ErrorColor, reqId, ginbase.CurHumanTime(), fmt.Sprintf(format, ps ...))
	}
}

