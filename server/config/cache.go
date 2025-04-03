// SPDX-FileCopyrightText: 2018-2025 caixw
//
// SPDX-License-Identifier: MIT

package config

import (
	"github.com/issue9/cache"

	"github.com/issue9/web"
	"github.com/issue9/web/locales"
)

// CacheBuilder 构建缓存客户端的方法
type CacheBuilder = func(dsn string) (cache.Driver, error)

type cacheConfig struct {
	// 表示缓存的方式
	//
	// 该值可通过 [RegisterCache] 注册， 默认支持以下几种：
	//  - memory 以内存作为缓存；
	//  - memcached 以 memcached 作为缓存；
	//  - redis 以 redis 作为缓存；
	Type string `yaml:"type" json:"type" xml:"type,attr" toml:"type"`

	// 表示连接缓存服务器的参数
	//
	// 不同类型其参数是不同的，以下是对应的格式说明：
	//  - memory: 不需要参数；
	//  - memcached: 则为服务器列表，多个服务器，以分号作为分隔；
	//  - redis: 符合 [Redis URI scheme] 的字符串；
	//
	// [Redis URI scheme]: https://www.iana.org/assignments/uri-schemes/prov/redis
	DSN string `yaml:"dsn" json:"dsn" xml:"dsn" toml:"dsn"`
}

func (conf *configOf[T]) buildCache() *web.FieldError {
	if conf.Cache == nil {
		return nil
	}

	b, found := cacheFactory.get(conf.Cache.Type)
	if !found {
		return web.NewFieldError("type", locales.InvalidValue)
	}

	drv, err := b(conf.Cache.DSN)
	if err != nil {
		return web.NewFieldError("dsn", err)
	}
	conf.cache = drv

	return nil
}
