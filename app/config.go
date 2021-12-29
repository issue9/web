// SPDX-License-Identifier: MIT

package app

import (
	"io/fs"
	"net/http"
	"time"

	"github.com/issue9/cache"
	"github.com/issue9/logs/v3"
	"github.com/issue9/logs/v3/config"
	"golang.org/x/text/language"

	"github.com/issue9/web/serialization"
	"github.com/issue9/web/server"
)

type webconfig[T any] struct {
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

	// 路由的相关设置
	//
	// 提供了对全局路由的设置，但是用户依然可以通过 server.Server.MuxGroups().AddRouter 忽略这些设置项。
	Router *router `yaml:"router,omitempty" json:"router,omitempty" xml:"router,omitempty"`

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
	Cache *cacheConfig `yaml:"cache,omitempty" json:"cache,omitempty" xml:"cache,omitempty"`
	cache cache.Cache

	// 日志初始化参数
	//
	// 如果为空，则初始化一个空日志，不会输出任何日志。
	Logs *config.Config `yaml:"logs,omitempty" json:"logs,omitempty" xml:"logs,omitempty"`
	logs *logs.Logs

	// 用户自定义的配置项
	User *T `yaml:"user,omitempty" json:"user,omitempty" xml:"user,omitempty"`
}

// NewOptions 从配置文件初始化 server.Options 实例
//
// files 指定从文件到对象的转换方法，同时用于配置文件和翻译内容；
// filename 用于指定项目的配置文件，根据扩展由 serialization.Files 负责在 f 查找文件加载；
//
// NOTE: 并不是所有的 server.Options 字段都是可序列化的，部分字段，比如 RouterOptions
// 需要用户在返回的对象上，自行作修改，当然这些本身有默认值，不修改也可以正常使用。
//
// T 表示用户自定义的数据项，可以实现 Sanitizer 接口，用于对加载后的数据进行自检。
func NewOptions[T any](files *serialization.Files, f fs.FS, filename string) (*server.Options, *T, error) {
	conf := &webconfig[T]{}
	if err := files.LoadFS(f, filename, conf); err != nil {
		return nil, nil, err
	}

	if err := conf.sanitize(); err != nil {
		if err2, ok := err.(*Error); ok {
			err2.Config = filename
		}
		return nil, nil, err
	}

	h := conf.HTTP
	return &server.Options{
		Port:          conf.Port,
		FS:            f,
		Location:      conf.location,
		Cache:         conf.cache,
		RouterOptions: conf.Router.options,
		HTTPServer: func(srv *http.Server) {
			srv.ReadTimeout = h.ReadTimeout.Duration()
			srv.ReadHeaderTimeout = h.ReadHeaderTimeout.Duration()
			srv.WriteTimeout = h.WriteTimeout.Duration()
			srv.IdleTimeout = h.IdleTimeout.Duration()
			srv.MaxHeaderBytes = h.MaxHeaderBytes
			srv.ErrorLog = conf.logs.ERROR()
			srv.TLSConfig = h.tlsConfig
		},
		Logs:  conf.logs,
		Files: files,
	}, conf.User, nil
}

func (conf *webconfig[T]) sanitize() error {
	if conf.Logs != nil {
		if err := conf.Logs.Sanitize(); err != nil {
			return &Error{Field: "logs", Message: err}
		}
	}
	l, err := logs.New(conf.Logs)
	if err != nil {
		return err
	}
	conf.logs = l

	if err := conf.buildCache(); err != nil {
		err.Field = "cache." + err.Field
		return err
	}

	if conf.Language != "" {
		tag, err := language.Parse(conf.Language)
		if err != nil {
			return &Error{Field: "language", Message: err}
		}
		conf.languageTag = tag
	}

	if err := conf.buildTimezone(); err != nil {
		return err
	}

	if conf.Router == nil {
		conf.Router = &router{}
	}
	if err := conf.Router.sanitize(); err != nil {
		err.Field = "router." + err.Field
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
		if s, ok := (interface{})(conf.User).(Sanitizer); ok {
			if err := s.Sanitize(); err != nil {
				err.Field = "user." + err.Field
				return err
			}
		}
	}

	return nil
}

func (conf *webconfig[T]) buildTimezone() *Error {
	if conf.Timezone == "" {
		return nil
	}

	loc, err := time.LoadLocation(conf.Timezone)
	if err != nil {
		return &Error{Field: "timezone", Message: err}
	}
	conf.location = loc

	return nil
}
