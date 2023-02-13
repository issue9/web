// SPDX-License-Identifier: MIT

package caches

import (
	"errors"
	"strconv"
	"time"

	"github.com/bradfitz/gomemcache/memcache"

	"github.com/issue9/web/cache"
)

type memcacheDriver struct {
	client *memcache.Client
}

type memcacheCounter struct {
	driver    *memcacheDriver
	key       string
	val       []byte
	originVal uint64
	ttl       int32
}

// NewMemcache 声明基于 memcached 的缓存系统
func NewMemcache(addr ...string) cache.Driver {
	return &memcacheDriver{
		client: memcache.New(addr...),
	}
}

func (d *memcacheDriver) Get(key string, val any) error {
	item, err := d.client.Get(key)
	if errors.Is(err, memcache.ErrCacheMiss) {
		return cache.ErrCacheMiss()
	} else if err != nil {
		return err
	}
	return Unmarshal(item.Value, val)
}

func (d *memcacheDriver) Set(key string, val any, ttl time.Duration) error {
	bs, err := Marshal(val)
	if err != nil {
		return err
	}

	return d.client.Set(&memcache.Item{
		Key:        key,
		Value:      bs,
		Expiration: int32(ttl.Seconds()),
	})
}

func (d *memcacheDriver) Delete(key string) error { return d.client.Delete(key) }

func (d *memcacheDriver) Exists(key string) bool {
	_, err := d.client.Get(key)
	return err == nil || !errors.Is(err, memcache.ErrCacheMiss)
}

func (d *memcacheDriver) Clean() error { return d.client.DeleteAll() }

func (d *memcacheDriver) Close() error { return d.client.Close() }

func (d *memcacheDriver) Counter(key string, val uint64, ttl time.Duration) cache.Counter {
	return &memcacheCounter{
		driver:    d,
		key:       key,
		val:       []byte(strconv.FormatUint(val, 10)),
		originVal: val,
		ttl:       int32(ttl.Seconds()),
	}
}

func (c *memcacheCounter) Incr(n uint64) (uint64, error) {
	if err := c.init(); err != nil {
		return 0, err
	}

	v, err := c.driver.client.Increment(c.key, n)
	if err == nil {
		err = c.driver.client.Touch(c.key, c.ttl)
	}

	if err != nil {
		return 0, err
	}
	return v, nil
}

func (c *memcacheCounter) Decr(n uint64) (uint64, error) {
	if err := c.init(); err != nil {
		return 0, err
	}

	v, err := c.driver.client.Decrement(c.key, n)
	if err == nil {
		err = c.driver.client.Touch(c.key, c.ttl)
	}

	if err != nil {
		return 0, err
	}
	return v, nil
}

func (c *memcacheCounter) init() error {
	err := c.driver.client.Add(&memcache.Item{
		Key:        c.key,
		Value:      c.val,
		Expiration: c.ttl,
	})
	if errors.Is(err, memcache.ErrNotStored) {
		return nil
	}
	return err
}

func (c *memcacheCounter) Value() (uint64, error) {
	item, err := c.driver.client.Get(c.key)
	if errors.Is(err, memcache.ErrCacheMiss) {
		return c.originVal, cache.ErrCacheMiss()
	} else if err != nil {
		return c.originVal, err
	}

	v := string(item.Value)
	if v == "0 " { // 零值?
		return 0, nil
	}
	return strconv.ParseUint(v, 10, 64)
}

func (c *memcacheCounter) Delete() error {
	err := c.driver.client.Delete(c.key)
	if errors.Is(err, memcache.ErrCacheMiss) {
		return nil
	}
	return err
}
