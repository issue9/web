// SPDX-License-Identifier: MIT

package web

import (
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/issue9/logs/v2"
	lc "github.com/issue9/logs/v2/config"
	"github.com/issue9/sliceutil"
	"github.com/issue9/web/module"

	"github.com/issue9/web/config"
	"github.com/issue9/web/context"
	"github.com/issue9/web/context/mimetype"
	"github.com/issue9/web/context/mimetype/gob"
	"github.com/issue9/web/internal/filesystem"
)

// 两个配置文件的名称
const (
	LogsFilename   = "logs.xml"
	ConfigFilename = "web.yaml"
)

// Web 项目的配置内容
type Web struct {
	XMLName struct{} `yaml:"-" json:"-" xml:"web"`

	// Debug 是否启用调试模式
	//
	// 该值可能会同时影响多个方面，比如是否启用 Pprof、panic 时的输出处理等
	Debug bool `yaml:"debug,omitempty" json:"debug,omitempty" xml:"debug,attr,omitempty"`

	// Root 网站的根目录所在
	//
	// 比如 https://example.com/api/
	Root  string `yaml:"root,omitempty" json:"root,omitempty" xml:"root,omitempty"`
	url   *url.URL
	addr  string
	isTLS bool

	// Plugins 指定插件，通过 glob 语法指定，比如：~/plugins/*.so
	// 为空表示没有插件。
	//
	// 当前仅支持部分系统：https://golang.org/pkg/plugin/
	Plugins string `yaml:"plugins,omitempty" json:"plugins,omitempty" xml:"plugins,omitempty"`

	// 网站的域名证书
	Certificates []*Certificate `yaml:"certificates,omitempty" json:"certificates,omitempty" xml:"certificates,omitempty"`

	// DisableOptions 是否禁用自动生成 OPTIONS 和 HEAD 请求的处理
	DisableOptions bool `yaml:"disableOptions,omitempty" json:"disableOptions,omitempty" xml:"disableOptions,omitempty"`
	DisableHead    bool `yaml:"disableHead,omitempty" json:"disableHead,omitempty" xml:"disableHead,omitempty"`

	// Headers 附加的报头信息
	//
	// 一些诸如跨域等报头信息，可以在此作设置。
	//
	// 报头信息可能在其它处理器被修改。
	Headers Map `yaml:"headers,omitempty" json:"headers,omitempty" xml:"headers,omitempty"`

	// Static 静态内容，键名为 URL 路径，键值为文件地址
	//
	// 比如在 Domain 和 Root 的值分别为 example.com 和 blog 时，
	// 将 Static 的值设置为 /admin ==> ~/data/assets/admin
	// 表示将 example.com/blog/admin/* 解析到 ~/data/assets/admin 目录之下。
	Static Map `yaml:"static,omitempty" json:"static,omitempty" xml:"static,omitempty"`

	// AllowedDomains 限定访问域名
	//
	// 若指定了此值，则只有此列表中指定的域名可以访问当前网页。
	// 诸如 IP 和其它域名的指向将不再启作用。
	//
	// 在 AllowedDomains 中至少存在一个及以上的域名时，Root 中所指的值会自动添加到此处。
	// AllowedDomains 为空时，并不会限定域名为 Domain 指定的域名。
	AllowedDomains []string `yaml:"allowedDomains,omitempty" json:"allowedDomains,omitempty" xml:"allowedDomains,omitempty"`

	// 应用于 http.Server 的几个变量。
	ReadTimeout       Duration `yaml:"readTimeout,omitempty" json:"readTimeout,omitempty" xml:"readTimeout,omitempty"`
	WriteTimeout      Duration `yaml:"writeTimeout,omitempty" json:"writeTimeout,omitempty" xml:"writeTimeout,omitempty"`
	IdleTimeout       Duration `yaml:"idleTimeout,omitempty" json:"idleTimeout,omitempty" xml:"idleTimeout,omitempty"`
	ReadHeaderTimeout Duration `yaml:"readHeaderTimeout,omitempty" json:"readHeaderTimeout,omitempty" xml:"readHeaderTimeout,omitempty"`
	MaxHeaderBytes    int      `yaml:"maxHeaderBytes,omitempty" json:"maxHeaderBytes,omitempty" xml:"maxHeaderBytes,omitempty"`

	// Timezone 时区名称，可以是 Asia/Shanghai 等，具体可参考：
	// https://en.wikipedia.org/wiki/List_of_tz_database_time_zones
	//
	// 为空和 Local(注意大小写) 值都会被初始化本地时间。
	Timezone string         `yaml:"timezone,omitempty" json:"timezone,omitempty" xml:"timezone,omitempty"`
	Location *time.Location `yaml:"-" json:"-" xml:"-"`

	Marshalers         map[string]mimetype.MarshalFunc   `yaml:"-" json:"-" xml:"-"`
	Unmarshalers       map[string]mimetype.UnmarshalFunc `yaml:"-" json:"-" xml:"-"`
	ResultBuilder      context.BuildResultFunc           `yaml:"-" json:"-" xml:"-"`
	ContextInterceptor func(*context.Context)            `yaml:"-" json:"-" xml:"-"`

	// 用于初始化日志系统的参数
	LogsConfig *lc.Config `yaml:"-" json:"-" xml:"-"`

	// 指定用于触发关闭服务的信号
	//
	// 如果为 nil，表示未指定任何信息，如果是长度为 0 的数组，则表示任意信号，
	// 如果指定了多个相同的值，则该信号有可能多次触发。
	ShutdownSignal []os.Signal `yaml:"-" json:"-" xml:"-"`

	// 指定关闭服务时的超时时间
	//
	// 如果此值不为 0，则在关闭服务时会调用 http.Server.Shutdown 函数等待关闭服务，
	// Shutdown 即为该方法的等待时间。
	ShutdownTimeout Duration `yaml:"shutdownTimeout,omitempty" json:"shutdownTimeout,omitempty" xml:"shutdownTimeout,omitempty"`

	logs       *logs.Logs
	httpServer *http.Server
	ctxServer  *context.Server
	modServer  *module.Server
	closed     chan struct{} // 当 shutdown 延时关闭时，通过此事件确定 Serve() 的返回时机。
}

func (web *Web) sanitize() (err error) {
	if web.ReadTimeout < 0 {
		return &config.FieldError{Field: "readTimeout", Message: "必须大于等于 0"}
	}

	if web.WriteTimeout < 0 {
		return &config.FieldError{Field: "writeTimeout", Message: "必须大于等于 0"}
	}

	if web.IdleTimeout < 0 {
		return &config.FieldError{Field: "idleTimeout", Message: "必须大于等于 0"}
	}

	if web.ReadHeaderTimeout < 0 {
		return &config.FieldError{Field: "readHeaderTimeout", Message: "必须大于等于 0"}
	}

	if web.MaxHeaderBytes < 0 {
		return &config.FieldError{Field: "maxHeaderBytes", Message: "必须大于等于 0"}
	}

	if web.url, err = url.Parse(web.Root); err != nil {
		return err
	}
	web.addr = ":" + web.url.Port()
	if web.addr == "" {
		if web.url.Scheme == "http" {
			web.addr = ":80"
		} else if web.url.Scheme == "https" {
			web.addr = ":443"
			web.isTLS = true
		}
	}

	if err := web.buildTimezone(); err != nil {
		return err
	}

	if err := web.checkStatic(); err != nil {
		return err
	}

	for _, c := range web.Certificates {
		if err := c.sanitize(); err != nil {
			return err
		}
	}

	return web.buildAllowedDomains()
}

func (web *Web) buildTimezone() error {
	if web.Timezone == "" {
		web.Timezone = "Local"
	}

	loc, err := time.LoadLocation(web.Timezone)
	if err != nil {
		return &config.FieldError{Field: "timezone", Message: err.Error()}
	}
	web.Location = loc

	return nil
}

func (web *Web) checkStatic() (err error) {
	for u, path := range web.Static {
		if !isURLPath(u) {
			return &config.FieldError{
				Field:   "static." + u,
				Message: "必须以 / 开头且不能以 / 结尾",
			}
		}

		if !filepath.IsAbs(path) {
			path, err = filepath.Abs(path)
			if err != nil {
				return &config.FieldError{Field: "static." + u, Message: err.Error()}
			}
		}

		if !filesystem.Exists(path) {
			return &config.FieldError{Field: "static." + u, Message: "对应的路径不存在"}
		}
		web.Static[u] = path
	}

	return nil
}

func isURLPath(path string) bool {
	return path[0] == '/' && path[len(path)-1] != '/'
}

func (web *Web) buildAllowedDomains() error {
	if len(web.AllowedDomains) == 0 {
		return nil
	}

	hostname := web.url.Hostname()
	if hostname != "" {
		cnt := sliceutil.Count(web.AllowedDomains, func(i int) bool { return web.AllowedDomains[i] == hostname })
		if cnt == 0 {
			web.AllowedDomains = append(web.AllowedDomains, hostname)
		}
	}
	return nil
}

func (web *Web) toTLSConfig() (*tls.Config, error) {
	cfg := &tls.Config{}
	for _, certificate := range web.Certificates {
		cert, err := tls.LoadX509KeyPair(certificate.Cert, certificate.Key)
		if err != nil {
			return nil, err
		}
		cfg.Certificates = append(cfg.Certificates, cert)
	}

	return cfg, nil
}

func loadConfig(configPath, logsPath string) (web *Web, err error) {
	web = &Web{}
	if err = config.LoadFile(configPath, web); err != nil {
		return nil, err
	}

	web.LogsConfig = &lc.Config{}
	if err = config.LoadFile(logsPath, web.LogsConfig); err != nil {
		return nil, err
	}

	return web, nil
}

// Classic 返回一个开箱即用的 Web 实例
func Classic(dir string) (*Web, error) {
	web, err := loadConfig(filepath.Join(dir, ConfigFilename), filepath.Join(dir, LogsFilename))
	if err != nil {
		return nil, err
	}

	web.Marshalers = map[string]mimetype.MarshalFunc{
		"application/json":       json.Marshal,
		"application/xml":        xml.Marshal,
		mimetype.DefaultMimetype: gob.Marshal,
	}

	web.Unmarshalers = map[string]mimetype.UnmarshalFunc{
		"application/json":       json.Unmarshal,
		"application/xml":        xml.Unmarshal,
		mimetype.DefaultMimetype: gob.Unmarshal,
	}

	web.ResultBuilder = context.DefaultResultBuilder

	return web, nil
}
