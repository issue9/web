// SPDX-License-Identifier: MIT

package config

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/issue9/cache"
	"github.com/issue9/logs/v2"
	"github.com/issue9/middleware/v3/header"
	"golang.org/x/text/message/catalog"

	"github.com/issue9/web/result"
	"github.com/issue9/web/server"
)

// Webconfig 配置内容
type Webconfig struct {
	XMLName struct{} `yaml:"-" json:"-" xml:"webconfig"`

	// 网站的根目录所在
	//
	// 比如 https://example.com/api/
	Root string `yaml:"root,omitempty" json:"root,omitempty" xml:"root,omitempty"`

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

	// 指定关闭服务时的超时时间
	//
	// 如果此值不为 0，则在关闭服务时会调用 http.Server.Shutdown 函数等待关闭服务，
	// 否则直接采用 http.Server.Close 立即关闭服务。
	ShutdownTimeout Duration `yaml:"shutdownTimeout,omitempty" json:"shutdownTimeout,omitempty" xml:"shutdownTimeout,omitempty"`

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
	// 是否禁用自动生成 OPTIONS 和 HEAD 请求的处理
	DisableOptions bool `yaml:"disableOptions,omitempty" json:"disableOptions,omitempty" xml:"disableOptions,attr,omitempty"`
	DisableHead    bool `yaml:"disableHead,omitempty" json:"disableHead,omitempty" xml:"disableHead,attr,omitempty"`
	SkipCleanPath  bool `yaml:"skipCleanPath,omitempty" json:"skipCleanPath,omitempty" xml:"skipCleanPath,attr,omitempty"`
}

// NewServer 返回 server.NewServer 对象
func (conf *Webconfig) NewServer(name, version string, l *logs.Logs, c catalog.Catalog, f result.BuildFunc) (*server.Server, error) {
	if err := conf.sanitize(); err != nil {
		return nil, err
	}

	o := &server.Options{
		Location:       conf.location,
		Cache:          conf.cache,
		DisableHead:    conf.Router.DisableHead,
		DisableOptions: conf.Router.DisableOptions,
		Catalog:        c,
		ResultBuilder:  f,
		SkipCleanPath:  conf.Router.SkipCleanPath,
		Root:           conf.Root,
		HTTPServer: func(srv *http.Server) {
			srv.ReadTimeout = conf.HTTP.ReadTimeout.Duration()
			srv.ReadHeaderTimeout = conf.HTTP.ReadHeaderTimeout.Duration()
			srv.WriteTimeout = conf.HTTP.WriteTimeout.Duration()
			srv.IdleTimeout = conf.HTTP.IdleTimeout.Duration()
			srv.MaxHeaderBytes = conf.HTTP.MaxHeaderBytes
			srv.ErrorLog = l.ERROR()
			srv.TLSConfig = conf.HTTP.tlsConfig
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

	if conf.HTTP != nil && conf.HTTP.Headers != nil {
		srv.AddMiddlewares(header.New(conf.HTTP.Headers).Middleware)
	}

	return srv, nil
}

func (conf *Webconfig) sanitize() error {
	if conf.ShutdownTimeout < 0 {
		return &Error{Field: "shutdownTimeout", Message: "必须大于等于 0"}
	}

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
	root, err := url.Parse(conf.Root)
	if err != nil {
		return err
	}
	if err := conf.HTTP.sanitize(root); err != nil {
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

// MarshalJSON json.Marshaler 接口
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

// UnmarshalJSON json.Unmarshaler 接口
func (d *Duration) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	tmp, err := time.ParseDuration(s)
	if err != nil {
		return err
	}

	*d = Duration(tmp)
	return nil
}

// MarshalYAML yaml.Marshaler 接口
func (d Duration) MarshalYAML() (interface{}, error) {
	return time.Duration(d).String(), nil
}

// UnmarshalYAML yaml.Unmarshaler 接口
func (d *Duration) UnmarshalYAML(u func(interface{}) error) error {
	var dur time.Duration
	if err := u(&dur); err != nil {
		return err
	}

	*d = Duration(dur)
	return nil
}

// MarshalXML xml.Marshaler 接口
func (d Duration) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(d.Duration().String(), start)
}

// UnmarshalXML xml.Unmarshaler 接口
func (d *Duration) UnmarshalXML(de *xml.Decoder, start xml.StartElement) error {
	var str string
	if err := de.DecodeElement(&str, &start); err != nil && err != io.EOF {
		return err
	}

	dur, err := time.ParseDuration(str)
	if err != nil {
		return err
	}

	*d = Duration(dur)

	return nil
}

// MarshalXMLAttr xml.MarshalerAttr
func (d Duration) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	return xml.Attr{Name: name, Value: d.Duration().String()}, nil
}

// UnmarshalXMLAttr xml.UnmarshalerAttr
func (d *Duration) UnmarshalXMLAttr(attr xml.Attr) error {
	dur, err := time.ParseDuration(attr.Value)
	if err != nil {
		return err
	}

	*d = Duration(dur)
	return nil
}
