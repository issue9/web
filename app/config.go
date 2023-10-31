// SPDX-License-Identifier: MIT

package app

import (
	"net/http"
	"runtime/debug"
	"time"

	"github.com/issue9/config"
	"golang.org/x/text/language"

	"github.com/issue9/web"
	"github.com/issue9/web/cache"
	"github.com/issue9/web/codec"
	"github.com/issue9/web/locales"
	"github.com/issue9/web/logs"
	"github.com/issue9/web/server"
)

type configOf[T any] struct {
	XMLName struct{} `yaml:"-" json:"-" xml:"web"`

	// 内存限制
	//
	// 如果小于等于 0，表示不设置该值。
	// 除非对该功能非常了解，否则不建议设置该值。
	//
	// 具体功能可参考 https://pkg.go.dev/runtime/debug#SetMemoryLimit
	MemoryLimit int64 `yaml:"memoryLimit,omitempty" json:"memoryLimit,omitempty" xml:"memoryLimit,attr,omitempty"`

	// 日志系统的配置项
	//
	// 如果为空，所有日志输出都将被抛弃。
	Logs    *logsConfig `yaml:"logs,omitempty" xml:"logs,omitempty" json:"logs,omitempty"`
	logs    *logs.Options
	cleanup []func() error

	// 指定默认语言
	//
	// 服务端的默认语言以及客户端未指定 accept-language 时的默认值。
	// 如果为空，则会尝试当前用户的语言。
	Language    string `yaml:"language,omitempty" json:"language,omitempty" xml:"language,attr,omitempty"`
	languageTag language.Tag

	// 与 HTTP 请求相关的设置项
	HTTP *httpConfig `yaml:"http,omitempty" json:"http,omitempty" xml:"http,omitempty"`
	http *http.Server

	// 时区名称
	//
	// 可以是 Asia/Shanghai 等，具体可参考：
	// https://en.wikipedia.org/wiki/List_of_tz_database_time_zones
	//
	// 为空和 Local(注意大小写) 值都会被初始化本地时间。
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
	compressors []*codec.Compression

	// 指定配置文件的序列化
	//
	// 如果为空，表示不支持。
	//
	// 可通过 [RegisterFileSerializer] 进行添加额外的序列化方法。默认可用为：
	//  - .yaml
	//  - .yml
	//  - .xml
	//  - .json
	FileSerializers []string `yaml:"fileSerializers,omitempty" json:"fileSerializers,omitempty" xml:"fileSerializers>fileSerializer,omitempty"`
	fileSerializers map[string]serializer
	config          *config.Config

	// 指定可用的 mimetype
	//
	// 如果为空，那么将不支持任何格式的内容输出。
	Mimetypes []*mimetypeConfig `yaml:"mimetypes,omitempty" json:"mimetypes,omitempty" xml:"mimetypes>mimetype,omitempty"`
	mimetypes []*codec.Mimetype

	// 唯一 ID 生成器
	//
	// 该值由 [RegisterIDGenerator] 注册而来，默认情况下，有以下三个选项：
	//  - date 日期格式，默认值；
	//  - string 普通的字符串；
	//  - number 数值格式；
	// NOTE: 一旦运行在生产环境，就不应该修改此属性，新的生成器无法保证生成的 ID 不会与之前的重复。
	IDGenerator string `yaml:"idGenerator,omitempty" json:"idGenerator,omitempty" xml:"idGenerator,omitempty"`
	idGenerator server.IDGenerator

	// Problem 中 type 字段的前缀
	ProblemTypePrefix string `yaml:"problemTypePrefix,omitempty" json:"problemTypePrefix,omitempty" xml:"problemTypePrefix,omitempty"`

	// 用户自定义的配置项
	User *T `yaml:"user,omitempty" json:"user,omitempty" xml:"user,omitempty"`

	// 由其它选项生成的初始化方法
	init []func(web.Server)
}

// NewServerOf 从配置文件初始化 [web.Server] 对象
//
// configDir 项目配置文件所在的目录；
// filename 用于指定项目的配置文件，相对于 configDir 文件系统。
// 序列化方法由 [RegisterFileSerializer] 注册的列表中根据 filename 的扩展名进行查找。
// 如果此值为空，将以 &web.Options{} 初始化 [web.Server]；
//
// T 表示用户自定义的数据项，该数据来自配置文件中的 user 字段。
// 如果实现了 [config.Sanitizer] 接口，则在加载后调用该接口中；
func NewServerOf[T any](name, version string, configDir, filename string) (web.Server, *T, error) {
	if filename == "" {
		c, err := config.AppDir(nil, configDir)
		if err != nil {
			return nil, nil, err
		}

		s, err := server.New(name, version, &server.Options{Config: c})
		return s, nil, err
	}

	conf, err := loadConfigOf[T](configDir, filename)
	if err != nil {
		return nil, nil, web.NewStackError(err)
	}

	if conf.MemoryLimit > 0 {
		debug.SetMemoryLimit(conf.MemoryLimit)
	}

	opt := &server.Options{
		Config:         conf.config,
		Location:       conf.location,
		Cache:          conf.cache,
		HTTPServer:     conf.http,
		Logs:           conf.logs,
		Language:       conf.languageTag,
		RoutersOptions: conf.HTTP.routersOptions,
		IDGenerator:    conf.idGenerator,
		Context: &server.Context{
			RequestIDKey: conf.HTTP.RequestID,
			Logs:         conf.Logs.ctxLogs,
		},
		Compressions:      conf.compressors,
		Mimetypes:         conf.mimetypes,
		ProblemTypePrefix: conf.ProblemTypePrefix,
		Init:              conf.init,
	}

	srv, err := server.New(name, version, opt)
	if err != nil {
		return nil, nil, web.NewStackError(err)
	}

	if len(conf.HTTP.Headers) > 0 {
		srv.UseMiddleware(web.MiddlewareFunc(func(next web.HandlerFunc) web.HandlerFunc {
			return func(ctx *web.Context) web.Responser {
				for _, hh := range conf.HTTP.Headers {
					ctx.Header().Add(hh.Key, hh.Value)
				}
				return next(ctx)
			}
		}))
	}

	srv.OnClose(conf.cleanup...)

	return srv, conf.User, nil
}

func (conf *configOf[T]) SanitizeConfig() *web.FieldError {
	if conf.Logs == nil {
		conf.Logs = &logsConfig{}
	}
	err := conf.Logs.build()
	if err != nil {
		return err.AddFieldParent("logs")
	}
	conf.logs = conf.Logs.logs
	conf.cleanup = conf.Logs.cleanup

	if conf.Language != "" {
		tag, err := language.Parse(conf.Language)
		if err != nil {
			return web.NewFieldError("language.", err)
		}
		conf.languageTag = tag
	}

	if err = conf.buildTimezone(); err != nil {
		return err
	}

	if conf.HTTP == nil {
		conf.HTTP = &httpConfig{}
	}
	if err = conf.HTTP.sanitize(); err != nil {
		return err.AddFieldParent("http")
	}
	conf.http = conf.HTTP.buildHTTPServer()

	if err = conf.sanitizeCompresses(); err != nil {
		return err.AddFieldParent("compressions")
	}

	if err = conf.sanitizeMimetypes(); err != nil {
		return err.AddFieldParent("mimetypes")
	}

	if err = conf.sanitizeFileSerializers(); err != nil {
		return err.AddFieldParent("fileSerializer")
	}

	if conf.IDGenerator == "" {
		conf.IDGenerator = "date"
	}
	if g, found := idGeneratorFactory[conf.IDGenerator]; found {
		f, srv := g()
		conf.idGenerator = f
		if srv != nil {
			conf.init = append(conf.init, func(s web.Server) { s.Services().Add(locales.UniqueIdentityGenerator, srv) })
		}
	}

	if err = conf.buildCache(); err != nil { // buildCache 依赖 IDGenerator
		return err.AddFieldParent("cache")
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
