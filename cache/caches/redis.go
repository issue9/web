// SPDX-License-Identifier: MIT

package caches

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/issue9/web/cache"
)

type redisDriver struct {
	conn *redis.Client
}

type redisCounter struct {
	driver    *redisDriver
	key       string
	val       string
	originVal uint64
	ttl       time.Duration
}

// redis 处理 DECRBY 的事务脚本
const redisDecrByScript = `local cnt = redis.call('DECRBY', KEYS[1], ARGV[1])
if cnt < 0 then
    redis.call('SET', KEYS[1], '0')
end
return (cnt < 0 and 0 or cnt)`

// NewRedisFromURL 声明基于 redis 的缓存系统
//
// url 为符合 [Redis URI scheme] 的字符串
//
// [Redis URI scheme]: https://www.iana.org/assignments/uri-schemes/prov/redis
func NewRedisFromURL(url string) (cache.Driver, error) {
	opt, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}
	return NewRedis(redis.NewClient(opt)), nil
}

// NewRedis 声明基于 redis 的缓存系统
func NewRedis(c *redis.Client) cache.Driver { return &redisDriver{conn: c} }

func (d *redisDriver) Get(key string, val any) error {
	bs, err := d.conn.Get(context.Background(), key).Bytes()
	if errors.Is(err, redis.Nil) {
		return cache.ErrCacheMiss()
	} else if err != nil {
		return err
	}

	return Unmarshal(bs, val)
}

func (d *redisDriver) Set(key string, val any, ttl time.Duration) error {
	bs, err := Marshal(val)
	if err != nil {
		return err
	}
	return d.conn.Set(context.Background(), key, bs, ttl).Err()
}

func (d *redisDriver) Delete(key string) error {
	return d.conn.Del(context.Background(), key).Err()
}

func (d *redisDriver) Exists(key string) bool {
	rslt, err := d.conn.Exists(context.Background(), key).Result()
	return err == nil && rslt > 0
}

func (d *redisDriver) Clean() error {
	return d.conn.FlushDB(context.Background()).Err()
}

func (d *redisDriver) Close() error { return d.conn.Close() }

func (d *redisDriver) Counter(key string, val uint64, ttl time.Duration) cache.Counter {
	return &redisCounter{
		driver:    d,
		key:       key,
		val:       strconv.FormatUint(val, 10),
		originVal: val,
		ttl:       ttl,
	}
}

func (c *redisCounter) Incr(n uint64) (uint64, error) {
	if err := c.init(); err != nil {
		return 0, err
	}

	rslt, err := c.driver.conn.IncrBy(context.Background(), c.key, int64(n)).Result()
	if err != nil {
		return 0, err
	}
	return uint64(rslt), nil
}

func (c *redisCounter) Decr(n uint64) (uint64, error) {
	if err := c.init(); err != nil {
		return 0, err
	}

	in := int64(n)
	v, err := c.driver.conn.Eval(context.Background(), redisDecrByScript, []string{c.key}, in).Int64()
	return uint64(v), err
}

func (c *redisCounter) init() error {
	return c.driver.conn.SetNX(context.Background(), c.key, c.val, c.ttl).Err()
}

func (c *redisCounter) Value() (uint64, error) {
	s, err := c.driver.conn.Get(context.Background(), c.key).Result()
	if errors.Is(err, redis.Nil) {
		return c.originVal, cache.ErrCacheMiss()
	} else if err != nil {
		return c.originVal, err
	}
	return strconv.ParseUint(s, 10, 64)
}

func (c *redisCounter) Delete() error {
	return c.driver.conn.Del(context.Background(), c.key).Err()
}
