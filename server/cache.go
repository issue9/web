// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package server

import (
	"time"

	"github.com/issue9/cache"
	"github.com/issue9/scheduled"

	"github.com/issue9/web"
	"github.com/issue9/web/locales"
)

// CacheBuilder 构建缓存客户端的方法
//
// drv 为缓存客户端对象；
// 如果是服务端客户端一体的，可通过 job 来指定服务端的定时回收任务；
type CacheBuilder = func(dsn string) (drv cache.Driver, job *Job, err error)

type Job struct {
	Ticker time.Duration
	Job    scheduled.JobFunc
}

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

func (conf *configOf[T]) buildCache() *web.FieldError {
	if conf.Cache == nil {
		return nil
	}

	b, found := cacheFactory.get(conf.Cache.Type)
	if !found {
		return web.NewFieldError("type", locales.InvalidValue)
	}

	drv, job, err := b(conf.Cache.DSN)
	if err != nil {
		return web.NewFieldError("dsn", err)
	}
	conf.cache = drv
	if job != nil {
		conf.init = append(conf.init, func(s web.Server) {
			s.Services().AddTicker(locales.RecycleLocalCache, job.Job, job.Ticker, false, false)
		})
	}

	return nil
}
