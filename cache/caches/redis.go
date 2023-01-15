// SPDX-License-Identifier: MIT

package caches

import (
	"errors"

	"github.com/gomodule/redigo/redis"

	"github.com/issue9/web/cache"
)

type redisDriver struct {
	conn redis.Conn
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

	return cache.Unmarshal(bs, val)
}

func (d *redisDriver) Set(key string, val any, seconds int) error {
	bs, err := cache.Marshal(val)
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
	v, _ := d.conn.Do("GET", key)
	return v != nil
}

func (d *redisDriver) Clean() error {
	_, err := d.conn.Do("FLUSHDB")
	return err
}

func (d *redisDriver) Close() error { return d.conn.Close() }
