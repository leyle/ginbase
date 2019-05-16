package ginbase

import (
	"fmt"
	"github.com/go-redis/redis"
	. "github.com/leyle/gsimplelog"
)

type RedisOption struct {
	Host string
	Port string
	Passwd string
	DbNum int
}

func (o *RedisOption) Addr() string {
	return fmt.Sprintf("%s:%s", o.Host, o.Port)
}

func (o *RedisOption) String() string {
	return fmt.Sprintf("[%s:%s][%s],db[%d]", o.Host, o.Port, o.Passwd, o.DbNum)
}

func NewRedisClient(opt *RedisOption) (*redis.Client, error) {
	option := &redis.Options{
		Addr: opt.Addr(),
		Password: opt.Passwd,
		DB: opt.DbNum,
	}

	c := redis.NewClient(option)
	_, err := c.Ping().Result()
	if err != nil {
		Logger.Errorf("ping redis[%s]失败, %s", opt.String(), err.Error())
		return nil, err
	}

	Logger.Debugf("连接 redis[%s]成功", opt.String())
	return c, nil
}