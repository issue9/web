// SPDX-License-Identifier: MIT

package caches

import (
	"errors"
	"strconv"

	"github.com/gomodule/redigo/redis"

	"github.com/issue9/web/cache"
)

type redisDriver struct {
	conn redis.Conn
}

type redisCounter struct {
	driver    *redisDriver
	key       string
	val       string
	originVal uint64
	ttl       int
}

// NewRedis 返回 redis 的缓存实现
func NewRedis(url string, o ...redis.DialOption) (cache.Driver, error) {
	c, err := redis.DialURL(url, o...)
	if err != nil {
		return nil, err
	}
	return &redisDriver{conn: c}, nil
}

func (d *redisDriver) Get(key string, val any) error {
	bs, err := redis.Bytes(d.conn.Do("GET", key))
	if errors.Is(err, redis.ErrNil) {
		return cache.ErrCacheMiss()
	} else if err != nil {
		return err
	}

	return Unmarshal(bs, val)
}

func (d *redisDriver) Set(key string, val any, seconds int) error {
	bs, err := Marshal(val)
	if err != nil {
		return err
	}

	if seconds == 0 {
		_, err = d.conn.Do("SET", key, string(bs))
		return err
	}

	_, err = d.conn.Do("SET", key, string(bs), "EX", seconds)
	return err
}

func (d *redisDriver) Delete(key string) error {
	_, err := d.conn.Do("DEL", key)
	return err
}

func (d *redisDriver) Exists(key string) bool {
	exists, _ := redis.Bool(d.conn.Do("EXISTS", key))
	return exists
}

func (d *redisDriver) Clean() error {
	_, err := d.conn.Do("FLUSHDB")
	return err
}

func (d *redisDriver) Close() error { return d.conn.Close() }

func (d *redisDriver) Counter(key string, val uint64, ttl int) cache.Counter {
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
	return redis.Uint64(c.driver.conn.Do("INCRBY", c.key, n))
}

func (c *redisCounter) Decr(n uint64) (uint64, error) {
	if err := c.init(); err != nil {
		return 0, err
	}
	v, err := redis.Int64(c.driver.conn.Do("DECRBY", c.key, n))
	if err != nil {
		return 0, err
	}
	if v < 0 {
		_, err = c.driver.conn.Do("INCRBY", c.key, n)
		return 0, err

	}
	return uint64(v), nil
}

func (c *redisCounter) init() error {
	_, err := c.driver.conn.Do("SET", c.key, c.val, "EX", c.ttl, "NX")
	return err
}

func (c *redisCounter) Value() (uint64, error) {
	s, err := redis.String(c.driver.conn.Do("GET", c.key))
	if errors.Is(err, redis.ErrNil) {
		return c.originVal, cache.ErrCacheMiss()
	} else if err != nil {
		return c.originVal, err
	}
	return strconv.ParseUint(s, 10, 64)
}
