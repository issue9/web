// SPDX-License-Identifier: MIT

package app

import (
	"io/fs"
	"net/http"
	"time"

	"github.com/issue9/cache"
	"github.com/issue9/localeutil"
	"github.com/issue9/logs/v4"
	"golang.org/x/text/language"

	"github.com/issue9/web/internal/serialization"
	"github.com/issue9/web/server"
)

type configOf[T any] struct {
	XMLName struct{} `yaml:"-" json:"-" xml:"web"`

	// 日志系统的配置项
	//
	// 如果为空，所有日志输出都将被抛弃。
	Logs    *logsConfig `yaml:"logs,omitempty" xml:"logs,omitempty" json:"logs,omitempty"`
	logs    *logs.Logs
	cleanup []func() error

	// 指定默认语言
	//
	// 当客户端未指定 Accept-Language 时，会采用此值，
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
	// 值可以为：memcached，redis，memory 和 file，用户也要以用 RegisterCache 注册新的缓存对象。
	Cache *cacheConfig `yaml:"cache,omitempty" json:"cache,omitempty" xml:"cache,omitempty"`
	cache cache.Cache

	// 压缩的相关配置
	//
	// 如果为空，那么不支持压缩功能。
	// 可通过 RegisterEncoding 注册新的压缩方法，默认可用为 gzip、brotli 和 deflate 三种类型。
	Encodings []*encodingConfig `yaml:"encodings,omitempty" json:"encodings,omitempty" xml:"encodings,omitempty"`
	encodings map[string]enc    // 启用的 ID

	// 默认的文件序列化列表
	//
	// 如果为空，表示默认不支持，后续可通过 Server.Files 进行添加。
	//
	// 可用类型为 .yaml、.yml、.xml 和 .json，可通过 RegisterFileSerializer 进行添加额外的序列化方法。
	Files []string `yaml:"files,omitempty" json:"files,omitempty" xml:"files,omitempty"`
	files map[string]serialization.Item

	// 指定可用的 mimetype
	//
	// 如果为空，那么将不支持任何格式的内容输出。
	Mimetypes []*mimetypeConfig `yaml:"mimetypes,omitempty" json:"mimetypes,omitempty" xml:"mimetype,omitempty"`
	problems  map[string]string

	// 用户自定义的配置项
	User *T `yaml:"user,omitempty" json:"user,omitempty" xml:"user,omitempty"`
}

// NewServerOf 从配置文件初始化 server.Server 实例
//
// fsys 项目依赖的文件系统，被用于 server.Options.FS，同时也是配置文件所在的目录；
// filename 用于指定项目的配置文件，根据扩展由 RegisterFileSerializer 负责在 fsys
// 查找文件加载，如果此值为空，将以 &server.Options{FS: fsys} 作为初始化条件；
//
// T 表示用户自定义的数据项，该数据来自配置文件中的 user 字段。
// 如果实现了 ConfigSanitizer 接口，则在加载后进行自检；
func NewServerOf[T any](name, version string, pb server.BuildProblemFunc, fsys fs.FS, filename string) (*server.Server, *T, error) {
	if filename == "" {
		s, err := server.New(name, version, &server.Options{FS: fsys})
		return s, nil, err
	}

	conf, err := loadConfigOf[T](fsys, filename)
	if err != nil {
		return nil, nil, err
	}

	opt := &server.Options{
		FS:             fsys,
		Location:       conf.location,
		Cache:          conf.cache,
		HTTPServer:     conf.http,
		Logs:           conf.logs,
		ProblemBuilder: pb,
		LanguageTag:    conf.languageTag,
	}
	srv, err := server.New(name, version, opt)
	if err != nil {
		return nil, nil, err
	}

	for name, s := range conf.files {
		if err := srv.Files().Serializer().Add(s.Marshal, s.Unmarshal, name); err != nil {
			return nil, nil, &ConfigError{Message: err, Field: "files", Path: filename}
		}
	}

	if err := conf.buildMimetypes(srv); err != nil {
		err.Field = "mimetypes." + err.Field
		err.Path = filename
		return nil, nil, err
	}

	if err := conf.buildEncodings(srv); err != nil {
		err.Field = "mimetypes." + err.Field
		err.Path = filename
		return nil, nil, err
	}

	srv.OnClose(conf.cleanup...)

	return srv, conf.User, nil
}

func (conf *configOf[T]) sanitize() *ConfigError {
	l, cleanup, err := conf.Logs.build()
	if err != nil {
		err.Field = "logs." + err.Field
		return err
	}
	conf.logs = l
	conf.cleanup = cleanup

	if err = conf.buildCache(); err != nil {
		err.Field = "cache." + err.Field
		return err
	}

	if conf.Language != "" {
		tag, err := language.Parse(conf.Language)
		if err != nil {
			return &ConfigError{Field: "language.", Message: err}
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
		err.Field = "http." + err.Field
		return err
	}
	conf.http = conf.HTTP.buildHTTPServer(conf.logs.StdLogger(logs.LevelError))

	if err = conf.sanitizeEncodings(); err != nil {
		err.Field = "encodings"
		return err
	}

	conf.files = make(map[string]serialization.Item, len(conf.Files))
	for _, name := range conf.Files {
		s, found := filesFactory[name]
		if !found {
			return &ConfigError{Field: "files", Message: localeutil.Error("not found serialization function for %s", name)}
		}
		conf.files[name] = s
	}

	if conf.User != nil {
		if s, ok := (any)(conf.User).(ConfigSanitizer); ok {
			if err := s.SanitizeConfig(); err != nil {
				err.Field = "user." + err.Field
				return err
			}
		}
	}

	return nil
}

func (conf *configOf[T]) buildTimezone() *ConfigError {
	if conf.Timezone == "" {
		return nil
	}

	loc, err := time.LoadLocation(conf.Timezone)
	if err != nil {
		return &ConfigError{Field: "timezone", Message: err}
	}
	conf.location = loc

	return nil
}
