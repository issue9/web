// SPDX-License-Identifier: MIT

package web

import (
	ctx "context"
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"net/http"
	"net/url"
	"os"
	"os/signal"
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
	XMLName struct{} `yaml:"-" json:"-" xml:"webconfig"`

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
	Headers pairs `yaml:"headers,omitempty" json:"headers,omitempty" xml:"headers,omitempty"`

	// Static 静态内容，键名为 URL 路径，键值为文件地址
	//
	// 比如在 Domain 和 Root 的值分别为 example.com 和 blog 时，
	// 将 Static 的值设置为 /admin ==> ~/data/assets/admin
	// 表示将 example.com/blog/admin/* 解析到 ~/data/assets/admin 目录之下。
	Static pairs `yaml:"static,omitempty" json:"static,omitempty" xml:"static,omitempty"`

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

	Marshalers   map[string]mimetype.MarshalFunc
	Unmarshalers map[string]mimetype.UnmarshalFunc

	ResultBuilder context.BuildResultFunc `yaml:"-" json:"-" xml:"-"`

	ContextInterceptor func(*context.Context) `yaml:"-" json:"-" xml:"-"`

	Logs *logs.Logs `yaml:"-" json:"-" xml:"-" config:".xml,logs.xml"`

	httpServer *http.Server
	ctxServer  *context.Server
	config     *config.Config
	modules    *module.Modules
	closed     chan struct{} // 当 shutdown 延时关闭时，通过此事件确定 Serve() 的返回时机。
}

// Certificate 证书管理
type Certificate struct {
	Cert string `yaml:"cert,omitempty" json:"cert,omitempty" xml:"cert,omitempty"`
	Key  string `yaml:"key,omitempty" json:"key,omitempty" xml:"key,omitempty"`
}

func (cert *Certificate) sanitize() *config.FieldError {
	if !filesystem.Exists(cert.Cert) {
		return &config.FieldError{Field: "cert", Message: "文件不存在"}
	}

	if !filesystem.Exists(cert.Key) {
		return &config.FieldError{Field: "key", Message: "文件不存在"}
	}

	return nil
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

// Grace 指定触发 Shutdown() 的信号，若为空，则任意信号都触发。
//
// 多次调用，则每次指定的信号都会起作用，如果由传递了相同的值，
// 则有可能多次触发 Shutdown()。
//
// NOTE: 传递空值，与不调用，其结果是不同的。
// 若是不调用，则不会处理任何信号；若是传递空值调用，则是处理任何要信号。
func (web *Web) Grace(dur time.Duration, sig ...os.Signal) {
	go func() {
		signalChannel := make(chan os.Signal)
		signal.Notify(signalChannel, sig...)

		<-signalChannel
		signal.Stop(signalChannel)
		close(signalChannel)

		ctx, c := ctx.WithTimeout(ctx.Background(), dur)
		defer c()

		if err := web.Shutdown(ctx); err != nil {
			web.Logs.Error(err)
		}
		web.Logs.Flush() // 保证内容会被正常输出到日志。
	}()
}

func loadConfig(configPath, logsPath string) (web *Web, err error) {
	conf := &config.Config{}

	web = &Web{}
	err = conf.Register("webconfig.yaml", configPath, web, config.LoadYAML, nil)
	if err != nil {
		return nil, err
	}
	if err = conf.Refresh("webconfig.yaml"); err != nil {
		return nil, err
	}

	logsConf := &lc.Config{}
	if err = conf.Register("logs.xml", logsPath, logsConf, config.LoadXML, nil); err != nil {
		return nil, err
	}
	if err = conf.Refresh("logs.xml"); err != nil {
		return nil, err
	}

	l := logs.New()
	if err = l.Init(logsConf); err != nil {
		return nil, err
	}

	web.Logs = l
	web.config = conf

	return web, nil
}

// Classic 返回一个开箱即用的 Web 实例
func Classic(dir string) (*Web, error) {
	web, err := loadConfig(filepath.Join(dir, ConfigFilename), filepath.Join(dir, LogsFilename))
	if err != nil {
		return nil, err
	}

	if err = web.Init(); err != nil {
		return nil, err
	}

	err = web.CTXServer().AddMarshals(map[string]mimetype.MarshalFunc{
		"application/json":       json.Marshal,
		"application/xml":        xml.Marshal,
		mimetype.DefaultMimetype: gob.Marshal,
	})
	if err != nil {
		return nil, err
	}

	err = web.CTXServer().AddUnmarshals(map[string]mimetype.UnmarshalFunc{
		"application/json":       json.Unmarshal,
		"application/xml":        xml.Unmarshal,
		mimetype.DefaultMimetype: gob.Unmarshal,
	})
	if err != nil {
		return nil, err
	}

	return web, nil
}
