// SPDX-License-Identifier: MIT

package web

import (
	"crypto/tls"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"time"

	lc "github.com/issue9/logs/v2/config"
	"github.com/issue9/middleware/v2"
	"golang.org/x/text/message/catalog"

	"github.com/issue9/web/config"
	"github.com/issue9/web/context"
	"github.com/issue9/web/context/mimetype"
	"github.com/issue9/web/internal/filesystem"
)

// 两个配置文件的名称
const (
	LogsFilename   = "logs.xml"
	ConfigFilename = "web.yaml"
)

type (
	// Filter 针对 Context 的中间件
	Filter = context.Filter

	// Middleware 中间件的类型定义
	Middleware = middleware.Middleware
)

// Config 提供了初始化 Web 对象的基本参数
type Config struct {
	XMLName struct{} `yaml:"-" json:"-" xml:"web"`

	// 调试信息的设置
	Debug *Debug `yaml:"debug,omitempty" json:"debug,omitempty" xml:"debug,omitempty"`

	// 网站的根目录所在
	//
	// 比如 https://example.com/api/
	Root  string `yaml:"root,omitempty" json:"root,omitempty" xml:"root,omitempty"`
	url   *url.URL
	addr  string
	isTLS bool

	// 指定插件的搜索方式
	//
	// 通过 glob 语法搜索插件，比如：
	//  ~/plugins/*.so
	// 为空表示没有插件。
	//
	// 当前仅支持部分系统，具体可查看：https://golang.org/pkg/plugin/
	Plugins string `yaml:"plugins,omitempty" json:"plugins,omitempty" xml:"plugins,omitempty"`

	// 网站的域名证书
	//
	// 该设置并不总是生效的，具体的说明可参考 TLSConfig 字段的说明。
	Certificates []*Certificate `yaml:"certificates,omitempty" json:"certificates,omitempty" xml:"certificates,omitempty"`

	// 是否禁用自动生成 OPTIONS 和 HEAD 请求的处理
	DisableOptions bool `yaml:"disableOptions,omitempty" json:"disableOptions,omitempty" xml:"disableOptions,omitempty"`
	DisableHead    bool `yaml:"disableHead,omitempty" json:"disableHead,omitempty" xml:"disableHead,omitempty"`

	// 指定静态内容
	//
	// 键名为 URL 路径，键值为文件地址
	//
	// 比如在 Domain 和 Root 的值分别为 example.com 和 blog 时，
	// 将 Static 的值设置为 /admin ==> ~/data/assets/admin
	// 表示将 example.com/blog/admin/* 解析到 ~/data/assets/admin 目录之下。
	Static Map `yaml:"static,omitempty" json:"static,omitempty" xml:"static,omitempty"`

	// 应用于 http.Server 的几个变量
	ReadTimeout       Duration `yaml:"readTimeout,omitempty" json:"readTimeout,omitempty" xml:"readTimeout,omitempty"`
	WriteTimeout      Duration `yaml:"writeTimeout,omitempty" json:"writeTimeout,omitempty" xml:"writeTimeout,omitempty"`
	IdleTimeout       Duration `yaml:"idleTimeout,omitempty" json:"idleTimeout,omitempty" xml:"idleTimeout,omitempty"`
	ReadHeaderTimeout Duration `yaml:"readHeaderTimeout,omitempty" json:"readHeaderTimeout,omitempty" xml:"readHeaderTimeout,omitempty"`
	MaxHeaderBytes    int      `yaml:"maxHeaderBytes,omitempty" json:"maxHeaderBytes,omitempty" xml:"maxHeaderBytes,omitempty"`

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

	// 用于初始化日志系统的参数
	//
	// 如果你是采用 LoadConfig 加载配置文件的话，此值可以留空，
	// 会自动读取 logs.xml 作为其内容。
	Logs *lc.Config `yaml:"logs,omitempty" json:"logs,omitempty" xml:"logs,omitempty"`

	// 本地化消息的管理组件
	//
	// 为空的情况下会引用 golang.org/x/text/message.DefaultCatalog 对象。
	//
	// golang.org/x/text/message/catalog 提供了 NewBuilder 和 NewFromMap
	// 等方式构建 Catalog 接口实例。
	Catalog catalog.Catalog `yaml:"-" json:"-" xml:"-"`

	// 指定中间件
	//
	// Middlewares 和 Filters 都表示中间件，两者的功能没有本质上的差别。
	// 之所以提供了两个类型，是因为 Middlewares 兼容 http.Handler 类型，
	// 可以对市面上大部分的中间件稍加改造即可使用，而 Filter
	// 则提供了部分 http.Handler 不存在的数据字段，且两者不能交替出现，
	// 二脆同时提供两种中间件。
	//
	// 在使用上，永远是 Middlewares 中的在 Filters 之前调用。
	Middlewares []Middleware `yaml:"-" json:"-" xml:"-"`
	Filters     []Filter     `yaml:"-" json:"-" xml:"-"`

	// 指定各类媒体类型的编解码函数
	Marshalers   map[string]mimetype.MarshalFunc   `yaml:"-" json:"-" xml:"-"`
	Unmarshalers map[string]mimetype.UnmarshalFunc `yaml:"-" json:"-" xml:"-"`

	// 指定生成 Result 的方法
	//
	// 可以为空，表示采用 CTXServer 的默认值。
	ResultBuilder context.BuildResultFunc `yaml:"-" json:"-" xml:"-"`

	// 指定 https 模式下的证书配置项
	//
	// 如果用户指定了 Certificates 字段，则会根据此字段生成，
	// 用户也可以自已覆盖此值，比如采用 golang.org/x/crypto/acme/autocert.Manager.TLSConfig
	// 配置 Let's Encrypt。
	TLSConfig *tls.Config `yaml:"-" json:"-" xml:"-"`

	// 返回给用户的错误提示信息
	//
	// 对键名作了一定的要求：要求最高的三位数必须是一个 HTTP 状态码，
	// 比如 40001，在返回给客户端时，会将 400 作为状态码展示给用户，
	// 同时又会将 40001 和对应的消息发送给用户。
	//
	// 该数据最终由 context.Server.AddMessages 添加。
	Results map[int]string `yaml:"-" json:"-" xml:"-"`
	results map[int]map[int]string

	// 指定用于触发关闭服务的信号
	//
	// 如果为 nil，表示未指定任何信息，如果是长度为 0 的数组，则表示任意信号，
	// 如果指定了多个相同的值，则该信号有可能多次触发。
	ShutdownSignal []os.Signal `yaml:"-" json:"-" xml:"-"`
}

func (conf *Config) sanitize() error {
	if conf.ReadTimeout < 0 {
		return &config.FieldError{Field: "readTimeout", Message: "必须大于等于 0"}
	}

	if conf.WriteTimeout < 0 {
		return &config.FieldError{Field: "writeTimeout", Message: "必须大于等于 0"}
	}

	if conf.IdleTimeout < 0 {
		return &config.FieldError{Field: "idleTimeout", Message: "必须大于等于 0"}
	}

	if conf.ReadHeaderTimeout < 0 {
		return &config.FieldError{Field: "readHeaderTimeout", Message: "必须大于等于 0"}
	}

	if conf.MaxHeaderBytes < 0 {
		return &config.FieldError{Field: "maxHeaderBytes", Message: "必须大于等于 0"}
	}

	if conf.ShutdownTimeout < 0 {
		return &config.FieldError{Field: "shutdownTimeout", Message: "必须大于等于 0"}
	}

	if err := conf.Debug.sanitize(); err != nil {
		return err
	}

	u, err := url.Parse(conf.Root)
	if err != nil {
		return err
	}
	conf.url = u
	if conf.url.Port() == "" {
		switch conf.url.Scheme {
		case "http", "":
			conf.addr = ":80"
		case "https":
			conf.addr = ":443"
			conf.isTLS = true
		default:
			return &config.FieldError{Field: "root", Message: "无效的 scheme"}
		}
	} else {
		conf.addr = ":" + conf.url.Port()
	}

	if err := conf.parseResults(); err != nil {
		return err
	}

	if err := conf.buildTimezone(); err != nil {
		return err
	}

	if err := conf.checkStatic(); err != nil {
		return err
	}

	if conf.isTLS && len(conf.Certificates) == 0 {
		return &config.FieldError{Field: "certificates", Message: "HTTPS 必须指定至少一张证书"}
	}
	return conf.buildTLSConfig()
}

func (conf *Config) parseResults() error {
	conf.results = map[int]map[int]string{}

	for code, msg := range conf.Results {
		if code < 999 {
			return fmt.Errorf("无效的错误代码 %d，必须是 HTTP 状态码的 10 倍以上", code)
		}

		status := code / 10
		for ; status > 999; status /= 10 {
		}

		rslt, found := conf.results[status]
		if found {
			rslt[code] = msg
		} else {
			conf.results[status] = map[int]string{code: msg}
		}
	}

	return nil
}

func (conf *Config) buildTimezone() error {
	if conf.Timezone == "" {
		conf.Timezone = "Local"
	}

	loc, err := time.LoadLocation(conf.Timezone)
	if err != nil {
		return &config.FieldError{Field: "timezone", Message: err.Error()}
	}
	conf.location = loc

	return nil
}

func (conf *Config) checkStatic() (err error) {
	for u, path := range conf.Static {
		if !isURLPath(u) {
			return &config.FieldError{
				Field:   "static." + u,
				Message: "必须以 / 开头且不能以 / 结尾",
			}
		}

		if !filesystem.Exists(path) {
			return &config.FieldError{Field: "static." + u, Message: "对应的路径不存在"}
		}
		conf.Static[u] = path
	}

	return nil
}

func isURLPath(path string) bool {
	return path[0] == '/' && path[len(path)-1] != '/'
}

func (conf *Config) buildTLSConfig() error {
	cfg := &tls.Config{}
	for _, certificate := range conf.Certificates {
		if err := certificate.sanitize(); err != nil {
			return err
		}

		cert, err := tls.LoadX509KeyPair(certificate.Cert, certificate.Key)
		if err != nil {
			return err
		}
		cfg.Certificates = append(cfg.Certificates, cert)
	}

	conf.TLSConfig = cfg
	return nil
}

// LoadConfig 加载指定目录下的配置文件用于初始化 *Config 实例
//
// 当前加载后的 conf.Logs == nil 时，会将 dir 目录下的 logs.xml 当作实例加载。
func LoadConfig(dir string) (conf *Config, err error) {
	confPath := filepath.Join(dir, ConfigFilename)

	conf = &Config{}
	if err = config.LoadFile(confPath, conf); err != nil {
		return nil, err
	}

	if conf.Logs == nil {
		logsPath := filepath.Join(dir, LogsFilename)
		conf.Logs = &lc.Config{}
		if err = config.LoadFile(logsPath, conf.Logs); err != nil {
			return nil, err
		}
	}

	return conf, nil
}
