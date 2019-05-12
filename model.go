package ginbase

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"time"
)

const (
	DEFAULT_LOCK_ACQUIRE_TIMEOUT = 5
	DEFAULT_LOCK_KEY_TIMEOUT = 5
)

type CurTime struct {
	Seconds int64 `json:"seconds" bson:"seconds"` // 精确到秒的时间戳
	HumanTime string `json:"humanTime" bson:"humanTime"` // 给人看的时间 2019-03-04 10:31:22
}

func GetCurTime() *CurTime {
	curT := time.Now()

	t := &CurTime{
		Seconds: curT.Unix(),
		HumanTime: curT.Format("2006-01-02 15:04:05"),
	}

	return t
}


func CurUnixTime() int64 {
	return time.Now().Unix()
}

func CurHumanTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func GetCurNoSpaceTime() string {
	return time.Now().Format("20060102150405")
}

type ApiRetDataForm struct {
	Code int `json:"code"`
	Msg string `json:"msg"`
	Data json.RawMessage `json:"data"`
}

type ReturnClientDataForm struct {
	Code int `json:"code"`
	Msg string `json:"msg"`
	Data interface{} `json:"data"`
}

func generateReturnData(code int, msg string, data interface{}) *ReturnClientDataForm {
	info := &ReturnClientDataForm {
		Code: code,
		Msg: msg,
		Data: data,
	}

	return info
}

func ReturnOKJson(c *gin.Context, data interface{}) {
	info := generateReturnData(200, "OK", data)
	c.JSON(200, info)
}

func ReturnErrJson(c *gin.Context, msg string) {
	Return400Json(c, 400, msg)
}

func Return400Json(c *gin.Context, code int, msg string) {
	info := generateReturnData(code, msg, "")
	e := fmt.Sprintf("错误码[%d]，错误信息[%s]", code, msg)
	fmt.Println(e)
	c.AbortWithStatusJSON(400, info)
}

func Return401Json(c *gin.Context, msg string) {
	info := generateReturnData(401, msg, "")
	e := fmt.Sprintf("错误码[%d]，错误信息[%s]", 401, msg)
	fmt.Println(e)
	c.AbortWithStatusJSON(401, info)
}

func Return403Json(c *gin.Context, msg string) {
	info := generateReturnData(403, msg, "")
	e := fmt.Sprintf("错误码[%d]，错误信息[%s]", 403, msg)
	fmt.Println(e)
	c.AbortWithStatusJSON(403, info)
}
