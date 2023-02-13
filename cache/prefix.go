// SPDX-License-Identifier: MIT

package cache

import "time"

type prefix struct {
	prefix string
	cache  Cache
}

// Prefix 生成一个带有统一前缀名称的缓存访问对象
//
//	c := NewMemory(...)
//	p := cache.Prefix(c, "prefix_")
//	p.Get("k1") // 相当于 c.Get("prefix_k1")
func Prefix(a Cache, p string) Cache { return &prefix{prefix: p, cache: a} }

func (p *prefix) Get(key string, v any) error {
	return p.cache.Get(p.prefix+key, v)
}

func (p *prefix) Set(key string, val any, seconds time.Duration) error {
	return p.cache.Set(p.prefix+key, val, seconds)
}

func (p *prefix) Delete(key string) error {
	return p.cache.Delete(p.prefix + key)
}

func (p *prefix) Exists(key string) bool {
	return p.cache.Exists(p.prefix + key)
}

func (p *prefix) Counter(key string, val uint64, ttl time.Duration) Counter {
	return p.cache.Counter(p.prefix+key, val, ttl)
}
