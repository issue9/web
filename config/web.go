// SPDX-License-Identifier: MIT

package config

import (
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/issue9/cache"
	"github.com/issue9/logs/v2"
	lc "github.com/issue9/logs/v2/config"
	"github.com/issue9/middleware/v2"
	"github.com/issue9/middleware/v2/errorhandler"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"

	"github.com/issue9/web/context"
	"github.com/issue9/web/context/contentype"
	"github.com/issue9/web/context/contentype/gob"
	"github.com/issue9/web/context/result"
	"github.com/issue9/web/internal/filesystem"
)

type (
	// Filter 针对 Context 的中间件
	Filter = context.Filter

	// Middleware 中间件的类型定义
	Middleware = middleware.Middleware

	// ErrorHandlerFunc 错误状态码对应的处理函数原型
	ErrorHandlerFunc = errorhandler.HandleFunc

	// Web 提供了初始化 Server 对象的基本参数
	Web struct {
		XMLName struct{} `yaml:"-" json:"-" xml:"web"`

		// 调试信息的设置
		Debug *Debug `yaml:"debug,omitempty" json:"debug,omitempty" xml:"debug,omitempty"`

		// 网站的根目录所在
		//
		// 比如 https://example.com/api/
		Root string `yaml:"root,omitempty" json:"root,omitempty" xml:"root,omitempty"`

		// 指定插件的搜索方式
		//
		// 通过 glob 语法搜索插件，比如：
		//  ~/plugins/*.so
		// 具体可参考：https://golang.org/pkg/path/filepath/#Glob
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
		SkipCleanPath  bool `yaml:"skipCleanPath,omitempty" json:"skipCleanPath,omitempty" xml:"skipCleanPath,omitempty"`

		// 指定静态内容
		//
		// 键名为 URL 路径，键值为文件地址
		//
		// 在 Root 的值为 example.com/blog 时，
		// 将 Static 的值设置为 /admin/{path} ==> ~/data/assets/admin
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

		// Cache 指定缓存对象
		//
		// 可看查看 github.com/issue9/cache 中相关的实现，
		// 用户也可以自己实现 github.com/issue9/cache.Cache 接口。
		//
		// 如果用户未指定，则会采用以下方式提供默认值：
		//  github.com/issue9/cache/memory.New(24 * time.Hour)
		Cache cache.Cache `yaml:"-" json:"-" xml:"-"`

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
		// 在使用上，永远是 Middlewares 在 Filters 之前调用。
		Middlewares []Middleware `yaml:"-" json:"-" xml:"-"`
		Filters     []Filter     `yaml:"-" json:"-" xml:"-"`

		// 指定各类媒体类型的编解码函数
		Marshalers   map[string]contentype.MarshalFunc   `yaml:"-" json:"-" xml:"-"`
		Unmarshalers map[string]contentype.UnmarshalFunc `yaml:"-" json:"-" xml:"-"`

		// 指定生成 Result 的方法
		//
		// 可以为空，表示采用 CTXServer 的默认值。
		ResultBuilder result.BuildFunc `yaml:"-" json:"-" xml:"-"`

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
		// 该数据最终由 context.Server.AddMessage 添加。
		Results map[int]Locale `yaml:"-" json:"-" xml:"-"`
		results map[int]map[int]Locale

		// 指定错误页面的处理方式
		ErrorHandlers []*ErrorHandler `yaml:"-" json:"-" xml:"-"`

		// 指定用于触发关闭服务的信号
		//
		// 如果为 nil，表示未指定任何信息，如果是长度为 0 的数组，则表示任意信号，
		// 如果指定了多个相同的值，则该信号有可能多次触发。
		ShutdownSignal []os.Signal `yaml:"-" json:"-" xml:"-"`
	}

	// Locale 用于描述本地化信息
	Locale struct {
		Key  message.Reference
		vals []interface{}
	}

	// Map 定义 map[string]string 类型
	//
	// 唯一的功能是为了 xml 能支持 map。
	Map map[string]string

	entry struct {
		XMLName struct{} `xml:"key"`
		Name    string   `xml:"name,attr"`
		Value   string   `xml:",chardata"`
	}

	// Debug 调试信息的配置
	Debug struct {
		Pprof string `yaml:"pprof,omitempty" json:"pprof,omitempty" xml:"pprof,omitempty"`
		Vars  string `yaml:"vars,omitempty" json:"vars,omitempty" xml:"vars,omitempty"`
	}

	// ErrorHandler 错误处理的配置
	ErrorHandler struct {
		Status  []int
		Handler ErrorHandlerFunc
	}

	// Duration 封装 time.Duration 以实现对 JSON、XML 和 YAML 的解析
	Duration time.Duration

	// Certificate 证书管理
	Certificate struct {
		Cert string `yaml:"cert,omitempty" json:"cert,omitempty" xml:"cert,omitempty"`
		Key  string `yaml:"key,omitempty" json:"key,omitempty" xml:"key,omitempty"`
	}
)

// Classic 返回一个开箱即用的 Server 实例
func Classic(logConfigFile, configFile string) (*context.Server, error) {
	logConf := &lc.Config{}
	if err := LoadFile(logConfigFile, logConf); err != nil {
		return nil, err
	}
	if err := logConf.Sanitize(); err != nil {
		return nil, err
	}

	l := logs.New()
	if err := l.Init(logConf); err != nil {
		return nil, err
	}

	web := &Web{}
	if err := LoadFile(configFile, web); err != nil {
		return nil, err
	}

	web.Marshalers = map[string]contentype.MarshalFunc{
		"application/json":         json.Marshal,
		"application/xml":          xml.Marshal,
		contentype.DefaultMimetype: gob.Marshal,
	}

	web.Unmarshalers = map[string]contentype.UnmarshalFunc{
		"application/json":         json.Unmarshal,
		"application/xml":          xml.Unmarshal,
		contentype.DefaultMimetype: gob.Unmarshal,
	}

	web.Results = map[int]Locale{
		40001: {Key: "无效的报头"},
		40002: {Key: "无效的地址"},
		40003: {Key: "无效的查询参数"},
		40004: {Key: "无效的报文"},
	}

	return web.NewServer(l)
}

// NewServer 返回 Server 对象
func (web *Web) NewServer(l *logs.Logs) (*context.Server, error) {
	if err := web.sanitize(); err != nil {
		return nil, err
	}

	srv, err := web.toCTXServer(l)
	if err != nil {
		return nil, err
	}

	if web.ShutdownSignal != nil {
		grace(srv, web.ShutdownTimeout.Duration(), web.ShutdownSignal...)
	}

	return srv, nil
}

func (web *Web) toCTXServer(l *logs.Logs) (*context.Server, error) {
	o := &context.Options{
		Location:       web.location,
		Cache:          web.Cache,
		DisableHead:    web.DisableHead,
		DisableOptions: web.DisableOptions,
		Catalog:        web.Catalog,
		ResultBuilder:  web.ResultBuilder,
		SkipCleanPath:  web.SkipCleanPath,
		Root:           web.Root,
		HTTPServer: func(srv *http.Server) {
			srv.ReadTimeout = web.ReadTimeout.Duration()
			srv.ReadHeaderTimeout = web.ReadHeaderTimeout.Duration()
			srv.WriteTimeout = web.WriteTimeout.Duration()
			srv.IdleTimeout = web.IdleTimeout.Duration()
			srv.MaxHeaderBytes = web.MaxHeaderBytes
			srv.ErrorLog = l.ERROR()
			srv.TLSConfig = web.TLSConfig
		},
	}
	srv, err := context.NewServer(l, o)
	if err != nil {
		return nil, err
	}

	for path, dir := range web.Static {
		if err := srv.Router().Static(path, dir); err != nil {
			return nil, err
		}
	}

	if err = srv.Mimetypes().AddMarshals(web.Marshalers); err != nil {
		return nil, err
	}
	if err = srv.Mimetypes().AddUnmarshals(web.Unmarshalers); err != nil {
		return nil, err
	}

	for status, rslt := range web.results {
		for code, l := range rslt {
			srv.AddMessage(status, code, l.Key, l.vals...)
		}
	}

	if web.Debug != nil {
		srv.SetDebugger(web.Debug.Pprof, web.Debug.Vars)
	}

	if len(web.Middlewares) > 0 {
		srv.AddMiddlewares(web.Middlewares...)
	}
	if len(web.Filters) > 0 {
		srv.AddFilters(web.Filters...)
	}

	for _, h := range web.ErrorHandlers {
		srv.SetErrorHandle(h.Handler, h.Status...)
	}

	if web.Plugins != "" {
		if err := srv.LoadPlugins(web.Plugins); err != nil {
			return nil, err
		}
	}

	return srv, nil
}

func grace(srv *context.Server, shutdownTimeout time.Duration, sig ...os.Signal) {
	go func() {
		signalChannel := make(chan os.Signal)
		signal.Notify(signalChannel, sig...)

		<-signalChannel
		signal.Stop(signalChannel)
		close(signalChannel)

		if err := srv.Close(shutdownTimeout); err != nil {
			srv.Logs().Error(err)
		}
		srv.Logs().Flush() // 保证内容会被正常输出到日志。
	}()
}

func (web *Web) sanitize() error {
	if web.ReadTimeout < 0 {
		return &FieldError{Field: "readTimeout", Message: "必须大于等于 0"}
	}

	if web.WriteTimeout < 0 {
		return &FieldError{Field: "writeTimeout", Message: "必须大于等于 0"}
	}

	if web.IdleTimeout < 0 {
		return &FieldError{Field: "idleTimeout", Message: "必须大于等于 0"}
	}

	if web.ReadHeaderTimeout < 0 {
		return &FieldError{Field: "readHeaderTimeout", Message: "必须大于等于 0"}
	}

	if web.MaxHeaderBytes < 0 {
		return &FieldError{Field: "maxHeaderBytes", Message: "必须大于等于 0"}
	}

	if web.ShutdownTimeout < 0 {
		return &FieldError{Field: "shutdownTimeout", Message: "必须大于等于 0"}
	}

	if web.Debug != nil {
		if err := web.Debug.sanitize(); err != nil {
			return err
		}
	}

	if err := web.parseResults(); err != nil {
		return err
	}

	if err := web.buildTimezone(); err != nil {
		return err
	}

	if err := web.checkStatic(); err != nil {
		return err
	}

	u, err := url.Parse(web.Root)
	if err != nil {
		return err
	}
	if u.Scheme == "https" && len(web.Certificates) == 0 {
		return &FieldError{Field: "certificates", Message: "HTTPS 必须指定至少一张证书"}
	}
	return web.buildTLSConfig()
}

func (web *Web) parseResults() error {
	web.results = map[int]map[int]Locale{}

	for code, msg := range web.Results {
		if code < 999 {
			return fmt.Errorf("无效的错误代码 %d，必须是 HTTP 状态码的 10 倍以上", code)
		}

		status := code / 10
		for ; status > 999; status /= 10 {
		}

		rslt, found := web.results[status]
		if found {
			rslt[code] = msg
		} else {
			web.results[status] = map[int]Locale{code: msg}
		}
	}

	return nil
}

func (web *Web) buildTimezone() error {
	if web.Timezone == "" {
		web.Timezone = "Local"
	}

	loc, err := time.LoadLocation(web.Timezone)
	if err != nil {
		return &FieldError{Field: "timezone", Message: err.Error()}
	}
	web.location = loc

	return nil
}

func (web *Web) checkStatic() (err error) {
	for u, path := range web.Static {
		if !isURLPath(u) {
			return &FieldError{
				Field:   "static." + u,
				Message: "必须以 / 开头且不能以 / 结尾",
			}
		}

		if !filesystem.Exists(path) {
			return &FieldError{Field: "static." + u, Message: "对应的路径不存在"}
		}
		web.Static[u] = path
	}

	return nil
}

func isURLPath(path string) bool {
	return path[0] == '/' && path[len(path)-1] != '/'
}

func (web *Web) buildTLSConfig() error {
	cfg := &tls.Config{}
	for _, certificate := range web.Certificates {
		if err := certificate.sanitize(); err != nil {
			return err
		}

		cert, err := tls.LoadX509KeyPair(certificate.Cert, certificate.Key)
		if err != nil {
			return err
		}
		cfg.Certificates = append(cfg.Certificates, cert)
	}

	web.TLSConfig = cfg
	return nil
}

// MarshalXML implement xml.Marshaler
func (p Map) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if len(p) == 0 {
		return nil
	}

	if err := e.EncodeToken(start); err != nil {
		return err
	}

	for k, v := range p {
		if err := e.Encode(entry{Name: k, Value: v}); err != nil {
			return err
		}
	}

	return e.EncodeToken(start.End())
}

// UnmarshalXML implement xml.Unmarshaler
func (p *Map) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	*p = Map{}

	for {
		e := &entry{}
		if err := d.Decode(e); errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return err
		}

		(*p)[e.Name] = e.Value
	}

	return nil
}

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
	return e.EncodeElement(time.Duration(d).String(), start)
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

func (cert *Certificate) sanitize() *FieldError {
	if !filesystem.Exists(cert.Cert) {
		return &FieldError{Field: "cert", Message: "文件不存在"}
	}

	if !filesystem.Exists(cert.Key) {
		return &FieldError{Field: "key", Message: "文件不存在"}
	}

	return nil
}

func (dbg *Debug) sanitize() *FieldError {
	if dbg.Pprof != "" && (dbg.Pprof[0] != '/' || dbg.Pprof[len(dbg.Pprof)-1] != '/') {
		return &FieldError{Field: "pprof", Message: "必须以 / 开始和结束"}
	}

	if dbg.Vars != "" && dbg.Vars[0] != '/' {
		return &FieldError{Field: "vars", Message: "必须以 / 开头"}
	}

	return nil
}
