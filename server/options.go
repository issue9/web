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
)

// RequestIDKey 报头中传递 request id 的报头名称
const RequestIDKey = header.RequestIDKey

// DefaultConfigDir 默认的配置目录地址
const DefaultConfigDir = "@.config"

type (
	// Options [web.Server] 的初始化参数
	//
	// 这些参数都有默认值，且无法在 [web.Server] 初始化之后进行更改。
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
		// 如果为空，将采用 [unique.NewDate] 作为生成方法。
		IDGenerator IDGenerator

		// 路由选项
		//
		// 如果为空，会添加 [Recovery] 作为默认值。
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
		// 框架中的日志输出时，如果该信息实现了 [LocaleStringer] 接口，
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
		// 默认值为 [web.ProblemAboutBlank]。
		ProblemTypePrefix string

		// Init 其它的一些初始化操作
		//
		// 在此可以让用户能实际操作 [Server] 之前对 Server 进行一些操作。
		Init []func(web.Server)
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

func sanitizeOptions(o *Options) (*Options, *config.FieldError) {
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
	cfg, err := config.BuildDir(nil, DefaultConfigDir)
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
		u := unique.NewDate(1000)
		o.IDGenerator = u.String
		o.Init = append(o.Init, func(s web.Server) {
			s.Services().Add(locales.UniqueIdentityGenerator, u)
		})
	}

	if o.Cache == nil {
		c, job := NewMemory()
		o.Cache = c
		o.Init = append(o.Init, func(s web.Server) { // AddTicker 依赖 IDGenerator
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
		o.RoutersOptions = []web.RouterOption{Recovery(o.logs.ERROR())}
	}

	if o.RequestIDKey == "" {
		o.RequestIDKey = RequestIDKey
	}

	c, fe := buildCodec(o.Mimetypes, o.Compressions)
	if err != nil {
		return nil, fe
	}
	o.codec = c

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
