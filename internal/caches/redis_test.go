// SPDX-License-Identifier: MIT

package caches

import (
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/issue9/assert/v3"

	"github.com/issue9/web/cache"
)

var _ cache.Cache = &redisDriver{}

var redisOptions = []redis.DialOption{
	redis.DialConnectTimeout(time.Second),
	redis.DialReadTimeout(time.Second),
	redis.DialWriteTimeout(time.Second),
}

const redisURL = "redis://localhost:6379"

func TestRedis(t *testing.T) {
	a := assert.New(t, false)

	c, err := NewRedis(redisURL, redisOptions...)
	a.NotError(err).NotNil(c)

	testCache(a, c)
	testObject(a, c)

	a.NotError(c.Close())
}

func TestRedis_Close(t *testing.T) {
	a := assert.New(t, false)

	c, err := NewRedis(redisURL, redisOptions...)
	a.NotError(err).NotNil(c)
	a.NotError(c.Set("key", "val", cache.Forever))
	a.NotError(c.Close())

	c, err = NewRedis(redisURL, redisOptions...)
	a.NotError(err).NotNil(c)
	var val string
	a.NotError(c.Get("key", &val)).Equal(val, "val")
}
