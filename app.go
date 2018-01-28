// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	stdctx "context"
	"errors"
	"net/http"
	"net/http/pprof"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/issue9/logs"
	"github.com/issue9/middleware/compress"
	"github.com/issue9/middleware/host"
	"github.com/issue9/middleware/recovery"
	"github.com/issue9/mux"
	charset "golang.org/x/text/encoding"

	"github.com/issue9/web/context"
	"github.com/issue9/web/encoding"
	"github.com/issue9/web/internal/config"
)

const pprofPath = "/debug/pprof/"

const (
	logsFilename   = "logs.xml" // 日志配置文件的文件名。
	configFilename = "web.yaml" // 配置文件的文件名。
)

// App 保存整个程序的运行环境，方便做整体的调度。
type App struct {
	configDir string
	config    *config.Config
	modules   []*Module
	server    *http.Server

	middleware Middleware // 应用于全局所有路由项的中间件
	mux        *mux.Mux
	router     *mux.Prefix

	// 根据配置文件，获取相应的输出编码和字符集。
	outputEncoding encoding.MarshalFunc
	outputCharset  charset.Encoding
}

// NewApp 初始化框架的基本内容。
func NewApp(configDir string, m Middleware) (*App, error) {
	dir, err := filepath.Abs(configDir)
	if err != nil {
		return nil, err
	}

	app := &App{
		configDir:  dir,
		middleware: m,
		modules:    make([]*Module, 0, 100),
	}

	if err := logs.InitFromXMLFile(app.File(logsFilename)); err != nil {
		return nil, err
	}

	if err = app.loadConfig(); err != nil {
		return nil, err
	}

	return app, nil
}

func (app *App) loadConfig() error {
	conf, err := config.Load(app.File(configFilename))
	if err != nil {
		return err
	}

	app.config = conf
	app.mux = mux.New(conf.DisableOptions, false, nil, nil)
	app.router = app.mux.Prefix(conf.Root)

	app.outputEncoding = encoding.Marshal(conf.OutputEncoding)
	if app.outputEncoding == nil {
		return errors.New("未找到 outputEncoding")
	}

	app.outputCharset = encoding.Charset(conf.OutputCharset)
	if app.outputCharset == nil {
		return errors.New("未找到 outputCharset")
	}

	return nil
}

// IsDebug 是否处在调试模式
func (app *App) IsDebug() bool {
	return app.config.Debug
}

// File 获取配置目录下的文件名
func (app *App) File(path ...string) string {
	paths := make([]string, 0, len(path)+1)
	paths = append(paths, app.configDir)
	return filepath.Join(append(paths, path...)...)
}

// Run 加载各个模块的数据，运行路由，执行监听程序。
//
// 必须得保证在调用 Run() 时，logs 包的所有功能是可用的，
// 之后的好多操作，都会将日志输出 logs 中的相关通道中。
func (app *App) Run() error {
	// 插件作为模块的一种实现方式，要在依赖关系之前加载
	if err := app.loadPlugins(); err != nil {
		return err
	}

	// 初始化各个模块之间的依赖关系
	if err := app.initDependency(); err != nil {
		return err
	}

	// 静态文件路由，在其它路由构建之前调用
	for url, dir := range app.config.Static {
		pattern := url + "{path}"
		app.router.Get(pattern, http.StripPrefix(url, compress.New(http.FileServer(http.Dir(dir)), logs.ERROR())))
	}

	h := app.buildHandler(app.mux)
	if app.middleware != nil {
		h = app.middleware(h)
	}

	app.server = &http.Server{
		Addr:         ":" + strconv.Itoa(app.config.Port),
		Handler:      h,
		ErrorLog:     logs.ERROR(),
		ReadTimeout:  app.config.ReadTimeout,
		WriteTimeout: app.config.WriteTimeout,
	}

	logs.Infof("开始监听[%v]端口", app.config.Port)

	if !app.config.HTTPS {
		return app.server.ListenAndServe()
	}

	return app.server.ListenAndServeTLS(app.config.CertFile, app.config.KeyFile)
}

// Close 立即关闭服务
func (app *App) Close() error {
	logs.Flush()

	return app.server.Close()
}

// Shutdown 关闭所有服务。
//
// 和 Close 的区别在于 Shutdown 会等待所有的服务完成之后才关闭，
// 等待时间由配置文件决定。
func (app *App) Shutdown() error {
	logs.Flush()

	if app.config.ShutdownTimeout <= 0 {
		return app.server.Close()
	}

	ctx, cancel := stdctx.WithTimeout(stdctx.Background(), app.config.ShutdownTimeout)
	defer cancel()
	return app.server.Shutdown(ctx)
}

// URL 构建一条基于 app.url 的完整 URL
func (app *App) URL(path string) string {
	if len(path) == 0 {
		return app.config.URL
	}

	if path[0] != '/' {
		path = "/" + path
	}
	return app.config.URL + path
}

// NewContext 根据当前配置，生成 context.Context 对象，若是出错则返回 nil
//
// encodingName 指定出输出时的编码方式，此编码名称必须已经通过 AddMarshal 添加；
// charsetName 指定输出时的字符集，此字符集名称必须已经通过 AddCharset 添加；
// strict 若为 true，则会验证用户的 Accept 报头是否接受 encodingName 编码。
// 输入时的编码与字符集信息从报头 Content-Type 中获取，若未指定字符集，则默认为 utf-8
func (app *App) NewContext(w http.ResponseWriter, r *http.Request) *context.Context {
	encName, charsetName := encoding.ParseContentType(r.Header.Get("Content-Type"))

	unmarshal := encoding.Unmarshal(encName)
	if unmarshal == nil {
		context.RenderStatus(w, http.StatusUnsupportedMediaType)
		return nil
	}

	inputCharset := encoding.Charset(charsetName)
	if inputCharset == nil {
		context.RenderStatus(w, http.StatusUnsupportedMediaType)
		return nil
	}

	if app.config.Strict {
		accept := r.Header.Get("Accept")
		if !strings.Contains(accept, app.config.OutputEncoding) && !strings.Contains(accept, "*/*") {
			context.RenderStatus(w, http.StatusNotAcceptable)
			return nil
		}
	}

	return &context.Context{
		Response:           w,
		Request:            r,
		OutputEncoding:     app.outputEncoding,
		OutputEncodingName: app.config.OutputEncoding,
		InputEncoding:      unmarshal,
		InputCharset:       inputCharset,
		OutputCharset:      app.outputCharset,
		OutputCharsetName:  app.config.OutputCharset,
	}
}

func logRecovery(w http.ResponseWriter, msg interface{}) {
	logs.Error(msg)
	context.RenderStatus(w, http.StatusInternalServerError)
}

func (app *App) buildHandler(h http.Handler) http.Handler {
	h = app.buildHosts(app.buildHeader(h))

	ff := logRecovery
	if app.config.Debug {
		ff = recovery.PrintDebug
	}
	h = recovery.New(h, ff)

	// NOTE: 在最外层添加调试地址，保证调试内容不会被其它 handler 干扰。
	if app.config.Debug {
		h = app.buildPprof(h)
	}

	return h
}

func (app *App) buildHosts(h http.Handler) http.Handler {
	if len(app.config.AllowedDomains) == 0 {
		return h
	}

	return host.New(h, app.config.AllowedDomains...)
}

func (app *App) buildHeader(h http.Handler) http.Handler {
	if len(app.config.Headers) == 0 {
		return h
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for k, v := range app.config.Headers {
			w.Header().Set(k, v)
		}
		h.ServeHTTP(w, r)
	})
}

// 根据 决定是否包装调试地址，调用前请确认是否已经开启 Pprof 选项
func (app *App) buildPprof(h http.Handler) http.Handler {
	logs.Debug("开启了调试功能，地址为：", pprofPath)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, pprofPath) {
			h.ServeHTTP(w, r)
			return
		}

		path := r.URL.Path[len(pprofPath):]
		switch path {
		case "cmdline":
			pprof.Cmdline(w, r)
		case "profile":
			pprof.Profile(w, r)
		case "symbol":
			pprof.Symbol(w, r)
		case "trace":
			pprof.Trace(w, r)
		default:
			pprof.Index(w, r)
		}
	}) // end return http.HandlerFunc
}
