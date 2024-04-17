// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package config

import (
	"strconv"
	"strings"

	"github.com/issue9/mux/v8/group"

	"github.com/issue9/web"
	"github.com/issue9/web/locales"
	"github.com/issue9/web/server/registry"
)

type (
	// StrategyBuilder 构建负载均衡算法的函数
	StrategyBuilder = func() *registry.Strategy

	// RegistryTypeBuilder 生成 [registry.Registry] 的方法
	//
	// 第一个参数为 [web.Cache] 类型，如果返回的 [registry.Registry] 为非 cache 时则第一个参数可以为 nil。
	RegistryTypeBuilder = func(web.Cache, *registry.Strategy, ...string) registry.Registry

	RouterMatcherBuilder = func(...string) web.RouterMatcher

	// registryConfig 注册服务中心的配置项
	registryConfig struct {
		// 配置的保存类型
		//
		// 该类型可通过 [RegisterRegistryType] 进行注册，默认情况下支持以下类型：
		//  - cache 以缓存系统作为储存类型；
		Type string `json:"type" yaml:"type" xml:"type"`

		// 负载均衡的方案
		//
		// 可通过 [RegisterStrategy] 进行注册，默认情况下支持以下类型：
		//  - random 随机；
		//  - weighted-random 带权重的随机；
		//  - round-robin 轮循；
		//  - weighted-round-robin 带权重的轮循；
		Strategy string `json:"strategy" yaml:"strategy" xml:"strategy"`
		s        StrategyBuilder

		// 传递 Type 的额外参数
		//
		// 会根据 args 的不同而不同：
		//  - cache 仅支持一个参数，为 [time.ParseDuration] 可解析的字符串；
		Args string `json:"args,omitempty" yaml:"args,omitempty" xml:"args>arg,omitempty"`

		registry registry.Registry
	}

	mapperConfig struct {
		// 微服务名称
		Name string `json:"name" yaml:"name" xml:"name"`

		// 判断某个请求是否进入当前微服务的方法
		//
		// 该值可通过 [RegisterRouterMatcher] 注册，默认情况下支持以下类型：
		//  - hosts 只限定域名；
		//  - prefix 包含特定前缀的访问地址；
		//  - version 在 accept 中指定的特定的版本号才行；
		//  - any 任意；
		Matcher string `json:"matcher" yaml:"matcher" xml:"matcher"`

		// 传递 Matcher 的额外参数
		//
		// 会根据 Matcher 的不同而不同：
		//  - hosts 以逗号分隔的域名列表；
		//  - prefix 以逗号分隔的 URL 前缀列表；
		//  - version 允许放行的版本号列表(以逗号分隔)，这些版本号出现在 accept 报头；
		//  - any 不需要参数；
		Args string `json:"args,omitempty" yaml:"args,omitempty" xml:"args>arg,omitempty"`
	}
)

func (conf *configOf[T]) buildMicro(c web.Cache) *web.FieldError {
	if conf.Registry != nil {
		if err := conf.Registry.build(c); err != nil {
			return err.AddFieldParent("registry")
		}
	}

	if conf.Peer != "" {
		conf.peer = conf.Registry.s().NewPeer()
		if err := conf.peer.UnmarshalText([]byte(conf.Peer)); err != nil {
			return web.NewFieldError("peer", err)
		}
	}

	if len(conf.Mappers) > 0 {
		conf.mapper = make(map[string]group.Matcher, len(conf.Mappers))
		for i, m := range conf.Mappers {
			mm, found := routerMatcherFactory.get(m.Matcher)
			if !found {
				return web.NewFieldError("mappers["+strconv.Itoa(i)+"].matcher", locales.ErrNotFound())
			}
			conf.mapper[m.Name] = mm(strings.Split(m.Args, ",")...)
		}
	}

	return nil
}

func (r *registryConfig) build(c web.Cache) *web.FieldError {
	t, found := registryTypeFactory.get(r.Type)
	if !found {
		return web.NewFieldError("type", locales.ErrNotFound())
	}

	s, found := strategyFactory.get(r.Strategy)
	if !found {
		return web.NewFieldError("strategy", locales.ErrNotFound())
	}
	r.s = s

	r.registry = t(c, s(), strings.Split(r.Args, ",")...)

	return nil
}
