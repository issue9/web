// SPDX-License-Identifier: MIT

package app

import (
	"errors"
	"log"
	"strings"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/gomodule/redigo/redis"
	"github.com/issue9/cache/file"
	cm "github.com/issue9/cache/memcache"
	"github.com/issue9/cache/memory"
	cr "github.com/issue9/cache/redis"

	"github.com/issue9/web/server"
)

var cacheFactory = map[string]CacheBuilder{}

type CacheBuilder func(dsn string) (server.Cache, error)

// 缓存的相关配置
type cacheConfig struct {
	// 表示缓存的方式
	//
	// 该值可通过 RegisterCache 注册， 默认支持以下几种：
	// - memory 以内存作为缓存；
	// - memcached 以 memcached 作为缓存；
	// - redis 以 redis 作为缓存；
	// - file 以文件作为缓存；
	Type string `yaml:"type" json:"type" xml:"type,attr"`

	// 表示连接缓存服务器的参数
	//
	// - memory，此值为 time.Duration 格式的参数，用于表示执行回收的间隔；
	// - memcached，则为服务器列表，多个服务器，以分号作为分隔；
	// - redis，则为符合 Redis URI scheme 的字符串，可参考 https://www.iana.org/assignments/uri-schemes/prov/redis；
	// - file，表示以半有分号分隔的参数列表，可以指定以下两个参数：
	//  - path 文件路径；
	//  - gc 执行回收的间隔，time.Duration 格式；
	DSN string `yaml:"dsn" json:"dsn" xml:"dsn"`
}

func (conf *configOf[T]) buildCache() *ConfigError {
	if conf.Cache == nil {
		conf.cache = memory.New(time.Hour)
		return nil
	}

	b, found := cacheFactory[conf.Cache.Type]
	if !found {
		return &ConfigError{Field: "type", Message: "无效的值"}
	}

	c, err := b(conf.Cache.DSN)
	if err != nil {
		return &ConfigError{Field: "dsn", Message: err}
	}
	conf.cache = c

	return nil
}

// RegisterCache 注册新的缓存方式
func RegisterCache(b CacheBuilder, name ...string) {
	if len(name) == 0 {
		panic("参数 name 不能为空")
	}

	for _, n := range name {
		cacheFactory[n] = b
	}
}

func init() {
	RegisterCache(func(dsn string) (server.Cache, error) {
		d, err := time.ParseDuration(dsn)
		if err != nil {
			return nil, err
		}
		return memory.New(d), nil
	}, "", "memory")

	RegisterCache(func(dsn string) (server.Cache, error) {
		return cm.New(memcache.New(strings.Split(dsn, ";")...)), nil
	}, "memcached", "memcache")

	RegisterCache(func(dsn string) (server.Cache, error) {
		c, err := redis.DialURL(dsn)
		if err != nil {
			return nil, err
		}
		return cr.New(c), nil
	}, "redis")

	RegisterCache(func(dsn string) (server.Cache, error) {
		args := strings.SplitN(dsn, ";", 2)
		if len(args) != 2 {
			return nil, errors.New("必须指定 path 和 gc 两个参数")
		}

		gc, err := time.ParseDuration(args[1])
		if err != nil {
			return nil, err
		}

		return file.New(args[0], gc, log.Default()), nil
	}, "file")
}
