// SPDX-License-Identifier: MIT

package app

import (
	"io/fs"
	"net/http"
	"time"

	"golang.org/x/text/language"

	"github.com/issue9/web/cache"
	"github.com/issue9/web/internal/errs"
	"github.com/issue9/web/internal/files"
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
	encodings map[string]enc    // 启用的 ID

	// 默认的文件序列化列表
	//
	// 如果为空，表示默认不支持，后续可通过 [server.Server.Files] 进行添加。
	//
	// 可通过 [RegisterFileSerializer] 进行添加额外的序列化方法。默认可用为：
	//  - .yaml
	//  - .yml
	//  - .xml
	//  - .json
	Files []string `yaml:"files,omitempty" json:"files,omitempty" xml:"files>file,omitempty"`
	files map[string]files.FileSerializer

	// 指定可用的 mimetype
	//
	// 如果为空，那么将不支持任何格式的内容输出。
	Mimetypes []*mimetypeConfig `yaml:"mimetypes,omitempty" json:"mimetypes,omitempty" xml:"mimetypes>mimetype,omitempty"`
	mimetypes []mimetype

	// 唯一 ID 生成器
	//
	// 该值由 [RegisterUniqueGenerator] 注册而来，默认情况下，有以下三个选项：
	//  - date 日期格式，默认值；
	//  - string 普通的字符串；
	//  - number 数值格式；
	UniqueGenerator string `yaml:"uniqueGenerator,omitempty" json:"uniqueGenerator,omitempty" xml:"uniqueGenerator,omitempty"`
	uniqueGenerator server.UniqueGenerator

	// 用户自定义的配置项
	User *T `yaml:"user,omitempty" json:"user,omitempty" xml:"user,omitempty"`
}

// ConfigSanitizer 对配置文件的数据验证和修正
type ConfigSanitizer interface {
	SanitizeConfig() *errs.FieldError
}

// NewServerOf 从配置文件初始化 [server.Server] 对象
//
// fsys 项目依赖的文件系统，被用于 [server.Options.FS]，同时也是配置文件所在的目录；
// filename 用于指定项目的配置文件，相对于 fsys 文件系统。
// 序列化方法由 [RegisterFileSerializer] 注册的列表中根据 filename 的扩展名进行查找。
// 如果此值为空，将以 &server.Options{FS: fsys} 初始化 [server.Server]；
//
// T 表示用户自定义的数据项，该数据来自配置文件中的 user 字段。
// 如果实现了 [ConfigSanitizer] 接口，则在加载后进行自检；
func NewServerOf[T any](name, version string, pb server.BuildProblemFunc, fsys fs.FS, filename string) (*server.Server, *T, error) {
	if filename == "" {
		s, err := server.New(name, version, &server.Options{FS: fsys})
		return s, nil, err
	}

	conf, err := loadConfigOf[T](fsys, filename)
	if err != nil {
		return nil, nil, err
	}

	// NOTE: 以下代码由 loadConfigOf 保证不会出错，所以所有错误一律 panic 处理。

	opt := &server.Options{
		FS:              fsys,
		Location:        conf.location,
		Cache:           conf.cache,
		HTTPServer:      conf.http,
		Logs:            conf.logs,
		ProblemBuilder:  pb,
		LanguageTag:     conf.languageTag,
		RoutersOptions:  conf.HTTP.routersOptions,
		UniqueGenerator: conf.uniqueGenerator,
		RequestIDKey:    conf.HTTP.RequestID,
	}

	srv, err := server.New(name, version, opt)
	if err != nil {
		panic(err)
	}

	if len(conf.HTTP.Headers) > 0 {
		srv.Routers().Use(server.MiddlewareFunc(func(next server.HandlerFunc) server.HandlerFunc {
			return func(ctx *server.Context) server.Responser {
				for _, hh := range conf.HTTP.Headers {
					ctx.Header().Add(hh.Key, hh.Value)
				}
				return next(ctx)
			}
		}))
	}

	for name, s := range conf.files {
		srv.Files().Add(s.Marshal, s.Unmarshal, name)
	}

	for _, item := range conf.mimetypes {
		srv.Mimetypes().Add(item.Name, item.Marshal, item.Unmarshal, item.Problem)
	}

	conf.buildEncodings(srv.Encodings())

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

	if err = conf.sanitizeFiles(); err != nil {
		return err.AddFieldParent("files")
	}

	if conf.UniqueGenerator == "" {
		conf.UniqueGenerator = "date"
	}
	if g, found := uniqueGeneratorFactory[conf.UniqueGenerator]; found {
		conf.uniqueGenerator = g()
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
