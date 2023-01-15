// SPDX-License-Identifier: MIT

package caches

import (
	"errors"

	"github.com/bradfitz/gomemcache/memcache"

	"github.com/issue9/web/cache"
)

type memcacheDriver struct {
	client *memcache.Client
}

// NewFromServers 声明一个新的 Memcache 实例
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
	return cache.Unmarshal(item.Value, val)
}

func (d *memcacheDriver) Set(key string, val any, seconds int) error {
	bs, err := cache.Marshal(val)
	if err != nil {
		return err
	}

	return d.client.Set(&memcache.Item{
		Key:        key,
		Value:      bs,
		Expiration: int32(seconds),
	})
}

func (d *memcacheDriver) Delete(key string) error {
	return d.client.Delete(key)
}

func (d *memcacheDriver) Exists(key string) bool {
	_, err := d.client.Get(key)
	return err == nil || !errors.Is(err, memcache.ErrCacheMiss)
}

func (d *memcacheDriver) Clean() error { return d.client.DeleteAll() }

func (d *memcacheDriver) Close() error {
	d.client = nil
	return nil
}
