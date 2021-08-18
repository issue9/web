// SPDX-License-Identifier: MIT

package config

import (
	"io/fs"
	"net/http"
	"time"

	"github.com/issue9/cache"
	"github.com/issue9/logs/v3"
	"github.com/issue9/logs/v3/config"

	"github.com/issue9/web/content"
	"github.com/issue9/web/server"
)

// Webconfig 配置内容
type Webconfig struct {
	XMLName struct{} `yaml:"-" json:"-" xml:"web"`

	// 网站端口
	//
	// 格式与 net/http.Server.Addr 相同
	Port string `yaml:"port,omitempty" json:"port,omitempty" xml:"port,attr,omitempty"`

	// 是否禁用自动生成 HEAD 请求
	DisableHead bool `yaml:"disableHead,omitempty" json:"disableHead,omitempty" xml:"disableHead,attr,omitempty"`

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

	// 此处列出的类型将不会被压缩
	//
	// 可以带 *，比如 text/* 表示所有 mime-type 为 text/ 开始的类型。
	IgnoreCompressTypes []string `yaml:"ignoreCompressTypes,omitempty" json:"ignoreCompressTypes,omitempty" xml:"ignoreCompressTypes,omitempty"`
}

// NewServer 从配置文件初始化 Server 实例
func NewServer(name, version string, b content.BuildResultFunc, f fs.FS, logsFilename, webFilename string) (*server.Server, error) {
	conf := &config.Config{}
	if err := LoadXML(f, logsFilename, conf); err != nil {
		return nil, err
	}

	l, err := logs.New(conf)
	if err != nil {
		return nil, err
	}

	webconfig := &Webconfig{}
	if err := LoadYAML(f, webFilename, webconfig); err != nil {
		return nil, err
	}

	return webconfig.NewServer(name, version, f, l, b)
}

// NewServer 返回 server.NewServer 对象
func (conf *Webconfig) NewServer(name, version string, fs fs.FS, l *logs.Logs, f content.BuildResultFunc) (*server.Server, error) {
	// NOTE: 公开此函数，方便第三方将 Webconfig 集成到自己的代码中

	if err := conf.sanitize(l); err != nil {
		return nil, err
	}

	h := conf.HTTP
	o := &server.Options{
		Port:          conf.Port,
		FS:            fs,
		Location:      conf.location,
		Cache:         conf.cache,
		DisableHead:   conf.DisableHead,
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
		Logs:                l,
		IgnoreCompressTypes: conf.IgnoreCompressTypes,
	}

	srv, err := server.New(name, version, o)
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

func (conf *Webconfig) sanitize(l *logs.Logs) error {
	if err := conf.buildCache(l); err != nil {
		err.Field = "cache." + err.Field
		return err
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
	if conf.Timezone != "" {
		return nil
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
