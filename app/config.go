// SPDX-License-Identifier: MIT

package app

import (
	"net/http"
	"time"

	"golang.org/x/text/language"

	"github.com/issue9/config"
	"github.com/issue9/web/cache"
	"github.com/issue9/web/internal/errs"
	"github.com/issue9/web/logs"
	"github.com/issue9/web/server"
)

type configOf[T any] struct {
	XMLName struct{} `yaml:"-" json:"-" xml:"web"`

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
	Encodings []*encodingConfig `yaml:"encodings,omitempty" json:"encodings,omitempty" xml:"encodings>encoding,omitempty"`
	encodings []*server.Encoding

	// 指定配置文件的序列化
	//
	// 如果为空，表示默认不支持，后续可通过 [server.Server.Config] 进行添加。
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
	mimetypes []*server.Mimetype

	// 唯一 ID 生成器
	//
	// 该值由 [RegisterUniqueGenerator] 注册而来，默认情况下，有以下三个选项：
	//  - date 日期格式，默认值；
	//  - string 普通的字符串；
	//  - number 数值格式；
	UniqueGenerator string `yaml:"uniqueGenerator,omitempty" json:"uniqueGenerator,omitempty" xml:"uniqueGenerator,omitempty"`
	uniqueGenerator server.UniqueGenerator

	// 错误代码的配置
	//
	// 可以为空，表示采用 [server.Options] 的默认值。
	Problem  *Problem `yaml:"problem,omitempty" json:"problem,omitempty" xml:"problem,omitempty"`
	problems *server.Problems

	// 用户自定义的配置项
	User *T `yaml:"user,omitempty" json:"user,omitempty" xml:"user,omitempty"`
}

// ConfigSanitizer 对配置文件的数据验证和修正
type ConfigSanitizer interface {
	SanitizeConfig() *errs.FieldError
}

// NewServerOf 从配置文件初始化 [server.Server] 对象
//
// c 项目依赖的文件系统，被用于 [server.Options.Config]，同时也是配置文件所在的目录；
// filename 用于指定项目的配置文件，相对于 fsys 文件系统。
// 序列化方法由 [RegisterFileSerializer] 注册的列表中根据 filename 的扩展名进行查找。
// 如果此值为空，将以 &server.Options{FS: fsys} 初始化 [server.Server]；
//
// T 表示用户自定义的数据项，该数据来自配置文件中的 user 字段。
// 如果实现了 [ConfigSanitizer] 接口，则在加载后进行自检；
func NewServerOf[T any](name, version string, configDir, filename string) (*server.Server, *T, error) {
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
		return nil, nil, errs.NewStackError(err)
	}

	opt := &server.Options{
		Config:     conf.config,
		Location:   conf.location,
		Cache:      conf.cache,
		HTTPServer: conf.http,
		Logs:       conf.logs,
		Locale: &server.Locale{
			Language: conf.languageTag,
		},
		RoutersOptions:  conf.HTTP.routersOptions,
		UniqueGenerator: conf.uniqueGenerator,
		RequestIDKey:    conf.HTTP.RequestID,
		Encodings:       conf.encodings,
		Mimetypes:       conf.mimetypes,
		Problems:        conf.problems,
	}

	srv, err := server.New(name, version, opt)
	if err != nil {
		return nil, nil, errs.NewStackError(err)
	}

	if len(conf.HTTP.Headers) > 0 {
		srv.UseMiddleware(server.MiddlewareFunc(func(next server.HandlerFunc) server.HandlerFunc {
			return func(ctx *server.Context) server.Responser {
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

func (conf *configOf[T]) sanitize() *errs.FieldError {
	l, cleanup, err := conf.Logs.build()
	if err != nil {
		return err.AddFieldParent("logs")
	}
	conf.logs = l
	conf.cleanup = cleanup

	if err = conf.buildCache(); err != nil {
		return err.AddFieldParent("cache")
	}

	if conf.Language != "" {
		tag, err := language.Parse(conf.Language)
		if err != nil {
			return errs.NewFieldError("language.", err)
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

	if err = conf.sanitizeEncodings(); err != nil {
		return err.AddFieldParent("encodings")
	}

	if err = conf.sanitizeMimetypes(); err != nil {
		return err.AddFieldParent("mimetypes")
	}

	if err = conf.sanitizeFileSerializers(); err != nil {
		return err.AddFieldParent("fileSerializer")
	}

	if conf.UniqueGenerator == "" {
		conf.UniqueGenerator = "date"
	}
	if g, found := uniqueGeneratorFactory[conf.UniqueGenerator]; found {
		conf.uniqueGenerator = g()
	}

	if conf.problems, err = conf.Problem.sanitize(); err != nil {
		return err.AddFieldParent("problem")
	}

	if conf.User != nil {
		if s, ok := (any)(conf.User).(ConfigSanitizer); ok {
			if err := s.SanitizeConfig(); err != nil {
				return err.AddFieldParent("user")
			}
		}
	}

	return nil
}

func (conf *configOf[T]) buildTimezone() *errs.FieldError {
	if conf.Timezone == "" {
		return nil
	}

	loc, err := time.LoadLocation(conf.Timezone)
	if err != nil {
		return errs.NewFieldError("timezone", err)
	}
	conf.location = loc

	return nil
}
