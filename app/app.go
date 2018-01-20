// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package app

import (
	stdctx "context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/issue9/logs"
	"github.com/issue9/middleware/compress"
	"github.com/issue9/mux"
	"github.com/issue9/web/context"
	"github.com/issue9/web/encoding"
)

const (
	logsFilename   = "logs.xml" // 日志配置文件的文件名。
	configFilename = "web.yaml" // 配置文件的文件名。
)

// BuildHandler 将一个 http.Handler 封装成另一个 http.Handler
//
// 一个中间件的接口定义，传递给 New() 函数，可以给全部的路由项添加一个中间件。
type BuildHandler func(http.Handler) http.Handler

// App 保存整个程序的运行环境，方便做整体的调度。
type App struct {
	config *config

	// 根据 config 中的相关变量生成网站的地址
	//
	// 包括协议、域名、端口和根目录等。
	url string

	modules []*Module

	mux    *mux.Mux
	router *mux.Prefix

	server *http.Server
}

// New 初始化框架的基本内容。
//
// confDir 指定了配置文件所在的目录。
// 框架默认的两个配置文件 logs.xml 和 web.yaml 都会从此目录下查找。
//
// 用户的自定义配置文件也可以存在此目录下，就可以通过
// App.File() 获取文件内容。
//
// builder 用来给 mux 对象加上一个统一的中间件。不需要可以传递空值。
func New(conf *config) (*App, error) {
	app := &App{}

	app.initFromConfig(conf)

	return app, nil
}

// IsDebug 是否处在调试模式
func (app *App) IsDebug() bool {
	return app.config.Debug
}

func (app *App) initFromConfig(conf *config) {
	app.config = conf
	app.modules = make([]*Module, 0, 100)
	app.mux = mux.New(conf.DisableOptions, false, nil, nil)
	app.router = app.mux.Prefix(conf.Root)

	if conf.HTTPS {
		app.url = "https://" + conf.Domain
		if conf.Port != httpsPort {
			app.url += ":" + strconv.Itoa(conf.Port)
		}
	} else {
		app.url = "http://" + conf.Domain
		if conf.Port != httpPort {
			app.url += ":" + strconv.Itoa(conf.Port)
		}
	}

	app.url += conf.Root
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

	var h http.Handler = app.mux
	if app.config.Build != nil {
		h = app.config.Build(h)
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

// Shutdown 关闭所有服务，之后 app 实例将不可再用，
//
// timeout 表示已有服务的等待时间。
// 若超过该时间，服务还未自动停止的，则会强制停止，若小于或等于 0 则立即重启。
func (app *App) Shutdown(timeout time.Duration) error {
	logs.Flush()

	if timeout <= 0 {
		return app.server.Close()
	}

	ctx, cancel := stdctx.WithTimeout(stdctx.Background(), timeout)
	defer cancel()
	return app.server.Shutdown(ctx)
}

// URL 构建一条基于 app.url 的完整 URL
func (app *App) URL(path string) string {
	if len(path) == 0 {
		return app.url
	}

	if path[0] != '/' {
		path = "/" + path
	}
	return app.url + path
}

// NewContext 根据当前配置，生成 context.Context 对象，若是出错则返回 nil
//
// encodingName 指定出输出时的编码方式，此编码名称必须已经通过 AddMarshal 添加；
// charsetName 指定输出时的字符集，此字符集名称必须已经通过 AddCharset 添加；
// strict 若为 true，则会验证用户的 Accept 报头是否接受 encodingName 编码。
// 输入时的编码与字符集信息从报头 Content-Type 中获取，若未指定字符集，则默认为 utf-8
func (app *App) NewContext(w http.ResponseWriter, r *http.Request) *context.Context {
	encName, charsetName := encoding.ParseContentType(r.Header.Get("Content-Type"))

	unmarshal, found := app.config.Unmarshals[encName]
	if !found {
		context.RenderStatus(w, http.StatusUnsupportedMediaType)
		return nil
	}

	inputCharset, found := app.config.Charset[charsetName]
	if !found {
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
		OutputEncoding:     app.config.outputEncoding,
		OutputEncodingName: app.config.OutputEncoding,
		InputEncoding:      unmarshal,
		InputCharset:       inputCharset,
		OutputCharset:      app.config.outputCharset,
		OutputCharsetName:  app.config.OutputCharset,
	}
}
