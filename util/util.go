package util

import (
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/gin-gonic/gin"
	. "github.com/leyle/ginbase/consolelog"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var HTTP_GET_TIMEOUT time.Duration = 10  // 10 seconds
var HTTP_POST_TIMEOUT time.Duration = 10 // 10 seconds

type CurTime struct {
	Second    int64  `json:"second" bson:"second"`
	HumanTime string `json:"humanTime" bson:"humanTime"` // 给人看的时间 2019-03-04 10:31:22
}

func GetCurTime() *CurTime {
	curT := time.Now()

	t := &CurTime{
		Second:    curT.Unix(),
		HumanTime: curT.Format("2006-01-02 15:04:05"),
	}

	return t
}

func CurUnixTime() int64 {
	return time.Now().Unix()
}

func CurMillisecond() int64 {
	return time.Now().UnixNano() / 1e6
}

func CurHumanTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func GetCurNoSpaceTime() string {
	return time.Now().Format("20060102150405")
}

func FmtTimestampTime(sec int64) string {
	tm := time.Unix(sec, 0)
	return tm.Format("2006-01-02 15:04:05")
}

type BaseStruct struct {
	Id      string   `json:"id" bson:"_id"`
	CreateT *CurTime `json:"createT" bson:"createT"`
	UpdateT *CurTime `json:"updateT" bson:"updateT"`
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

func HttpPost(reqUrl string, data []byte, headers map[string]string) (*http.Response, error) {
	return httpRequest(http.MethodPost, reqUrl, data, headers)
}

func HttpPut(reqUrl string, data []byte, headers map[string]string) (*http.Response, error) {
	return httpRequest(http.MethodPut, reqUrl, data, headers)
}

func HttpDelete(reqUrl string, data []byte, headers map[string]string) (*http.Response, error) {
	return httpRequest(http.MethodDelete, reqUrl, data, headers)
}

func httpRequest(method, reqUrl string, data []byte, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest(method, reqUrl, bytes.NewBuffer(data))
	if err != nil {
		Logger.Errorf("", "[%s %s] 创建失败, %s", method, reqUrl, err.Error())
		return nil, err
	}

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	client := &http.Client{
		Timeout: HTTP_POST_TIMEOUT * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		Logger.Errorf("", "对 [%s %s] 发起 client.Do() 操作失败, %s", method, reqUrl, err.Error())
		return nil, err
	}

	return resp, nil
}

func HttpGet(reqUrl string, values map[string][]string, headers map[string]string) (*http.Response, error) {
	// url query paramaters
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
		Logger.Errorf("", "生成 http get newrequest 失败, %s", err.Error())
		return nil, err
	}
	req.URL.RawQuery = urlV.Encode()

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	client := &http.Client{
		Timeout: HTTP_GET_TIMEOUT * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		Logger.Errorf("", "发起get请求do[%s]失败, %s", reqUrl, err.Error())
		return nil, err
	}

	Logger.Debugf("", "HttGet Url: [%v]", reqUrl)
	return resp, nil
}

var MAX_ONE_PAGE_SIZE = 100

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

	if size > MAX_ONE_PAGE_SIZE {
		size = MAX_ONE_PAGE_SIZE
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

// 计算时间偏移,以目标日期的0时开始
func GetDayStartTimeByOffsetDay(base time.Time, offsetDay int) time.Time {
	t := base.AddDate(0, 0, offsetDay)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// 计算星期的偏移
func GetDayStartTimeByWeekdayOffset(base time.Time, weekday time.Weekday, offsetWeek int) time.Time {
	offsetDay := int(weekday-base.Weekday()) + (7 * offsetWeek)
	t := GetDayStartTimeByOffsetDay(base, offsetDay)
	return t
}

// 计算指定星期的指定小时的 time
func GetTimeByWeekdayOffsetAndHourOffset(base time.Time, weekday time.Weekday, offsetWeek, offsetHour int) time.Time {
	dayStart := GetDayStartTimeByWeekdayOffset(base, weekday, offsetWeek)
	dst := dayStart.Add(time.Duration(offsetHour) * time.Hour)
	return dst
}

// 转换字符串元到分
// 加 0.5 的解释见这里 https://stackoverflow.com/questions/46491966/golang-float-to-int-conversion
func ConvertStrYuanToIntFen(amt string) (int64, error) {
	f, err := strconv.ParseFloat(amt, 64)
	if err != nil {
		Logger.Errorf("", "转换元到分时，解析字符串[%s]到浮点数失败, %s", amt, err.Error())
		return -1, err
	}

	f = f * 100
	return int64(f + 0.5), nil
}

// 转换分到元字符串
func ConvertIntFenToStrYuan(fen int64) string {
	y := float64(fen) / 100.0
	s := strconv.FormatFloat(y, 'f', 2, 64)
	return s
}

// 转换元浮点数到分整数
func ConvertFloatYuanToIntFen(amt float64) int64 {
	amt = amt * 100
	return int64(amt + 0.5)
}

// 转换分整数到元浮点数
func ConvertIntFenToFloatYuan(amt int64) float64 {
	return float64(amt) / 100.0
}
