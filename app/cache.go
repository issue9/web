// SPDX-License-Identifier: MIT

package app

import (
	"strings"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/gomodule/redigo/redis"
	"github.com/issue9/cache/file"
	cm "github.com/issue9/cache/memcache"
	"github.com/issue9/cache/memory"
	cr "github.com/issue9/cache/redis"
)

// 缓存的相关配置
type cacheConfig struct {
	// 表示缓存的方式
	//
	// 目前支持以下几种试：
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

func (conf *webconfig[T]) buildCache() *Error {
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
	case "file":
		args := strings.SplitN(conf.Cache.DSN, ";", 2)
		if len(args) != 2 {
			return &Error{Field: "dsn", Message: "必须指定 path 和 gc 两个参数"}
		}

		gc, err := time.ParseDuration(args[1])
		if err != nil {
			return &Error{Field: "dsn", Message: err.Error()}
		}

		conf.cache = file.New(args[0], gc, conf.logs.ERROR())
	default:
		return &Error{Field: "type", Message: "无效的值"}
	}
	return nil
}
