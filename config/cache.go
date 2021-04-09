// SPDX-License-Identifier: MIT

package config

import (
	"strings"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/gomodule/redigo/redis"
	cm "github.com/issue9/cache/memcache"
	"github.com/issue9/cache/memory"
	cr "github.com/issue9/cache/redis"
)

// Cache 缓存的相关配置
type Cache struct {
	// 表示缓存的方式
	//
	// 目前支持以下几种试：
	// - memory 以内存作为缓存；
	// - memcached 以 memcached 作为缓存；
	// - redis 以 redis 作为缓存；
	Type string `yaml:"type" json:"type" xml:"type,attr"`

	// 表示连接缓存服务器的参数
	//
	// 如果类型为 memory，此值为 time.Duration 格式的参数，用于表示执行回收的间隔；
	// 如果为 memcached，则为服务器列表，多个服务器，以分号作为分隔；
	// 如果为 redis，则为符合 Redis URI scheme 的字符串，可参考 https://www.iana.org/assignments/uri-schemes/prov/redis
	DSN string `yaml:"dsn" json:"dsn" xml:"dsn"`
}

func (conf *Webconfig) buildCache() *Error {
	if conf.Cache == nil {
		conf.cache = memory.New(time.Hour)
		return nil
	}

	switch conf.Cache.Type {
	case "memory", "":
		d, err := time.ParseDuration(conf.Cache.DSN)
		if err != nil {
			return &Error{Field: "dsn", Message: err.Error()}
		}
		conf.cache = memory.New(d)
	case "memcached", "memcache":
		c := memcache.New(strings.Split(conf.Cache.DSN, ";")...)
		conf.cache = cm.New(c)
	case "redis":
		c, err := redis.DialURL(conf.Cache.DSN)
		if err != nil {
			return &Error{Field: "dsn", Message: err.Error()}
		}
		conf.cache = cr.New(c)
	default:
		return &Error{Field: "type", Message: "无效的值"}
	}
	return nil
}
