// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package server

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"time"

	"github.com/issue9/cache"
	"github.com/issue9/cache/caches/memcache"
	"github.com/issue9/cache/caches/memory"
	"github.com/issue9/cache/caches/redis"
	"github.com/issue9/config"
	"github.com/issue9/localeutil"
	"github.com/issue9/logs/v7"
	"github.com/issue9/unique/v2"
	"golang.org/x/text/language"
	"golang.org/x/text/message/catalog"
	"gopkg.in/yaml.v3"

	"github.com/issue9/web"
	"github.com/issue9/web/internal/header"
	"github.com/issue9/web/internal/locale"
	"github.com/issue9/web/locales"
	"github.com/issue9/web/selector"
	"github.com/issue9/web/server/registry"
)

const (
	RequestIDKey     = header.RequestIDKey // 报头中传递 request id 的报头名称
	DefaultConfigDir = "@.config"          // 默认的配置目录地址
)

const (
	typeHTTP int = iota
	typeGateway
	typeService
)

type (
	// Options [web.Server] 的初始化参数
	//
	// 这些参数都有默认值，且无法在 [web.Server] 初始化之后进行更改。
	//
	// 初始化方式，可以直接采用 &Options{...} 的方式，表示所有项都采用默认值。
	// 也可以采用 [LoadOptions] 从配置文件中加载相应在的数据进行初始化。
	Options struct {
		// 项目的配置项
		Config *Config
		config *config.Config

		// 服务器的时区
		//
		// 默认值为 [time.Local]
		Location *time.Location

		// 缓存系统
		//
		// 内置了以下几种驱动：
		//  - [NewMemory]
		//  - [NewMemcache]
		//  - [NewRedisFromURL]
		// 如果为空，采用 [NewMemory] 作为默认值。
		Cache cache.Driver

		// 日志的相关设置
		//
		// 如果此值为空，表示不会输出任何信息。
		Logs *Logs
		logs *logs.Logs

		// http.Server 实例的值
		//
		// 可以为零值。
		HTTPServer *http.Server

		// 生成唯一字符串的方法
		//
		// 供 [Server.UniqueID] 使用。
		//
		// 如果为空，将采用 [unique.NewString] 作为生成方法。
		//
		// NOTE: 该值的修改，可能造成项目中的唯一 ID 不再唯一。
		IDGenerator IDGenerator

		// 路由选项
		//
		// 如果为空，会添加 [web.Recovery] 作为默认值。
		RoutersOptions []web.RouterOption

		// 指定获取 x-request-id 内容的报头名
		//
		// 如果为空，则采用 [RequestIDKey] 作为默认值
		RequestIDKey string

		// 可用的压缩类型
		//
		// 默认为空。表示不需要该功能。
		Compressions []*Compression

		// 指定可用的 mimetype
		//
		// 默认采用 [JSONMimetypes]。
		Mimetypes []*Mimetype

		codec *web.Codec // 由 Compressions 和 Mimetypes 形成

		// 默认的语言标签
		//
		// 在用户请求的报头中没有匹配的语言标签时，会采用此值作为该用户的本地化语言，
		// 同时也用来初始化 [Server.Locale.Printer]。
		//
		// 框架中的日志输出时，如果该信息实现了 [web.LocaleStringer] 接口，
		// 将会转换成此设置项的语言。
		//
		// 如果为空，则会尝试读取当前系统的本地化信息。
		Language language.Tag

		// 本地化的数据
		//
		// 如果为空，则会被初始化成一个空对象。
		// Catalog 中会强行插入一条 tag 与 [Options.Language] 相同的翻译项，
		// 以保证能正确构建 [web.Server.Printer] 对象。
		Catalog *catalog.Builder

		locale *locale.Locale

		// ProblemTypePrefix 所有 type 字段的前缀
		//
		// 如果该值为 [web.ProblemAboutBlank]，将不输出 ID 值；其它值则作为前缀添加。
		// 空值是合法的值，表示不需要添加前缀。
		ProblemTypePrefix string

		// OnRender 可实现对渲染结果的调整
		//
		// NOTE: 该值的修改，可能造成所有接口返回数据结构的变化。
		OnRender func(status int, body any) (int, any)

		// Init 其它的一些初始化操作
		//
		// 在此可以让用户能实际操作 [Server] 之前对其进行一些修改。
		Init []func(web.Server)

		// 以下微服务相关的设置

		// Registry 作为微服务时的注册中心实例
		//
		// NOTE: 仅在 [NewService] 和 [NewGateway] 中才会有效果。
		Registry registry.Registry

		// Peer 作为微服务终端时的地址
		//
		// NOTE: 仅在 [NewService] 中才会有效果。
		Peer selector.Peer

		// Mapper 作为微服务网关时的 URL 映射关系
		//
		// NOTE: 仅在 [NewGateway] 中才会有效果。
		Mapper Mapper
	}

	// Config 项目配置文件的配置
	Config struct {
		// Dir 项目配置目录
		//
		// 如果涉及到需要读取配置文件的，可以指定此对象，之后可通过此对象统一处理各类配置文件。
		// 如果为空，则会采用 [DefaultConfigDir]。
		Dir string

		// Serializers 支持的序列化方法列表
		//
		// 如果为空，则会默认支持 yaml、json 两种方式；
		Serializers []*FileSerializer
	}

	// FileSerializer 对于文件序列化的配置
	FileSerializer struct {
		// Exts 支持的扩展名
		Exts []string

		// Marshal 序列化方法
		Marshal config.MarshalFunc

		// Unmarshal 反序列化方法
		Unmarshal config.UnmarshalFunc
	}

	// IDGenerator 生成唯一 ID 的函数
	IDGenerator = func() string
)

func (c *Config) asConfig() (*config.Config, error) {
	s := make(config.Serializer, len(c.Serializers))
	for _, ser := range c.Serializers {
		s.Add(ser.Marshal, ser.Unmarshal, ser.Exts...)
	}

	return config.BuildDir(s, c.Dir)
}

func sanitizeOptions(o *Options, t int) (*Options, *config.FieldError) {
	if o == nil {
		o = &Options{}
	}

	if o.Config == nil {
		o.Config = &Config{
			Dir: DefaultConfigDir,
			Serializers: []*FileSerializer{
				{Exts: []string{".yaml", ".yml"}, Marshal: yaml.Marshal, Unmarshal: yaml.Unmarshal},
				{Exts: []string{".json"}, Marshal: json.Marshal, Unmarshal: json.Unmarshal},
				{Exts: []string{".xml"}, Marshal: xml.Marshal, Unmarshal: xml.Unmarshal},
			},
		}
	}
	cfg, err := o.Config.asConfig()
	if err != nil {
		return nil, config.NewFieldError("Config", err)
	}
	o.config = cfg

	if o.Location == nil {
		o.Location = time.Local
	}

	if o.HTTPServer == nil {
		o.HTTPServer = &http.Server{}
	}

	if o.IDGenerator == nil {
		u := unique.NewString(1000)
		o.IDGenerator = u.String
		o.Init = append(o.Init, func(s web.Server) {
			s.Services().Add(locales.UniqueIdentityGenerator, u)
		})
	}

	if o.Cache == nil {
		c, job := NewMemory()
		o.Cache = c
		o.Init = append(o.Init, func(s web.Server) {
			s.Services().AddTicker(locales.RecycleLocalCache, job, time.Minute, false, false)
		})
	}

	if o.Language == language.Und {
		tag, err := localeutil.DetectUserLanguageTag()
		if err != nil {
			return nil, config.NewFieldError("Language", err)
		}
		o.Language = tag
	}

	if o.Catalog == nil {
		o.Catalog = catalog.NewBuilder(catalog.Fallback(o.Language))
	}

	o.locale = locale.New(o.Language, o.config, o.Catalog)

	if err := o.buildLogs(o.locale.Printer()); err != nil {
		return nil, err
	}

	if len(o.RoutersOptions) == 0 {
		o.RoutersOptions = []web.RouterOption{web.Recovery(http.StatusInternalServerError, o.logs.ERROR())}
	}

	if o.RequestIDKey == "" {
		o.RequestIDKey = RequestIDKey
	}

	c, fe := buildCodec(o.Mimetypes, o.Compressions)
	if fe != nil {
		return nil, fe
	}
	o.codec = c

	switch t {
	case typeHTTP: // 不需要处理任何数据
	case typeGateway:
		if o.Mapper == nil {
			return nil, web.NewFieldError("Mapper", locales.CanNotBeEmpty)
		}
		for k, v := range o.Mapper {
			if v == nil {
				return nil, web.NewFieldError("Mapper["+k+"]", locales.CanNotBeEmpty)
			}
		}
		if o.Registry == nil {
			return nil, web.NewFieldError("Registry", locales.CanNotBeEmpty)
		}
	case typeService:
		if o.Peer == nil {
			return nil, web.NewFieldError("Peer", locales.CanNotBeEmpty)
		}
		if o.Registry == nil {
			return nil, web.NewFieldError("Registry", locales.CanNotBeEmpty)
		}
	default:
		panic("参数 t 取值错误")
	}

	return o, nil
}

// NewMemory 声明基于内在的缓存对象
func NewMemory() (cache.Driver, web.JobFunc) {
	d, job := memory.New()
	return d, func(now time.Time) error {
		job(now)
		return nil
	}
}

// NewRedisFromURL 声明基于 redis 的缓存对象
//
// 参数说明可参考 [redis.NewFromURL]。
func NewRedisFromURL(url string) (cache.Driver, error) { return redis.NewFromURL(url) }

// NewMemcache 声明基于 memcache 的缓存对象
//
// 参数说明可参考 [memcache.New]。
func NewMemcache(addr ...string) cache.Driver { return memcache.New(addr...) }

func (o *Options) internalServer(name, version string, s web.Server) *web.InternalServer {
	return web.InternalNewServer(s, name, version, o.Location, o.logs, o.IDGenerator, o.locale, o.Cache, o.codec, o.RequestIDKey, o.ProblemTypePrefix, o.OnRender, o.RoutersOptions...)
}

// NumberID 构建数字形式的唯一 ID
//
// NOTE: 基于时间戳，不能保证多实例模式下也具有唯一性。
func NumberID(buffSize int) (IDGenerator, web.Service) {
	u := unique.NewNumber(buffSize)
	return u.String, u
}

// StringID 构建包含任意字符的唯一 ID
//
// NOTE: 基于时间戳，不能保证多实例模式下也具有唯一性。
func StringID(buffSize int) (IDGenerator, web.Service) {
	u := unique.NewString(buffSize)
	return u.String, u
}

// DateID 构建日期格式的唯一 ID
//
// NOTE: 基于时间戳，不能保证多实例模式下也具有唯一性。
func DateID(buffSize int) (IDGenerator, web.Service) {
	u := unique.NewDate(buffSize)
	return u.String, u
}

// Render200 统一 API 的返回格式
//
// 状态码统一为 200；返回对象统一为 [Render200Response]；
func Render200(status int, body any) (int, any) {
	return http.StatusOK, &Render200Response{OK: !web.IsProblem(status), Status: status, Body: body}
}

// Render200Response API 统一的返回格式
type Render200Response struct {
	XMLName struct{} `json:"-" yaml:"-" xml:"body"`
	OK      bool     `json:"ok" yaml:"ok" xml:"ok,attr"`
	Status  int      `json:"status" yaml:"status" xml:"status,attr"`
	Body    any      `json:"body" yaml:"body" xml:"body"`
}
