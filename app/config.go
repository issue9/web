// SPDX-License-Identifier: MIT

package app

import (
	"io/fs"
	"net/http"
	"time"

	"github.com/issue9/logs/v4"
	"golang.org/x/text/language"

	"github.com/issue9/web/serialization"
	"github.com/issue9/web/server"
)

type configOf[T any] struct {
	XMLName struct{} `yaml:"-" json:"-" xml:"web"`

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

	// 忽略的压缩类型
	//
	// 可以有通配符，比如 image/* 表示任意 image/ 开头的内容。
	IgnoreEncodings []string `yaml:"ignoreEncodings,omitempty" json:"ignoreEncodings,omitempty" xml:"ignoreEncoding,omitempty"`

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
	cache server.Cache

	// 用户自定义的配置项
	User *T `yaml:"user,omitempty" json:"user,omitempty" xml:"user,omitempty"`
}

// NewOptionsOf 从配置文件初始化 server.Options 实例
//
// l 日志系统；
// files 指定从文件到对象的转换方法，同时用于配置文件和翻译内容；
// fsys 项目依赖的文件系统，被用于 server.Options.FS，同时也是配置文件所在的目录；
// filename 用于指定项目的配置文件，根据扩展由 serialization.Files 负责在 fsys 查找文件加载；
//
// T 表示用户自定义的数据项，该数据来自配置文件中的 user 字段。
// 如果实现了 ConfigSanitizer 接口，则在加载后进行自检；
//
// NOTE: 并不是所有的 server.Options 字段都是可序列化的，部分字段，比如 RouterOptions
// 需要用户在返回的对象上，自行作修改，当然这些本身有默认值，不修改也可以正常使用。
func NewOptionsOf[T any](l *logs.Logs, files *serialization.Files, fsys fs.FS, filename string) (*server.Options, *T, error) {
	if l == nil {
		panic("l 不能为空值")
	}

	conf := &configOf[T]{}
	if err := files.LoadFS(fsys, filename, conf); err != nil {
		return nil, nil, err
	}

	if err := conf.sanitize(l); err != nil {
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
			srv.ErrorLog = l.StdLogger(logs.LevelError)
			srv.TLSConfig = h.tlsConfig
		},
		Logs:            l,
		FileSerializers: files,
		IgnoreEncodings: conf.IgnoreEncodings,
		LanguageTag:     conf.languageTag,
	}, conf.User, nil
}

func (conf *configOf[T]) sanitize(l *logs.Logs) *ConfigError {
	if err := conf.buildCache(); err != nil {
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

	if err := conf.buildTimezone(); err != nil {
		return err
	}

	if conf.HTTP == nil {
		conf.HTTP = &httpConfig{}
	}
	if err := conf.HTTP.sanitize(); err != nil {
		err.Field = "http." + err.Field
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
