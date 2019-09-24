package test

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/leyle/ginbase/consolelog"
	"github.com/leyle/ginbase/constant"
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
	c.Set(constant.ReqIdKey, util.GenerateDataId())

	reqId := util.GenerateDataId()

	consolelog.Logger.Debug(reqId, hello)
	consolelog.Logger.Debugf(reqId, "[Requst]%s world", hello)

	consolelog.Logger.Info(reqId, hello)
	consolelog.Logger.Infof(reqId, "[Requst]%s world", hello)

	consolelog.Logger.Warn(reqId, hello)
	consolelog.Logger.Warnf(reqId, "[Requst]%s world", hello)

	consolelog.Logger.Error(reqId, hello)
	consolelog.Logger.Errorf(reqId, "[Requst]%s world", hello)
}
