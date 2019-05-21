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
	"net/url"
	"strconv"
	"strings"
	"time"
	. "github.com/leyle/gsimplelog"
)

const (
	HTTP_GET_TIMEOUT = 10 // 10 seconds
	HTTP_POST_TIMEOUT = 10 // 10 seconds
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

func HttpPost(reqUrl string, data []byte) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, reqUrl, bytes.NewBuffer(data))
	if err != nil {
		Logger.Errorf("对 [%s] 创建 request 失败, %s", reqUrl, err.Error())
		return nil, err
	}

	client := &http.Client{
		Timeout: HTTP_POST_TIMEOUT * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		Logger.Errorf("对 [%s] 发起 client.Do() 操作失败, %s", reqUrl, err.Error())
		return nil, err
	}

	return resp, nil
}

func HttpGet(reqUrl string, values map[string][]string) (*http.Response, error) {
	// https://golang.org/pkg/net/url/#Values
	urlV := url.Values{}
	for k, vs := range values {
		if len(vs) == 1 {
			urlV.Set(k, vs[0])
		} else if len(vs) > 1 {
			for _, v := range vs {
				urlV.Add(k, v)
			}
		}
	}

	req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
	if err != nil {
		Logger.Errorf("生成 http get newrequest 失败, %s", err.Error())
		return nil, err
	}
	req.URL.RawQuery = urlV.Encode()

	client := &http.Client{
		Timeout: HTTP_GET_TIMEOUT * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		Logger.Errorf("发起get请求do[%s]失败, %s", reqUrl, err.Error())
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