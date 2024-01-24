// SPDX-License-Identifier: MIT

package app

import (
	"strconv"
	"strings"
	"time"

	"github.com/issue9/mux/v7/group"

	"github.com/issue9/web"
	"github.com/issue9/web/locales"
	"github.com/issue9/web/server/registry"
)

var (
	strategyFactory      = newRegister[StrategyBuilder]()
	typeFactory          = newRegister[RegistryTypeBuilder]()
	routerMatcherFactory = newRegister[RouterMatcherBuilder]()
)

type (
	// StrategyBuilder 构建负载均衡算法的函数
	StrategyBuilder = func() *registry.Strategy

	// RegistryTypeBuilder 生成 [registry.Registry] 的方法
	//
	// 第一个参数为 [web.Cache] 类型，在非 cache 时，该类型应该是无用的。
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
	}
)

func RegisterStrategy(f StrategyBuilder, name string) { strategyFactory.register(f, name) }

func RegisterRegistryType(f RegistryTypeBuilder, name string) { typeFactory.register(f, name) }

func RegisterRouterMatcher(f RouterMatcherBuilder, name string) {
	routerMatcherFactory.register(f, name)
}

func (conf *configOf[T]) buildMicro(c web.Cache) *web.FieldError {
	if conf.Registry != nil {
		if err := conf.Registry.build(conf.cache); err != nil {
			return err
		}
	}

	if conf.Peer != "" {
		conf.peer = conf.Registry.s().NewPeer()
		if err := conf.peer.UnmarshalText([]byte(conf.Peer)); err != nil {
			return web.NewFieldError("peer", err)
		}
	}

	if len(conf.Mappers) > 0 {
		conf.mapper = registry.Mapper{}
		for i, m := range conf.Mappers {
			mm, found := routerMatcherFactory.get(m.Matcher)
			if !found {
				return web.NewFieldError("mappers["+strconv.Itoa(i)+"].matcher", locales.ErrNotFound(m.Matcher))
			}
			conf.mapper[m.Name] = mm()
		}
	}

	return nil
}

func (r *registryConfig) build(c web.Cache) *web.FieldError {
	t, found := typeFactory.get(r.Type)
	if !found {
		return web.NewFieldError("type", locales.ErrNotFound(r.Type))
	}

	s, found := strategyFactory.get(r.Strategy)
	if !found {
		return web.NewFieldError("strategy", locales.ErrNotFound(r.Strategy))
	}
	r.s = s

	r.registry = t(c, s(), strings.Split(r.Args, ",")...)

	return nil
}

func init() {
	RegisterStrategy(registry.NewRandomStrategy, "random")
	RegisterStrategy(registry.NewRoundRobinStrategy, "round-robin")
	RegisterStrategy(registry.NewWeightedRandomStrategy, "weighted-random")
	RegisterStrategy(registry.NewWeightedRoundRobinStrategy, "weighted-round-robin")

	RegisterRegistryType(func(c web.Cache, s *registry.Strategy, arg ...string) registry.Registry {
		if c == nil {
			panic("参数 o.Cache 不能为空")
		}

		if len(arg) != 1 {
			panic("参数 arg 数量必须为 1")
		}
		freq, err := time.ParseDuration(arg[0])
		if err != nil {
			panic(err)
		}

		return registry.NewCache(c, s, freq)
	}, "cache")

	RegisterRouterMatcher(func(arg ...string) web.RouterMatcher { return group.NewHosts(false, arg...) }, "hosts")
	RegisterRouterMatcher(func(arg ...string) web.RouterMatcher { return group.NewPathVersion(arg[0], arg[1:]...) }, "prefix")
	RegisterRouterMatcher(func(arg ...string) web.RouterMatcher { return group.NewHeaderVersion(arg[0], arg[0], nil, arg[2:]...) }, "version")
	RegisterRouterMatcher(func(arg ...string) web.RouterMatcher { return nil }, "any")
}
