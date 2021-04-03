package ratelimit

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gomodule/redigo/redis"
)

const (
	RedisReqPerMinPattern             string = "dynlimits_reqpermin_%s"
	RedisSlidingCountersWindowPattern string = "dynlimits_scw_%d_%s"
)

// RedisPoolConf contains the params to configure a redispool
// to be used as the source of ratelimit data
type RedisPoolConf struct {
	Address          string
	MaxIdleMinutes   int
	ConnectTimeoutMs int64
	ReadTimeoutMs    int64
}

// NewRedisPoolConf creates a new redis pool configuration
// with default values
func NewRedisPoolConf(address string) *RedisPoolConf {
	return &RedisPoolConf{
		Address:          address,
		MaxIdleMinutes:   300,
		ConnectTimeoutMs: 300,
		ReadTimeoutMs:    200,
	}
}

// NewRedisPool creates a redis pool to be used as
// the storage for rate limit data
func NewRedisPool(conf *RedisPoolConf) *redis.Pool {
	if conf == nil {
		// if no conf provided we default to localhost and default port
		conf = NewRedisPoolConf("localhost:6379")
	}
	return &redis.Pool{
		MaxIdle:     300,
		IdleTimeout: time.Minute,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", conf.Address, redis.DialDatabase(0),
				redis.DialConnectTimeout(300*time.Millisecond),
				redis.DialReadTimeout(200*time.Millisecond))
		},
		TestOnBorrow: func(c redis.Conn, ti time.Time) error {
			if time.Since(ti) < time.Minute {
				return nil
			}
			return fmt.Errorf("redis timeout")
		},
	}
}

// AdToRedisSlidingCountersWindow increments the counters for a
// a give second timestamp
func AddToRedisSlidingCountersWindow(conn redis.Conn, key string,
	timestampSec int64) error {

	min := timestampSec / 60
	secIdx := timestampSec % 60
	sliceName := fmt.Sprintf(RedisSlidingCountersWindowPattern, min, key)
	var err error
	if err = conn.Send("HINCRBY", sliceName, secIdx, 1); err != nil {
		return err
	}
	_, err = conn.Do("EXPIRE", sliceName, 3*60)
	return err
}

// SetRedisRateLimit updates the number of allowed requests per minute
// for a given key
func SetRedisRateLimit(conn redis.Conn, key string, reqPerMin int64) error {
	_, err := conn.Do("SET", fmt.Sprintf(RedisReqPerMinPattern, key), reqPerMin)
	return err
}

// SendSetRedisRateLimit adds the updated number of allowed requests per
// minute to the redis connection command buffer
func SendSetRedisRateLimit(conn redis.Conn, key string, reqPerMin int64) error {
	return conn.Send("SET", fmt.Sprintf(RedisReqPerMinPattern, key), reqPerMin)
}

// GetRedisSlidingCountersWindow returns an SlidingCountersWindow for
// a given Key
func GetRedisSlidingCountersWindow(conn redis.Conn, key string,
	timestampSec int64) (*SlidingCountersWindow, error) {
	min := timestampSec / 60
	curSecIdx := timestampSec % 60

	rrl := SlidingCountersWindow{
		ReqPerMin: 600,
	}

	rateLimitKey := fmt.Sprintf(RedisReqPerMinPattern, key)
	curSlice := fmt.Sprintf(RedisSlidingCountersWindowPattern, min, key)
	prevSlice := fmt.Sprintf(RedisSlidingCountersWindowPattern, min-1, key)

	// fetch current slice of seconds, and previous one, to
	// get our "custom" slice ending  at the current second
	var err error
	if err = conn.Send("MULTI"); err != nil {
		return nil, err
	}
	if err = conn.Send("GET", rateLimitKey); err != nil {
		return nil, err
	}
	if err = conn.Send("HGETALL", curSlice); err != nil {
		return nil, err
	}
	if err = conn.Send("HGETALL", prevSlice); err != nil {
		return nil, err
	}
	res, err := redis.Values(conn.Do("EXEC"))
	if err != nil {
		return nil, err
	}

	// read the limit (if not set the result would be nil)
	if res[0] == nil {
		// there is no rate limit for this api key + endpoint
		// TODO: define per configuration what to do if there
		// is no apikey entry rate limit.
		// For now, by default, is closed.
		return nil, fmt.Errorf("limit for api key and endpoint not found")
	}

	reqsPerMinBytes, ok := res[0].([]byte)
	if ok {
		reqPerMin, err := strconv.ParseInt(string(reqsPerMinBytes), 10, 64)
		if err == nil {
			rrl.ReqPerMin = reqPerMin
		} // else, means we have a weird format here ! who set this value !?
	}

	for idx, ires := range res[1:3] {
		keyvals, ok := ires.([]interface{})
		if !ok {
			return nil, fmt.Errorf("cannot get key val pair: %#v", ires)
		}
		if len(keyvals)%2 != 0 {
			return nil, fmt.Errorf("hmap keyval odd result")
		}

		numKeyVals := len(keyvals)
		for idxKVP := 0; idxKVP < numKeyVals; idxKVP += 2 {
			k, v, err := getHMapPair(keyvals, idxKVP)
			if err != nil {
				// TODO: log this error
				// fmt.Printf("err --- > %s\n", err.Error())
				continue
			}
			// if is the second slice, is the one from 60 secs in the past:
			k -= int64(60 * idx)
			/*
				Find the index in the window array.
				The last position (59) corresponds to the current second index
				[0 0 ... 0 curSecIdx]

				So, we need to substract the sliceSecond from curSecond,
				and offset it by 59
			*/
			secIdx := 59 - (int64(curSecIdx) - k)
			if secIdx >= 0 && secIdx < 60 {
				rrl.Window[secIdx] = v
				rrl.Sum += v
			}
		}
	}
	return &rrl, nil
}

// getHMapPair reads a couple of int64 from an slice of bytes
// read from redis, starting at idx offset.
func getHMapPair(keyvals []interface{}, idx int) (int64, int64, error) {
	keyBytes, ok := keyvals[idx].([]byte)
	if !ok {
		return 0, 0, fmt.Errorf("cannot read key bytes")
	}
	keyStr := string(keyBytes)
	key, err := strconv.ParseInt(keyStr, 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("cannot convert key to int %s", keyStr)
	}
	valBytes, ok := keyvals[idx+1].([]byte)
	if !ok {
		return 0, 0, fmt.Errorf("cannot read val bytes")
	}
	valStr := string(valBytes)
	val, err := strconv.ParseInt(valStr, 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("cannot convert val to int %s", valStr)
	}

	// fmt.Printf("k: %d, v: %d\n", key, val)
	return key, val, nil
}
