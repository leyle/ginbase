package ginbase

import (
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"strconv"
	"strings"
	"time"
	. "github.com/leyle/gsimplelog"
)

func Sha256(data string) string {
	h := sha256.New()
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func Md5(data string) string {
	m := md5.New()
	m.Write([]byte(data))
	return hex.EncodeToString(m.Sum(nil))
}

func GenerateDataId() string {
	id := bson.NewObjectId().Hex()
	return id
}

func GenerateHashPasswd(loginId, rawPasswd string) string {
	d := strings.ToLower(loginId) + rawPasswd
	h := Sha256(d)
	return h
}

// 生成 token
// userid + curtimesec 然后 sha256 hash 值
func GenerateToken(userId string) string {
	d := fmt.Sprintf("%s%d", userId, time.Now().Nanosecond())
	h := Md5(d)
	h = strings.ToUpper(h)
	return h
}

func HttpPost(url string, data []byte) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		Logger.Errorf("对 [%s] 创建 request 失败, %s", url, err.Error())
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		Logger.Errorf("对 [%s] 发起 client.Do() 操作失败, %s", err.Error())
		return nil, err
	}

	return resp, nil
}

func GetPageAndSize(c *gin.Context) (page, size, skip int) {
	p := c.Query("page")
	s := c.Query("size")

	if p != "" {
		page, _ = strconv.Atoi(p)
	} else {
		page = 1
	}

	if s != "" {
		size, _ = strconv.Atoi(s)
	} else {
		size = 10
	}

	if page < 1 {
		page = 1
	}

	if size > 20 {
		size = 20
	}

	skip = (page - 1) * size

	return
}

// 去重 string
type void struct{}
func UniqueStringArray(items []string) []string {
	var member void
	datas := make(map[string]void)
	for _, item := range items {
		datas[item] = member
	}
	var ret []string
	for k, _ := range datas {
		ret = append(ret, k)
	}

	return ret
}