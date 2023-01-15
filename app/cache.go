// SPDX-License-Identifier: MIT

package app

import (
	"strings"
	"time"

	"github.com/issue9/web/cache"
	"github.com/issue9/web/cache/caches"
	"github.com/issue9/web/errs"
)

var cacheFactory = map[string]CacheBuilder{}

type CacheBuilder func(dsn string) (cache.Driver, error)

type cacheConfig struct {
	// 表示缓存的方式
	//
	// 该值可通过 [RegisterCache] 注册， 默认支持以下几种：
	//  - memory 以内存作为缓存；
	//  - memcached 以 memcached 作为缓存；
	//  - redis 以 redis 作为缓存；
	Type string `yaml:"type" json:"type" xml:"type,attr"`

	// 表示连接缓存服务器的参数
	//
	// 不同类型其参数是不同的，以下是对应的格式说明：
	//  - memory，此值为 [time.Duration] 格式的参数，用于表示执行回收的间隔；
	//  - memcached，则为服务器列表，多个服务器，以分号作为分隔；
	//  - redis，符合 [Redis URI scheme] 的字符串；
	//
	// [Redis URI scheme]: https://www.iana.org/assignments/uri-schemes/prov/redis
	DSN string `yaml:"dsn" json:"dsn" xml:"dsn"`
}

func (conf *configOf[T]) buildCache() *errs.ConfigError {
	if conf.Cache == nil {
		conf.cache = caches.NewMemory(time.Hour)
		return nil
	}

	b, found := cacheFactory[conf.Cache.Type]
	if !found {
		return errs.NewConfigError("type", errs.NewLocaleError("invalid value %s", conf.Cache.Type))
	}

	c, err := b(conf.Cache.DSN)
	if err != nil {
		return errs.NewConfigError("dsn", err)
	}
	conf.cache = c

	return nil
}

// RegisterCache 注册新的缓存方式
//
// name 为缓存的名称，如果存在同名，则会覆盖。
func RegisterCache(b CacheBuilder, name ...string) {
	if len(name) == 0 {
		panic("参数 name 不能为空")
	}

	for _, n := range name {
		cacheFactory[n] = b
	}
}

func init() {
	RegisterCache(func(dsn string) (cache.Driver, error) {
		d, err := time.ParseDuration(dsn)
		if err != nil {
			return nil, err
		}
		return caches.NewMemory(d), nil
	}, "", "memory")

	RegisterCache(func(dsn string) (cache.Driver, error) {
		return caches.NewMemcache(strings.Split(dsn, ";")...), nil
	}, "memcached", "memcache")

	RegisterCache(func(dsn string) (cache.Driver, error) {
		return caches.NewRedis(dsn)
	}, "redis")
}
