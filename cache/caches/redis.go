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

// NewRedis 返回 redis 的缓存实现
func NewRedis(url string) (cache.Driver, error) {
	c, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}
	return &redisDriver{conn: redis.NewClient(c)}, nil
}

func (d *redisDriver) Get(key string, val any) error {
	bs, err := d.conn.Get(context.Background(), key).Bytes()
	if errors.Is(err, redis.Nil) {
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
	return d.conn.Set(context.Background(), key, bs, time.Duration(seconds)*time.Second).Err()
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

func (d *redisDriver) Counter(key string, val uint64, ttl int) cache.Counter {
	return &redisCounter{
		driver:    d,
		key:       key,
		val:       strconv.FormatUint(val, 10),
		originVal: val,
		ttl:       time.Duration(ttl) * time.Second,
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
	v, err := c.driver.conn.DecrBy(context.Background(), c.key, in).Result()
	if err != nil {
		return 0, err
	}

	if v < 0 {
		_, err = c.driver.conn.IncrBy(context.Background(), c.key, in).Result()
		return 0, err

	}
	return uint64(v), nil
}

func (c *redisCounter) init() error {
	cmd := c.driver.conn.SetNX(context.Background(), c.key, c.val, time.Duration(c.ttl))
	return cmd.Err()
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
