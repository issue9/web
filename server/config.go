// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package server

import (
	"runtime/debug"
	"time"

	"github.com/issue9/cache"
	"github.com/issue9/config"
	"golang.org/x/text/language"

	"github.com/issue9/web"
	"github.com/issue9/web/locales"
	"github.com/issue9/web/selector"
)

type configOf[T any] struct {
	XMLName struct{} `yaml:"-" json:"-" xml:"web"`

	// 内存限制
	//
	// 如果小于等于 0，表示不设置该值。
	// 具体功能可参考[文档]。除非对该功能非常了解，否则不建议设置该值。
	//
	// [文档]: https://pkg.go.dev/runtime/debug#SetMemoryLimit
	MemoryLimit int64 `yaml:"memoryLimit,omitempty" json:"memoryLimit,omitempty" xml:"memoryLimit,attr,omitempty"`

	// 日志系统的配置项
	//
	// 如果为空，所有日志输出都将被抛弃。
	Logs *logsConfig `yaml:"logs,omitempty" xml:"logs,omitempty" json:"logs,omitempty"`

	// 指定默认语言
	//
	// 服务端的默认语言以及客户端未指定 accept-language 时的默认值。
	// 如果为空，则会尝试当前用户的语言。
	Language    string `yaml:"language,omitempty" json:"language,omitempty" xml:"language,attr,omitempty"`
	languageTag language.Tag

	// 与 HTTP 请求相关的设置项
	HTTP *httpConfig `yaml:"http,omitempty" json:"http,omitempty" xml:"http,omitempty"`

	// 时区名称
	//
	// 可以是 Asia/Shanghai 等，具体可参考[文档]。
	//
	// 为空和 Local(注意大小写) 值都会被初始化本地时间。
	//
	// [文档]: https://en.wikipedia.org/wiki/List_of_tz_database_time_zones
	Timezone string `yaml:"timezone,omitempty" json:"timezone,omitempty" xml:"timezone,omitempty"`
	location *time.Location

	// 指定缓存对象
	//
	// 如果为空，则会采用内存作为缓存对象。
	Cache *cacheConfig `yaml:"cache,omitempty" json:"cache,omitempty" xml:"cache,omitempty"`
	cache cache.Driver

	// 压缩的相关配置
	//
	// 如果为空，那么不支持压缩功能。
	Compressors []*compressConfig `yaml:"compressions,omitempty" json:"compressions,omitempty" xml:"compressions>compression,omitempty"`
	compressors []*Compression

	// 指定配置文件的序列化
	//
	// 可通过 [RegisterFileSerializer] 进行添加额外的序列化方法。默认可用为：
	//  - yaml 支持 .yaml 和 .yml 两种后缀名的文件
	//  - xml 支持 .xml 后缀名的文件
	//  - json 支持 .json 后缀名的文件
	//
	// 如果为空，表示支持以上所有格式。
	FileSerializers []string `yaml:"fileSerializers,omitempty" json:"fileSerializers,omitempty" xml:"fileSerializers>fileSerializer,omitempty"`
	config          *Config

	// 指定可用的 mimetype
	//
	// 如果为空，那么将不支持任何格式的内容输出。
	Mimetypes []*mimetypeConfig `yaml:"mimetypes,omitempty" json:"mimetypes,omitempty" xml:"mimetypes>mimetype,omitempty"`
	mimetypes []*Mimetype

	// 唯一 ID 生成器
	//
	// 该值由 [RegisterIDGenerator] 注册而来，默认情况下，有以下三个选项：
	//  - date 日期格式，默认值；
	//  - string 普通的字符串；
	//  - number 数值格式；
	// NOTE: 一旦运行在生产环境，就不应该修改此属性，新的生成器无法保证生成的 ID 不会与之前的重复。
	IDGenerator string `yaml:"idGenerator,omitempty" json:"idGenerator,omitempty" xml:"idGenerator,omitempty"`
	idGenerator IDGenerator

	// Problem 中 type 字段的前缀
	ProblemTypePrefix string `yaml:"problemTypePrefix,omitempty" json:"problemTypePrefix,omitempty" xml:"problemTypePrefix,omitempty"`

	// 指定服务发现和注册中心
	//
	// NOTE: 作为微服务和网关时才会有效果
	Registry *registryConfig `yaml:"registry,omitempty" json:"registry,omitempty" xml:"registry,omitempty"`

	// 作为微服务时的节点地址
	//
	// NOTE: 作为微服务时才会有效果
	Peer string `yaml:"peer,omitempty" json:"peer,omitempty" xml:"peer,omitempty"`
	peer selector.Peer

	// 作为微服务网关时的外部请求映射方式
	//
	// NOTE: 作为微服务的网关时才会有效果
	Mappers []*mapperConfig `yaml:"mappers,omitempty" json:"mappers,omitempty" xml:"mappers>mapper,omitempty"`
	mapper  Mapper

	// 用户自定义的配置项
	User *T `yaml:"user,omitempty" json:"user,omitempty" xml:"user,omitempty"`

	// 由其它选项生成的初始化方法
	init []func(web.Server)
}

// LoadOptions 从配置文件初始化 [Options] 对象
//
// configDir 项目配置文件所在的目录；
// filename 用于指定项目的配置文件，相对于 configDir 文件系统。
// 如果此值为空，将返回 &Options{Config: &Config{Dir: configDir}}；
//
// 序列化方法由 [RegisterFileSerializer] 注册的列表中根据 filename 的扩展名进行查找。
//
// T 表示用户自定义的数据项，该数据来自配置文件中的 user 字段。
// 如果实现了 [config.Sanitizer] 接口，则在加载后调用该接口中；
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
func LoadOptions[T any](configDir, filename string) (*Options, *T, error) {
	if filename == "" {
		return &Options{Config: &Config{Dir: configDir}}, nil, nil
	}

	conf, err := loadConfigOf[T](configDir, filename)
	if err != nil {
		return nil, nil, web.NewStackError(err)
	}

	return &Options{
		Config:            conf.config,
		Location:          conf.location,
		Cache:             conf.cache,
		HTTPServer:        conf.HTTP.httpServer,
		Logs:              conf.Logs.logs,
		Language:          conf.languageTag,
		RoutersOptions:    conf.HTTP.routersOptions,
		IDGenerator:       conf.idGenerator,
		RequestIDKey:      conf.HTTP.RequestID,
		Compressions:      conf.compressors,
		Mimetypes:         conf.mimetypes,
		ProblemTypePrefix: conf.ProblemTypePrefix,
		Init:              conf.init,
	}, conf.User, nil
}

func (conf *configOf[T]) SanitizeConfig() *web.FieldError {
	if conf.MemoryLimit > 0 {
		conf.init = append(conf.init, func(web.Server) { debug.SetMemoryLimit(conf.MemoryLimit) })
	}

	if err := conf.buildCache(); err != nil {
		return err.AddFieldParent("cache")
	}

	if conf.Logs == nil {
		conf.Logs = &logsConfig{}
	}

	if err := conf.Logs.build(); err != nil {
		return err.AddFieldParent("logs")
	}
	conf.init = append(conf.init, func(s web.Server) { s.OnClose(conf.Logs.cleanup...) })

	if conf.Language != "" {
		tag, err := language.Parse(conf.Language)
		if err != nil {
			return web.NewFieldError("language.", err)
		}
		conf.languageTag = tag
	}

	if err := conf.buildTimezone(); err != nil {
		return err
	}

	if conf.HTTP == nil {
		conf.HTTP = &httpConfig{}
	}
	if err := conf.HTTP.sanitize(); err != nil {
		return err.AddFieldParent("http")
	}
	if conf.HTTP.init != nil {
		conf.init = append(conf.init, conf.HTTP.init)
	}

	if err := conf.sanitizeCompresses(); err != nil {
		return err.AddFieldParent("compressions")
	}

	if err := conf.sanitizeMimetypes(); err != nil {
		return err
	}

	if err := conf.sanitizeFileSerializers(); err != nil {
		return err.AddFieldParent("fileSerializer")
	}

	if conf.IDGenerator == "" {
		conf.IDGenerator = "date"
	}
	if g, found := idGeneratorFactory.get(conf.IDGenerator); found {
		f, srv := g()
		conf.idGenerator = f
		if srv != nil {
			conf.init = append(conf.init, func(s web.Server) { s.Services().Add(locales.UniqueIdentityGenerator, srv) })
		}
	}

	if err := conf.buildMicro(conf.cache); err != nil {
		return err
	}

	if conf.User != nil {
		if s, ok := (any)(conf.User).(config.Sanitizer); ok {
			if err := s.SanitizeConfig(); err != nil {
				return err.AddFieldParent("user")
			}
		}
	}

	return nil
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
func CheckConfigSyntax[T any](configDir, filename string) error {
	_, err := loadConfigOf[T](configDir, filename)
	return err
}
