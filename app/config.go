// SPDX-License-Identifier: MIT

package app

import (
	"io/fs"
	"net/http"
	"time"

	"github.com/issue9/cache"
	"github.com/issue9/logs/v4"
	"golang.org/x/text/language"

	"github.com/issue9/web/internal/encoding"
	"github.com/issue9/web/serialization"
	"github.com/issue9/web/server"
)

type configOf[T any] struct {
	XMLName struct{} `yaml:"-" json:"-" xml:"web"`

	cleanup []server.CleanupFunc

	// 日志系统的配置项
	//
	// 如果为空，所有日志输出都将被抛弃。
	Logs *logsConfig `yaml:"logs,omitempty" xml:"logs,omitempty" json:"logs,omitempty"`
	logs *logs.Logs

	// 指定默认语言
	//
	// 当客户端未指定 Accept-Language 时，会采用此值，
	// 如果为空，则会尝试当前用户的语言。
	Language    string `yaml:"language,omitempty" json:"language,omitempty" xml:"language,attr,omitempty"`
	languageTag language.Tag

	// 网站端口
	//
	// 格式与 net/http.Server.Addr 相同。可以为空，表示由 net/http.Server 确定其默认值。
	Port string `yaml:"port,omitempty" json:"port,omitempty" xml:"port,attr,omitempty"`

	// 与 HTTP 请求相关的设置项
	HTTP *httpConfig `yaml:"http,omitempty" json:"http,omitempty" xml:"http,omitempty"`

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
	Encodings *encodingsConfig `yaml:"encodings,omitempty" json:"encodings,omitempty" xml:"encodings,omitempty"`
	encoding  *encoding.Encodings

	// 指定可用的 mimetype
	//
	// 如果为空，那么将不支持任何格式的内容输出。
	Mimetypes []*mimetypeConfig `yaml:"mimetypes,omitempty" json:"mimetypes,omitempty" xml:"mimetype,omitempty"`
	mimetypes *serialization.Mimetypes

	// 用户自定义的配置项
	User *T `yaml:"user,omitempty" json:"user,omitempty" xml:"user,omitempty"`
}

// NewOptionsOf 从配置文件初始化 server.Options 实例
//
// 并不是所有的 server.Options 字段都能从 NewOptionsOf 中获得值，
// 像 Mimetypes、ResultBuilder 等可能改变程序行为的字段，
// 并不允许从配置文件中进行修改。
//  opt, user, err := app.NeOptionsOf(...)
//  opt.Mimetypes = serialization.NewMimetypes()
//  opt.Mimetypes.Add(...)
//  srv := server.New("app", "1.0.0", opt)
//
// files 指定从文件到对象的转换方法，同时用于配置文件和翻译内容；
// fsys 项目依赖的文件系统，被用于 server.Options.FS，同时也是配置文件所在的目录；
// filename 用于指定项目的配置文件，根据扩展由 serialization.Files 负责在 fsys 查找文件加载；
//
// T 表示用户自定义的数据项，该数据来自配置文件中的 user 字段。
// 如果实现了 ConfigSanitizer 接口，则在加载后进行自检；
func NewOptionsOf[T any](files *serialization.Files, fsys fs.FS, filename string) (*server.Options, *T, error) {
	conf := &configOf[T]{}
	if err := files.LoadFS(fsys, filename, conf); err != nil {
		return nil, nil, err
	}

	if err := conf.sanitize(); err != nil {
		err.Path = filename
		return nil, nil, err
	}

	h := conf.HTTP
	return &server.Options{
		FS:       fsys,
		Location: conf.location,
		Cache:    conf.cache,
		Port:     conf.Port,
		HTTPServer: func(srv *http.Server) {
			srv.ReadTimeout = h.ReadTimeout.Duration()
			srv.ReadHeaderTimeout = h.ReadHeaderTimeout.Duration()
			srv.WriteTimeout = h.WriteTimeout.Duration()
			srv.IdleTimeout = h.IdleTimeout.Duration()
			srv.MaxHeaderBytes = h.MaxHeaderBytes
			srv.ErrorLog = conf.logs.StdLogger(logs.LevelError)
			srv.TLSConfig = h.tlsConfig
		},
		Logs:            conf.logs,
		FileSerializers: files,
		Encodings:       conf.encoding,
		Mimetypes:       conf.mimetypes,
		LanguageTag:     conf.languageTag,
		Cleanup:         conf.cleanup,
	}, conf.User, nil
}

func (conf *configOf[T]) sanitize() *ConfigError {
	l, cleanup, err := conf.Logs.build()
	if err != nil {
		err.Field = "logs." + err.Field
		return err
	}
	conf.logs = l
	conf.cleanup = append(conf.cleanup, cleanup...)

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

	conf.encoding, err = conf.Encodings.build(l.ERROR())
	if err != nil {
		err.Field = "encodings." + err.Field
		return err
	}

	conf.mimetypes, err = conf.buildMimetypes()
	if err != nil {
		err.Field = "mimetypes." + err.Field
		return err
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
