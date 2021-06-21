// SPDX-License-Identifier: MIT

package config

import (
	"io/fs"
	"net/http"
	"time"

	"github.com/issue9/cache"
	"github.com/issue9/logs/v2"
	"github.com/issue9/logs/v2/config"
	"golang.org/x/text/message/catalog"

	"github.com/issue9/web/content"
	"github.com/issue9/web/server"
)

const (
	logsConfigFilename = "logs.xml"
	webconfigFilename  = "web.yaml"
)

// Webconfig 配置内容
type Webconfig struct {
	XMLName struct{} `yaml:"-" json:"-" xml:"web"`

	// 网站端口
	//
	// 格式与 net/http.Server.Addr 相同
	Port string `yaml:"port,omitempty" json:"port,omitempty" xml:"port,attr,omitempty"`

	// 与路由设置相关的配置项
	Router *Router `yaml:"router,omitempty" json:"router,omitempty" xml:"router,omitempty"`

	// 与 HTTP 请求相关的设置项
	HTTP *HTTP `yaml:"http,omitempty" json:"http,omitempty" xml:"http,omitempty"`

	// 指定插件的搜索方式
	//
	// 通过 glob 语法搜索插件，比如：
	//  ~/plugins/*.so
	// 具体可参考：https://golang.org/pkg/path/filepath/#Glob
	// 为空表示没有插件。
	//
	// 当前仅支持部分系统，具体可查看：https://golang.org/pkg/plugin/
	Plugins string `yaml:"plugins,omitempty" json:"plugins,omitempty" xml:"plugins,omitempty"`

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
	Cache *Cache `yaml:"cache,omitempty" json:"cache,omitempty" xml:"cache,omitempty"`
	cache cache.Cache
}

// Router 路由的相关配置
type Router struct {
	DisableHead bool `yaml:"disableHead,omitempty" json:"disableHead,omitempty" xml:"disableHead,attr,omitempty"`
}

// NewServer 从配置文件初始化 Server 实例
func NewServer(name, version string, f fs.FS, c catalog.Catalog, b content.BuildResultFunc) (*server.Server, error) {
	conf := &config.Config{}
	if err := LoadXML(f, logsConfigFilename, conf); err != nil {
		return nil, err
	}

	l := logs.New()
	if err := l.Init(conf); err != nil {
		return nil, err
	}

	webconfig := &Webconfig{}
	if err := LoadYAML(f, webconfigFilename, webconfig); err != nil {
		return nil, err
	}

	return webconfig.NewServer(name, version, f, l, c, b)
}

// NewServer 返回 server.NewServer 对象
func (conf *Webconfig) NewServer(name, version string, fs fs.FS, l *logs.Logs, c catalog.Catalog, f content.BuildResultFunc) (*server.Server, error) {
	// NOTE: 公开此函数，方便第三方将 Webconfig 集成到自己的代码中

	if err := conf.sanitize(); err != nil {
		return nil, err
	}

	h := conf.HTTP
	o := &server.Options{
		Port:          conf.Port,
		FS:            fs,
		Location:      conf.location,
		Cache:         conf.cache,
		DisableHead:   conf.Router.DisableHead,
		Catalog:       c,
		ResultBuilder: f,
		HTTPServer: func(srv *http.Server) {
			srv.ReadTimeout = h.ReadTimeout.Duration()
			srv.ReadHeaderTimeout = h.ReadHeaderTimeout.Duration()
			srv.WriteTimeout = h.WriteTimeout.Duration()
			srv.IdleTimeout = h.IdleTimeout.Duration()
			srv.MaxHeaderBytes = h.MaxHeaderBytes
			srv.ErrorLog = l.ERROR()
			srv.TLSConfig = h.tlsConfig
		},
	}
	srv, err := server.New(name, version, l, o)
	if err != nil {
		return nil, err
	}

	if conf.Plugins != "" {
		if err := srv.LoadPlugins(conf.Plugins); err != nil {
			return nil, err
		}
	}

	return srv, nil
}

func (conf *Webconfig) sanitize() error {
	if err := conf.buildCache(); err != nil {
		err.Field = "cache." + err.Field
		return err
	}

	if conf.Router == nil {
		conf.Router = &Router{}
	}

	if err := conf.buildTimezone(); err != nil {
		return err
	}

	if conf.HTTP == nil {
		conf.HTTP = &HTTP{}
	}
	if err := conf.HTTP.sanitize(); err != nil {
		err.Field = "http." + err.Field
		return err
	}

	return nil
}

func (conf *Webconfig) buildTimezone() error {
	if conf.Timezone == "" {
		conf.Timezone = "Local"
	}

	loc, err := time.LoadLocation(conf.Timezone)
	if err != nil {
		return &Error{Field: "timezone", Message: err.Error()}
	}
	conf.location = loc

	return nil
}

// Duration 封装 time.Duration 以实现对 JSON、XML 和 YAML 的解析
type Duration time.Duration

// Duration 转换成 time.Duration
func (d Duration) Duration() time.Duration {
	return time.Duration(d)
}

// MarshalText encoding.TextMarshaler 接口
func (d Duration) MarshalText() ([]byte, error) {
	return []byte(time.Duration(d).String()), nil
}

// UnmarshalText encoding.TextUnmarshaler 接口
func (d *Duration) UnmarshalText(b []byte) error {
	v, err := time.ParseDuration(string(b))
	if err == nil {
		*d = Duration(v)
	}
	return err
}
