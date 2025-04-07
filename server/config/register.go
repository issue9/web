// SPDX-FileCopyrightText: 2018-2025 caixw
//
// SPDX-License-Identifier: MIT

package config

import (
	"compress/flate"
	"compress/gzip"
	"compress/lzw"
	sj "encoding/json"
	sx "encoding/xml"
	"fmt"
	"slices"
	"strings"
	"time"

	st "github.com/BurntSushi/toml"
	"github.com/andybalholm/brotli"
	sy "github.com/goccy/go-yaml"
	"github.com/issue9/cache"
	"github.com/issue9/cache/caches/memcache"
	"github.com/issue9/cache/caches/memory"
	"github.com/issue9/cache/caches/redis"
	"github.com/issue9/config"
	"github.com/issue9/mux/v9"
	"github.com/issue9/unique/v2"

	"github.com/issue9/web"
	"github.com/issue9/web/compressor"
	"github.com/issue9/web/mimetype/cbor"
	"github.com/issue9/web/mimetype/form"
	"github.com/issue9/web/mimetype/gob"
	"github.com/issue9/web/mimetype/html"
	"github.com/issue9/web/mimetype/json"
	"github.com/issue9/web/mimetype/nop"
	"github.com/issue9/web/mimetype/xml"
	"github.com/issue9/web/mimetype/yaml"
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
	compressorFactory     = newRegister[compressor.Compressor]()
	idGeneratorFactory    = newRegister[IDGeneratorBuilder]()
	mimetypesFactory      = newRegister[mimetype]()
	fileSerializerFactory = newRegister[fileSerializer]()
	routerMatcherFactory  = newRegister[RouterMatcherBuilder]()
	onRenderFactory       = newRegister[func(int, any) (int, any)]()

	strategyFactory     = newRegister[StrategyBuilder]()
	registryTypeFactory = newRegister[RegistryTypeBuilder]()
)

// IDGeneratorBuilder 构建生成唯一 ID 的方法
//
// f 表示生成唯一 ID 的方法；s 为 f 依赖的服务，可以为空；
type IDGeneratorBuilder = func() (f func() string, s web.Service)

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

// RegisterCompression 注册压缩方法
//
// id 表示此压缩方法的唯一 ID，这将在配置文件中被引用；
func RegisterCompression(id string, c compressor.Compressor) { compressorFactory.register(c, id) }

// RegisterIDGenerator 注册唯一 ID 生成器
//
// id 表示唯一 ID，这将在配置文件中被引用。如果同名会被覆盖；
func RegisterIDGenerator(id string, b IDGeneratorBuilder) { idGeneratorFactory.register(b, id) }

// RegisterMimetype 注册用于序列化用户提交数据的方法
//
// name 为名称，这将在配置文件中被引用，如果存在同名，则会覆盖。
func RegisterMimetype(m web.MarshalFunc, u web.UnmarshalFunc, name string) {
	mimetypesFactory.register(mimetype{marshal: m, unmarshal: u}, name)
}

// RegisterFileSerializer 注册用于文件序列化的方法
//
// name 为当前数据的名称，这将在配置文件中被引用，如果存在同名，则会覆盖；
// ext 为文件的扩展名；
func RegisterFileSerializer(name string, m config.MarshalFunc, u config.UnmarshalFunc, ext ...string) {
	for _, e := range ext {
		for k, s := range fileSerializerFactory.items {
			if slices.Index(s.exts, e) >= 0 {
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

func RegisterOnRender(f func(int, any) (int, any), name string) { onRenderFactory.register(f, name) }

func init() {
	// RegisterLogsHandler

	RegisterLogsHandler(newFileLogsHandler, "file")
	RegisterLogsHandler(newTermLogsHandler, "term")
	RegisterLogsHandler(newSMTPLogsHandler, "smtp")

	// RegisterCache

	RegisterCache(func(dsn string) (cache.Driver, error) {
		return memory.New(), nil
	}, "memory")

	RegisterCache(func(dsn string) (cache.Driver, error) {
		return memcache.New(strings.Split(dsn, ";")...), nil
	}, "memcached", "memcache")

	RegisterCache(func(dsn string) (cache.Driver, error) {
		return redis.NewFromURL(dsn)
	}, "redis")

	// RegisterCompression

	RegisterCompression("deflate-default", compressor.NewDeflate(flate.DefaultCompression, nil))
	RegisterCompression("deflate-best-compression", compressor.NewDeflate(flate.BestCompression, nil))
	RegisterCompression("deflate-best-speed", compressor.NewDeflate(flate.BestSpeed, nil))

	RegisterCompression("gzip-default", compressor.NewGzip(gzip.DefaultCompression))
	RegisterCompression("gzip-best-compression", compressor.NewGzip(gzip.BestCompression))
	RegisterCompression("gzip-best-speed", compressor.NewGzip(gzip.BestSpeed))

	RegisterCompression("compress-lsb-8", compressor.NewLZW(lzw.LSB, 8))
	RegisterCompression("compress-msb-8", compressor.NewLZW(lzw.MSB, 8))

	RegisterCompression("br-default", compressor.NewBrotli(brotli.WriterOptions{Quality: brotli.DefaultCompression}))
	RegisterCompression("br-best-compression", compressor.NewBrotli(brotli.WriterOptions{Quality: brotli.BestCompression}))
	RegisterCompression("br-best-speed", compressor.NewBrotli(brotli.WriterOptions{Quality: brotli.BestSpeed}))

	RegisterCompression("zstd-default", compressor.NewZstd())

	// RegisterIDGenerator

	RegisterIDGenerator("date", func() (func() string, web.Service) {
		u := unique.NewNumber(100)
		return u.String, u
	})
	RegisterIDGenerator("string", func() (func() string, web.Service) {
		u := unique.NewString(100)
		return u.String, u
	})
	RegisterIDGenerator("number", func() (func() string, web.Service) {
		u := unique.NewDate(100)
		return u.String, u
	})

	// RegisterMimetype

	RegisterMimetype(json.Marshal, json.Unmarshal, "json")
	RegisterMimetype(yaml.Marshal, yaml.Unmarshal, "yaml")
	RegisterMimetype(cbor.Marshal, cbor.Unmarshal, "cbor")
	RegisterMimetype(xml.Marshal, xml.Unmarshal, "xml")
	RegisterMimetype(html.Marshal, html.Unmarshal, "html")
	RegisterMimetype(form.Marshal, form.Unmarshal, "form")
	RegisterMimetype(gob.Marshal, gob.Unmarshal, "gob")
	RegisterMimetype(nop.Marshal, nop.Unmarshal, "nop")

	// RegisterFileSerializer

	RegisterFileSerializer("json", sj.Marshal, sj.Unmarshal, ".json")
	RegisterFileSerializer("xml", sx.Marshal, sx.Unmarshal, ".xml")
	RegisterFileSerializer("yaml", sy.Marshal, sy.Unmarshal, ".yaml", ".yml")
	RegisterFileSerializer("toml", st.Marshal, st.Unmarshal, ".toml")

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

	// OnRender

	RegisterOnRender(server.Render200, "render200")
}
