package ginbase

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	. "github.com/leyle/gsimplelog"
	"github.com/pkg/errors"
	"net/http"
)

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
	ReturnJson(c, 200, 200, "OK", data)
}

func ReturnErrJson(c *gin.Context, msg string) {
	Return400Json(c, 400, msg)
}

func Return400Json(c *gin.Context, code int, msg string) {
	ReturnJson(c, 400, code, msg, "")
}

func Return401Json(c *gin.Context, msg string) {
	ReturnJson(c, 401, 401, msg, "")
}

func Return403Json(c *gin.Context, msg string) {
	ReturnJson(c, 403, 403, msg, "")
}

func ReturnJson(c *gin.Context, statusCode, code int, msg string, data interface{}) {
	text := http.StatusText(statusCode)
	if text == "" {
		panic(errors.New("错误的 statusCode"))
	}

	ret := generateReturnData(code, msg, data)
	if statusCode != http.StatusOK {
		Logger.Errorf("RESPONSE[%s][%s][%s][%d][%d][%s]", c.GetString(REQUEST_ID), c.Request.RequestURI, c.Request.Method, statusCode, code, msg)
		c.AbortWithStatusJSON(statusCode, ret)
	} else {
		Logger.Debugf("RESPONSE[%s][%s][%s][%s]", c.GetString(REQUEST_ID), c.Request.RequestURI, c.Request.Method, data)
		c.JSON(statusCode, ret)
	}
}