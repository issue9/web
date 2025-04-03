// SPDX-FileCopyrightText: 2018-2025 caixw
//
// SPDX-License-Identifier: MIT

//go:generate web htmldoc -lang=zh-CN -dir=./ -o=./CONFIG.md -object=configOf

// Package config 从配置文件加载 [server.Options]
package config

import (
	"io/fs"
	"log"
	"runtime/debug"
	"time"

	"github.com/issue9/cache"
	"github.com/issue9/config"
	"github.com/issue9/localeutil"
	"github.com/issue9/mux/v9"
	"golang.org/x/text/language"
	"golang.org/x/text/message/catalog"

	"github.com/issue9/web"
	"github.com/issue9/web/internal/locale"
	"github.com/issue9/web/locales"
	"github.com/issue9/web/selector"
	"github.com/issue9/web/server"
)

// 在项目正式运行之后，对于配置项的修改应该慎之又慎，
// 不当的修改可能导致项目运行过程中出错，比如改变了唯一 ID
// 的生成规则，可能会导致新生成的唯一 ID 与之前的 ID 重复。
type configOf[T comparable] struct {
	XMLName struct{} `yaml:"-" json:"-" toml:"-" xml:"web"`

	dir string

	// 内存限制
	//
	// 如果小于等于 0，表示不设置该值。
	// 具体功能可参考[文档]。除非对该功能非常了解，否则不建议设置该值。
	//
	// [文档]: https://pkg.go.dev/runtime/debug#SetMemoryLimit
	MemoryLimit int64 `yaml:"memoryLimit,omitempty" json:"memoryLimit,omitempty" xml:"memoryLimit,attr,omitempty" toml:"memoryLimit,omitempty"`

	// 日志系统的配置项
	//
	// 如果为空，所有日志输出都将被抛弃。
	Logs *logsConfig `yaml:"logs,omitempty" xml:"logs,omitempty" json:"logs,omitempty" toml:"logs,omitempty"`

	// 指定默认语言
	//
	// 服务端的默认语言以及客户端未指定 accept-language 时的默认值。
	// 如果为空，则会尝试当前用户的语言。
	Language    string `yaml:"language,omitempty" json:"language,omitempty" xml:"language,attr,omitempty" toml:"language,omitempty"`
	languageTag language.Tag

	// 与 HTTP 请求相关的设置项
	HTTP *httpConfig `yaml:"http,omitempty" json:"http,omitempty" xml:"http,omitempty" toml:"http,omitempty"`

	// 时区名称
	//
	// 可以是 Asia/Shanghai 等，具体可参考[文档]。
	//
	// 为空和 Local(注意大小写) 值都会被初始化本地时间。
	//
	// [文档]: https://en.wikipedia.org/wiki/List_of_tz_database_time_zones
	Timezone string `yaml:"timezone,omitempty" json:"timezone,omitempty" xml:"timezone,omitempty" toml:"timezone,omitempty"`
	location *time.Location

	// 指定缓存对象
	//
	// 如果为空，则会采用内存作为缓存对象。
	Cache *cacheConfig `yaml:"cache,omitempty" json:"cache,omitempty" xml:"cache,omitempty" toml:"cache,omitempty"`
	cache cache.Driver

	// 指定配置文件的序列化
	//
	// 可通过 [RegisterFileSerializer] 进行添加额外的序列化方法。默认为空，可以有以下可选值：
	//  - yaml 支持 .yaml 和 .yml 两种后缀名的文件
	//  - xml 支持 .xml 后缀名的文件
	//  - json 支持 .json 后缀名的文件
	//  - toml 支持 .toml 后缀名的文件
	//
	// 如果为空，表示支持以上所有格式。
	FileSerializers []string `yaml:"fileSerializers,omitempty" json:"fileSerializers,omitempty" xml:"fileSerializers>fileSerializer,omitempty" toml:"fileSerializers,omitempty"`
	config          *config.Config

	// 压缩的相关配置
	//
	// 如果为空，那么不支持压缩功能。
	Compressors []*compressConfig `yaml:"compressions,omitempty" json:"compressions,omitempty" xml:"compressions>compression,omitempty" toml:"compressions,omitempty"`

	// 指定可用的 mimetype
	//
	// 如果为空，那么将不支持任何格式的内容输出。
	Mimetypes []*mimetypeConfig `yaml:"mimetypes,omitempty" json:"mimetypes,omitempty" xml:"mimetypes>mimetype,omitempty" toml:"mimetypes,omitempty"`

	codec *web.Codec

	// 唯一 ID 生成器
	//
	// 该值由 [RegisterIDGenerator] 注册而来，默认情况下，有以下三个选项：
	//  - date 日期格式，默认值；
	//  - string 普通的字符串；
	//  - number 数值格式；
	// NOTE: 一旦运行在生产环境，就不应该修改此属性，除非能确保新的函数生成的 ID 不与之前生成的 ID 重复。
	IDGenerator string `yaml:"idGenerator,omitempty" json:"idGenerator,omitempty" xml:"idGenerator,omitempty" toml:"idGenerator,omitempty"`
	idGenerator func() string

	// Problem 中 type 字段的前缀
	ProblemTypePrefix string `yaml:"problemTypePrefix,omitempty" json:"problemTypePrefix,omitempty" xml:"problemTypePrefix,omitempty" toml:"problemTypePrefix,omitempty"`

	// OnRender 修改渲染结构
	//
	// 可通过 [RegisterOnRender] 进行添加额外的序列化方法。默认为空，可以有以下可选值：
	//  - render200 所有输出都是以 [server.RenderResponse] 作为返回对象；
	OnRender string `yaml:"onRender,omitempty" json:"onRender,omitempty" xml:"onRender,omitempty" toml:"onRender,omitempty"`
	onRender func(int, any) (int, any)

	// 指定服务发现和注册中心
	//
	// NOTE: 作为微服务和网关时才会有效果
	Registry *registryConfig `yaml:"registry,omitempty" json:"registry,omitempty" xml:"registry,omitempty" toml:"registry,omitempty"`

	// 作为微服务时的节点地址
	//
	// NOTE: 作为微服务时才会有效果
	Peer string `yaml:"peer,omitempty" json:"peer,omitempty" xml:"peer,omitempty" toml:"peer,omitempty"`
	peer selector.Peer

	// 作为微服务网关时的外部请求映射方式
	//
	// NOTE: 作为微服务的网关时才会有效果
	Mappers []*mapperConfig `yaml:"mappers,omitempty" json:"mappers,omitempty" xml:"mappers>mapper,omitempty" toml:"mappers,omitempty"`
	mapper  map[string]mux.Matcher

	// 用户自定义的配置项
	User T `yaml:"user,omitempty" json:"user,omitempty" xml:"user,omitempty" toml:"user,omitempty"`

	// 由其它选项生成的初始化方法
	init []func(*server.Options)
}

// Load 从配置文件初始化 [server.Options] 对象
//
// configDir 项目配置文件所在的目录；
// filename 用于指定项目的配置文件，相对于 configDir 文件系统。
// 如果此值为空，将返回 &Options{Config: config.Dir(nil, configDir)}；
//
// 序列化方法由 [RegisterFileSerializer] 注册的列表中根据 filename 的扩展名进行查找。
//
// T 表示用户自定义的数据项，该数据来自配置文件中的 user 字段。
// 如果实现了 [config.Sanitizer] 接口，则在加载后调用该接口；
//
// # 配置文件
//
// 对于配置文件各个字段的定义，可参考当前目录下的 CONFIG.html。
// 配置文件中除了固定的字段之外，还提供了泛型变量 User 用于指定用户自定义的额外字段。
//
// # 注册函数
//
// 当前包提供大量的注册函数，以用将某些无法直接采用序列化的内容转换可序列化的。
// 比如通过 [RegisterCompression] 将 `gzip-default` 等字符串表示成压缩算法，
// 以便在配置文件进行指定。
//
// 所有的注册函数处理逻辑上都相似，碰上同名的会覆盖，否则是添加。
// 且默认情况下都提供了一些可选项，只有在用户需要额外添加自己的内容时才需要调用注册函数。
func Load[T comparable](configDir, filename string) (*server.Options, T, error) {
	var zero T
	if filename == "" {
		return &server.Options{Config: config.Dir(nil, configDir)}, zero, nil
	}

	conf, err := loadConfigOf[T](configDir, filename)
	if err != nil {
		return nil, zero, web.NewStackError(err)
	}

	o := &server.Options{
		Config:            conf.config,
		Location:          conf.location,
		Cache:             conf.cache,
		HTTPServer:        conf.HTTP.httpServer,
		Logs:              conf.Logs.logs,
		Language:          conf.languageTag,
		RoutersOptions:    make([]web.RouterOption, 0, 5),
		IDGenerator:       conf.idGenerator,
		RequestIDKey:      conf.HTTP.RequestID,
		Codec:             conf.codec,
		ProblemTypePrefix: conf.ProblemTypePrefix,
		OnRender:          conf.onRender,
		Plugins:           make([]web.Plugin, 0, 5),
	}

	for _, i := range conf.init {
		i(o)
	}

	return o, conf.User, nil
}

func (conf *configOf[T]) SanitizeConfig() *web.FieldError {
	if conf.MemoryLimit > 0 {
		conf.init = append(conf.init, func(*server.Options) { debug.SetMemoryLimit(conf.MemoryLimit) })
	}

	if conf.Language != "" {
		tag, err := language.Parse(conf.Language)
		if err != nil {
			return web.NewFieldError("language", err)
		}
		conf.languageTag = tag
	}

	if err := conf.buildCache(); err != nil {
		return err.AddFieldParent("cache")
	}

	if err := conf.buildLogs(); err != nil {
		return err
	}

	if err := conf.buildTimezone(); err != nil {
		return err
	}

	if err := conf.buildHTTP(); err != nil {
		return err
	}

	if err := conf.buildCodec(); err != nil {
		return err
	}

	if err := conf.buildConfig(); err != nil {
		return err
	}

	conf.buildIDGen()

	if conf.OnRender != "" {
		if or, found := onRenderFactory.get(conf.OnRender); found {
			conf.onRender = or
		} else {
			return web.NewFieldError("onRender", locales.ErrNotFound())
		}
	}

	if err := conf.buildMicro(conf.cache); err != nil {
		return err
	}

	var zero T
	if conf.User != zero {
		if s, ok := (any)(conf.User).(config.Sanitizer); ok {
			if err := s.SanitizeConfig(); err != nil {
				return err.AddFieldParent("user")
			}
		}
	}

	return nil
}

func (conf *configOf[T]) buildIDGen() {
	if conf.IDGenerator == "" {
		conf.IDGenerator = "date"
	}
	if g, found := idGeneratorFactory.get(conf.IDGenerator); found {
		f, srv := g()
		conf.idGenerator = f
		if srv != nil {
			conf.init = append(conf.init, func(o *server.Options) {
				o.Plugins = append(o.Plugins, web.PluginFunc(func(s web.Server) {
					s.Services().Add(locales.UniqueIdentityGenerator, srv)
				}))
			})
		}
	}
}

func (conf *configOf[T]) buildTimezone() *web.FieldError {
	if conf.Timezone == "" {
		return nil
	}

	loc, err := time.LoadLocation(conf.Timezone)
	if err != nil {
		return config.NewFieldError("timezone", err)
	}
	conf.location = loc

	return nil
}

// CheckConfigSyntax 检测配置项语法是否正确
func CheckConfigSyntax[T comparable](configDir, filename string) error {
	_, err := loadConfigOf[T](configDir, filename)
	return err
}

// NewPrinter 根据参数指定的配置文件构建一个本地化的打印对象
//
// 语言由 [localeutil.DetectUserLanguageTag] 决定。
// fsys 指定了加载本地化文件的文件系统，glob 则指定了加载的文件匹配规则；
// 对于文件的序列化方式则是根据后缀名从由 [RegisterFileSerializer] 注册的项中查找。
func NewPrinter(glob string, fsys ...fs.FS) (*localeutil.Printer, error) {
	tag, err := localeutil.DetectUserLanguageTag()
	if err != nil {
		log.Println(err) // 输出错误，但是不中断执行
	}

	b := catalog.NewBuilder(catalog.Fallback(tag))
	if err := locale.Load(buildSerializerFromFactory(), b, glob, fsys...); err != nil {
		return nil, err
	}

	p, _ := locale.NewPrinter(tag, b)
	return p, nil
}
