package configtemplate

import (
	"fmt"
	"strconv"
	"strings"
)

type ServerConf struct {
	Host string `json:"host" yaml:"host"`
	Port string `json:"host" yaml:"port"`
	Schema string `json:"host" yaml:"schema"`
	Domain string `json:"host" yaml:"domain"`
}

func (s *ServerConf) GetServerAddr() string {
	return fmt.Sprintf("%s:%s", s.Host, s.Port)
}

func (s *ServerConf) GetFullApi(path string) string {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return fmt.Sprintf("%s://%s%s", s.Schema, s.Domain, path)
}

type DbConf struct {
	Host string `json:"host" yaml:"host"`
	Port string `json:"port" yaml:"port"`
	User string `json:"user" yaml:"user"`
	Passwd string `json:"passwd" yaml:"passwd"`
	Database string `json:"database" yaml:"database"` // 如果是 redis 之类的，此处是 db number 的字符串形式
}

func (d *DbConf) GetMongodbURI() string {
	// mongodb://myuser:mypass@localhost:40001,otherhost:40001/mydb
	url := fmt.Sprintf("mongodb://%s:%s@%s:%s/%s?connect=direct", d.User, d.Passwd, d.Host, d.Port, d.Database)
	return url
}

func (d *DbConf) GetRedisDbNum() int {
	dbn, err := strconv.Atoi(d.Database)
	if err != nil {
		fmt.Println(err.Error())
		panic(err)
	}
	return dbn
}

