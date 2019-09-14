package test

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/leyle/ginbase"
	"github.com/leyle/ginbase/consolelog"
	"github.com/leyle/ginbase/util"
	"testing"
)

func TestColorFormat(t *testing.T) {
	// a := fmt.Sprintf("%c[1;0;34m[DEBUG]%c[0m\n", 0x1B, 0x1B)
	fmt.Print(consolelog.DebugColor)
	fmt.Print(consolelog.InfoColor)
	fmt.Print(consolelog.WarnColor)
	fmt.Print(consolelog.ErrorColor)

	fmt.Println("xxxxxxxxxxxxxxxxxxxxx")

	hello := "some hello msg"

	c := &gin.Context{}
	c.Set(ginbase.ReqIdKey, util.GenerateDataId())

	consolelog.Logger.Debug(c, hello)
	consolelog.Logger.Debugf(c, "[Requst]%s world", hello)

	consolelog.Logger.Info(c, hello)
	consolelog.Logger.Infof(c, "[Requst]%s world", hello)

	consolelog.Logger.Warn(c, hello)
	consolelog.Logger.Warnf(c, "[Requst]%s world", hello)

	consolelog.Logger.Error(c, hello)
	consolelog.Logger.Errorf(c, "[Requst]%s world", hello)
}
