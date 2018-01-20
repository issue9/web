// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package app

import (
	stdctx "context"
	"net/http"
	"strconv"
	"time"

	"github.com/issue9/logs"
	"github.com/issue9/middleware/compress"
	"github.com/issue9/mux"
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

	// 保存着所有的 http.Server 实例。
	//
	// 除了 mux 所依赖的 http.Server 实例之外，
	// 还有诸如 80 端口跳转等产生的 http.Server 实例。
	// 记录这些 server，关闭服务时需要将这些全部都关闭。
	servers []*http.Server
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
	app.servers = make([]*http.Server, 0, 5)

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
	return app.listen(app.buildHandler(h))
}

// Shutdown 关闭所有服务，之后 app 实例将不可再用，
//
// timeout 表示已有服务的等待时间。
// 若超过该时间，服务还未自动停止的，则会强制停止，若小于或等于 0 则立即重启。
func (app *App) Shutdown(timeout time.Duration) error {
	logs.Flush()

	if timeout <= 0 {
		for _, srv := range app.servers {
			if err := srv.Close(); err != nil {
				return err
			}
		}
	} else {
		// BUG(caixw) 多个服务之间，会依赖关闭，实际时间可能远远大于 timeout
		for _, srv := range app.servers {
			if err := closeServer(srv, timeout); err != nil {
				return err
			}
		} // end for
	}

	app.servers = app.servers[:0]
	return nil
}

func closeServer(srv *http.Server, timeout time.Duration) error {
	ctx, cancel := stdctx.WithTimeout(stdctx.Background(), timeout)
	defer cancel()

	return srv.Shutdown(ctx)
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
