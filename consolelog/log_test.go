package consolelog

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/leyle/ginbase"
	"github.com/leyle/ginbase/returnfun"
	"testing"
)

func TestColorFormat(t *testing.T) {
	// a := fmt.Sprintf("%c[1;0;34m[DEBUG]%c[0m\n", 0x1B, 0x1B)
	fmt.Print(DebugColor)
	fmt.Print(InfoColor)
	fmt.Print(WarnColor)
	fmt.Print(ErrorColor)

	fmt.Println("xxxxxxxxxxxxxxxxxxxxx")

	hello := "some hello msg"

	c := &gin.Context{}
	c.Set(ginbase.ReqIdKey, returnfun.GenerateDataId())

	Logger.Debug(c, hello)
	Logger.Debugf(c, "[Requst]%s world", hello)

	Logger.Info(c, hello)
	Logger.Infof(c, "[Requst]%s world", hello)

	Logger.Warn(c, hello)
	Logger.Warnf(c, "[Requst]%s world", hello)

	Logger.Error(c, hello)
	Logger.Errorf(c, "[Requst]%s world", hello)
}
