// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package config

import (
	"compress/flate"
	"compress/gzip"
	"compress/lzw"
	"encoding/xml"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/andybalholm/brotli"
	"github.com/goccy/go-yaml"
	"github.com/issue9/cache"
	"github.com/issue9/cache/drivers/memcache"
	"github.com/issue9/cache/drivers/memory"
	"github.com/issue9/cache/drivers/redis"
	"github.com/issue9/config"
	"github.com/issue9/mux/v10"
	"github.com/issue9/unique/v2"

	"github.com/issue9/web"
	"github.com/issue9/web/negotiation"
	"github.com/issue9/web/negotiation/form"
	"github.com/issue9/web/server"
	"github.com/issue9/web/server/registry"
)

type register[T any] struct {
	items map[string]T
}

func newRegister[T any]() *register[T] { return &register[T]{items: make(map[string]T, 5)} }

func (r *register[T]) register(v T, name ...string) {
	if len(name) == 0 {
		panic("必须指定至少一个 name 参数")
	}

	if i := slices.Index(name, ""); i >= 0 {
		panic("参数 name 中不能包含空字符串")
	}

	for _, n := range name {
		if _, found := r.items[n]; found {
			panic("存在同名的媒体类型:" + n)
		}
		r.items[n] = v
	}
}

func (r *register[T]) get(name string) (T, bool) {
	v, found := r.items[name]
	return v, found
}

// 以下为所有 register 的实例化类型及关联的操作

var (
	logHandlersFactory    = newRegister[LogsHandlerBuilder]()
	cacheFactory          = newRegister[CacheBuilder]()
	compressorFactory     = newRegister[web.Compressor]()
	idGeneratorFactory    = newRegister[IDGeneratorBuilder]()
	codecFactory          = newRegister[web.Codec]()
	fileSerializerFactory = newRegister[fileSerializer]()
	routerMatcherFactory  = newRegister[RouterMatcherBuilder]()

	strategyFactory     = newRegister[StrategyBuilder]()
	registryTypeFactory = newRegister[RegistryTypeBuilder]()
)

// IDGeneratorBuilder 构建生成唯一 ID 的方法
//
// f 表示生成唯一 ID 的方法；s 为 f 依赖的服务，可以为空；
type IDGeneratorBuilder = func() (f func() string, s web.ServiceFunc)

// RegisterLogsHandler 注册日志的 [LogsWriterBuilder]
//
// name 为缓存的名称，这将在配置文件中被引用，如果存在同名，则会覆盖。
func RegisterLogsHandler(b LogsHandlerBuilder, name ...string) {
	logHandlersFactory.register(b, name...)
}

// RegisterCache 注册新的缓存方式
//
// name 为缓存的名称，这将在配置文件中被引用，如果存在同名，则会覆盖。
func RegisterCache(b CacheBuilder, name ...string) { cacheFactory.register(b, name...) }

// RegisterCompressor 注册压缩方法
//
// id 表示此压缩方法的唯一 ID，这将在配置文件中被引用；
func RegisterCompressor(id string, c web.Compressor) { compressorFactory.register(c, id) }

// RegisterIDGenerator 注册唯一 ID 生成器
//
// id 表示唯一 ID，这将在配置文件中被引用。如果同名会被覆盖；
func RegisterIDGenerator(id string, b IDGeneratorBuilder) { idGeneratorFactory.register(b, id) }

// RegisterCodec 注册用于序列化用户提交数据的方法
//
// name 为名称，这将在配置文件中被引用，如果存在同名，则会覆盖。
func RegisterCodec(mt web.Codec, name string) { codecFactory.register(mt, name) }

// RegisterFileSerializer 注册用于文件序列化的方法
//
// name 为当前数据的名称，这将在配置文件中被引用，如果存在同名，则会覆盖；
// ext 为文件的扩展名；
func RegisterFileSerializer(name string, m config.MarshalFunc, u config.UnmarshalFunc, ext ...string) {
	for _, e := range ext {
		for k, s := range fileSerializerFactory.items {
			if slices.Contains(s.exts, e) {
				panic(fmt.Sprintf("扩展名 %s 已经注册到 %s", e, k))
			}
		}
	}
	fileSerializerFactory.register(fileSerializer{marshal: m, unmarshal: u, exts: ext}, name)
}

func RegisterStrategy(f StrategyBuilder, name string) { strategyFactory.register(f, name) }

func RegisterRegistryType(f RegistryTypeBuilder, name string) { registryTypeFactory.register(f, name) }

func RegisterRouterMatcher(f RouterMatcherBuilder, name string) {
	routerMatcherFactory.register(f, name)
}

func init() {
	// RegisterLogsHandler

	RegisterLogsHandler(newFileLogsHandler, "file")
	RegisterLogsHandler(newTermLogsHandler, "term")
	RegisterLogsHandler(newSMTPLogsHandler, "smtp")

	// RegisterCache

	RegisterCache(func(dsn string) (cache.Driver, error) { return memory.New(), nil }, "memory")
	RegisterCache(func(dsn string) (cache.Driver, error) { return redis.NewFromURL(dsn) }, "redis")
	RegisterCache(func(dsn string) (cache.Driver, error) {
		return memcache.New(strings.Split(dsn, ";")...), nil
	}, "memcached", "memcache")

	// RegisterCompression

	RegisterCompressor("deflate-default", negotiation.NewDeflate(flate.DefaultCompression, nil))
	RegisterCompressor("deflate-best-compression", negotiation.NewDeflate(flate.BestCompression, nil))
	RegisterCompressor("deflate-best-speed", negotiation.NewDeflate(flate.BestSpeed, nil))

	RegisterCompressor("gzip-default", negotiation.NewGzip(gzip.DefaultCompression))
	RegisterCompressor("gzip-best-compression", negotiation.NewGzip(gzip.BestCompression))
	RegisterCompressor("gzip-best-speed", negotiation.NewGzip(gzip.BestSpeed))

	RegisterCompressor("compress-lsb-8", negotiation.NewLZW(lzw.LSB, 8))
	RegisterCompressor("compress-msb-8", negotiation.NewLZW(lzw.MSB, 8))

	RegisterCompressor("br-default", negotiation.NewBrotli(brotli.WriterOptions{Quality: brotli.DefaultCompression}))
	RegisterCompressor("br-best-compression", negotiation.NewBrotli(brotli.WriterOptions{Quality: brotli.BestCompression}))
	RegisterCompressor("br-best-speed", negotiation.NewBrotli(brotli.WriterOptions{Quality: brotli.BestSpeed}))

	RegisterCompressor("zstd-default", negotiation.NewZstd())

	// RegisterIDGenerator

	RegisterIDGenerator("date", func() (func() string, web.ServiceFunc) {
		u := unique.NewNumber(100)
		return u.String, u.Serve
	})
	RegisterIDGenerator("string", func() (func() string, web.ServiceFunc) {
		u := unique.NewString(100)
		return u.String, u.Serve
	})
	RegisterIDGenerator("number", func() (func() string, web.ServiceFunc) {
		u := unique.NewDate(100)
		return u.String, u.Serve
	})

	// RegisterCodec

	RegisterCodec(negotiation.NewJSON(), "json")
	RegisterCodec(negotiation.NewYAML(), "yaml")
	RegisterCodec(negotiation.NewCBOR(), "cbor")
	RegisterCodec(negotiation.NewXML(), "xml")
	RegisterCodec(form.New(), "form")
	RegisterCodec(negotiation.NewGOB(), "gob")
	RegisterCodec(negotiation.NewNOP(), "nop")

	// RegisterFileSerializer

	RegisterFileSerializer("json", server.JSONMarshal, server.JSONUnmarshal, ".json")
	RegisterFileSerializer("xml", xml.Marshal, xml.Unmarshal, ".xml")
	RegisterFileSerializer("yaml", yaml.Marshal, yaml.Unmarshal, ".yaml", ".yml")
	RegisterFileSerializer("toml", toml.Marshal, toml.Unmarshal, ".toml")
	RegisterFileSerializer("jsonc", server.JSONCMarshal, server.JSONCUnmarshal, ".jsonc")

	// micro

	RegisterStrategy(registry.NewRandomStrategy, "random")
	RegisterStrategy(registry.NewRoundRobinStrategy, "round-robin")
	RegisterStrategy(registry.NewWeightedRandomStrategy, "weighted-random")
	RegisterStrategy(registry.NewWeightedRoundRobinStrategy, "weighted-round-robin")

	RegisterRegistryType(func(c web.Cache, s *registry.Strategy, arg ...string) registry.Registry {
		if c == nil {
			panic("参数 c 不能为空")
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

	RegisterRouterMatcher(func(arg ...string) web.RouterMatcher { return mux.NewHosts(false, arg...) }, "hosts")
	RegisterRouterMatcher(func(arg ...string) web.RouterMatcher { return mux.NewPathVersion(arg[0], arg[1:]...) }, "prefix")
	RegisterRouterMatcher(func(arg ...string) web.RouterMatcher { return mux.NewHeaderVersion(arg[0], arg[0], nil, arg[2:]...) }, "version")
	RegisterRouterMatcher(func(arg ...string) web.RouterMatcher { return nil }, "any")
}
