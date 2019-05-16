package ginbase

import (
	"github.com/go-redis/redis"
	"time"
	. "github.com/leyle/gsimplelog"
)

const (
	DEFAULT_LOCK_ACQUIRE_TIMEOUT = 5 // 秒
	DEFAULT_LOCK_KEY_TIMEOUT = 5
)

// acquire lock
// 返回 true 的时候，就会同步返回一个 set 的 val
// 这个 val 可以作为后续 del key 的时候凭证
// timeout 都是秒
func AcquireLock(r *redis.Client, resource string, acquireTimeout, lockTimeout int) (string, bool) {
	if acquireTimeout <= 0 {
		acquireTimeout = DEFAULT_LOCK_ACQUIRE_TIMEOUT
	}
	if lockTimeout <= 0 {
		lockTimeout = DEFAULT_LOCK_KEY_TIMEOUT
	}

	val := GenerateDataId()
	lockTimeoutD := time.Duration(lockTimeout) * time.Second
	endTime := time.Now().Add(time.Duration(acquireTimeout) * time.Second)
	for time.Now().Unix() < endTime.Unix() {
		ok, err := r.SetNX(resource, val, lockTimeoutD).Result()
		if err != nil {
			Logger.Errorf("设置[%s]的锁失败, %s", resource, err.Error())
			return "", false
		}

		if ok {
			return val, true
		} else {
			time.Sleep(10 * time.Millisecond)
			continue
		}
	}
	return "", false
}

// release lock
func ReleaseLock(r *redis.Client, resource, val string) bool {
	v, err := r.Get(resource).Result()
	if err != nil && err != redis.Nil {
		Logger.Errorf("释放[%s]的锁失败, %s", err.Error())
		return false
	}

	if err == redis.Nil {
		return true
	}

	if v == val {
		r.Del(resource)
		return true
	} else {
		// 数据已被其他人加锁，那么此处可以认为是 ok 的
		return true
	}
}