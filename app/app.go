// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package app

import (
	ctx "context"
	"errors"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/issue9/logs"
	"github.com/issue9/middleware/compress"
	"github.com/issue9/mux"
	"github.com/issue9/utils"

	"github.com/issue9/web/context"
)

// BuildHandler 将一个 http.Handler 封装成另一个 http.Handler
//
// 一个中间件的接口定义，传递给 New() 函数，可以给全部的路由项添加一个中间件。
type BuildHandler func(http.Handler) http.Handler

// App 保存整个程序的运行环境，方便做整体的调度。
type App struct {
	configDir string
	config    *config
	url       string

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
// 框架默认的两个配置文件都会从此目录下查找。
// confDir 下面必须包含 logs.xml 与 web.yaml 两个配置文件。
func New(confDir string) (*App, error) {
	confDir, err := filepath.Abs(confDir)
	if err != nil {
		return nil, err
	}
	app := &App{configDir: confDir}

	if !utils.FileExists(confDir) {
		return nil, errors.New("配置文件的目录不存在")
	}

	if err = logs.InitFromXMLFile(app.File(logsFilename)); err != nil {
		return nil, err
	}

	conf, err := loadConfig(app.File(configFilename))
	if err != nil {
		return nil, err
	}

	app.initFromConfig(conf)

	return app, nil
}

func (app *App) initFromConfig(conf *config) {
	app.config = conf
	app.modules = make([]*Module, 0, 100)
	app.mux = mux.New(!conf.Options, false, nil, nil)
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

// Run 运行路由，执行监听程序。
//
// builder 用来给 mux 对象加上一个统一的中间件。不需要可以传递空值。
//
// 必须得保证在调用 Run() 时，logs 包的所有功能是可用的，
// 之后的好多操作，都会将日志输出 logs 中的相关通道中。
func (app *App) Run(build BuildHandler) error {
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
	if build != nil {
		h = build(h)
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
		for _, srv := range app.servers {
			ctx, cancel := ctx.WithTimeout(ctx.Background(), timeout)
			defer cancel()

			if err := srv.Shutdown(ctx); err != nil {
				return err
			}
		}
	}

	app.servers = app.servers[:0]
	return nil
}

// File 获取相对于配置目录下的文件。
func (app *App) File(path string) string {
	return filepath.Join(app.configDir, path)
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
func (app *App) NewContext(w http.ResponseWriter, r *http.Request) *context.Context {
	conf := app.config
	ctx, err := context.New(w, r, conf.OutputEncoding, conf.OutputCharset, conf.Strict)

	switch {
	case err == context.ErrUnsupportedContentType:
		context.RenderStatus(w, http.StatusUnsupportedMediaType)
		return nil
	case err == context.ErrClientNotAcceptable:
		context.RenderStatus(w, http.StatusNotAcceptable)
		return nil
	}

	return ctx
}
