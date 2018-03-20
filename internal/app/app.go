// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package app

import (
	"errors"
	"net/http"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/issue9/logs"
	"github.com/issue9/middleware/compress"
	"github.com/issue9/mux"
	xencoding "golang.org/x/text/encoding"

	"github.com/issue9/web/context"
	"github.com/issue9/web/encoding"
	"github.com/issue9/web/internal/config"
	"github.com/issue9/web/module"
)

const (
	logsFilename   = "logs.xml" // 日志配置文件的文件名。
	configFilename = "web.yaml" // 配置文件的文件名。
)

// App 程序运行实例
type App struct {
	configDir string
	config    *config.Config
	modules   *module.Modules

	middleware module.Middleware // 应用于全局所有路由项的中间件
	mux        *mux.Mux
	router     *mux.Prefix
	server     *http.Server

	// 根据配置文件，获取相应的输出编码和字符集。
	outputMimeType encoding.MarshalFunc
	outputCharset  xencoding.Encoding
}

// New 声明一个新的 App 实例
func New(configDir string, m module.Middleware) (*App, error) {
	dir, err := filepath.Abs(configDir)
	if err != nil {
		return nil, err
	}

	app := &App{
		configDir:  dir,
		middleware: m,
	}

	if err := logs.InitFromXMLFile(app.File(logsFilename)); err != nil {
		return nil, err
	}

	if err = app.loadConfig(); err != nil {
		return nil, err
	}

	return app, nil
}

// Debug 是否处于调试模式
func (app *App) Debug() bool {
	return app.config.Debug
}

func (app *App) loadConfig() error {
	conf, err := config.Load(app.File(configFilename))
	if err != nil {
		return err
	}

	app.config = conf
	app.mux = mux.New(conf.DisableOptions, false, nil, nil)
	app.router = app.mux.Prefix(conf.Root)

	app.outputMimeType = encoding.Marshal(conf.OutputMimeType)
	if app.outputMimeType == nil {
		return errors.New("未找到 outputMimeType")
	}

	app.outputCharset = encoding.Charset(conf.OutputCharset)
	if app.outputCharset == nil {
		return errors.New("未找到 outputCharset")
	}

	app.modules = module.NewModules(app.router)

	return nil
}

// NewModule 声明一个新的模块
func (app *App) NewModule(name, desc string, deps ...string) *module.Module {
	return app.modules.New(name, desc, deps...)
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
	if err := app.modules.Init(); err != nil {
		return err
	}

	// 静态文件路由，在其它路由构建之前调用
	for url, dir := range app.config.Static {
		pattern := path.Join(app.config.Root, url+"{path}")
		fs := http.FileServer(http.Dir(dir))
		app.router.Get(pattern, http.StripPrefix(url, compress.New(fs, logs.ERROR())))
	}

	var h http.Handler = app.mux
	if app.middleware != nil {
		h = app.middleware(app.mux)
	}
	app.server = &http.Server{
		Addr:         ":" + strconv.Itoa(app.config.Port),
		Handler:      buildHandler(app.config, h), // 依赖全局变量 conf
		ErrorLog:     logs.ERROR(),
		ReadTimeout:  app.config.ReadTimeout,
		WriteTimeout: app.config.WriteTimeout,
	}

	if !app.config.HTTPS {
		return app.server.ListenAndServe()
	}

	return app.server.ListenAndServeTLS(app.config.CertFile, app.config.KeyFile)
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
func (app *App) NewContext(w http.ResponseWriter, r *http.Request) *context.Context {
	encName, charsetName, err := encoding.ParseContentType(r.Header.Get("Content-Type"))

	if err != nil {
		context.RenderStatus(w, http.StatusUnsupportedMediaType)
		return nil
	}

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
		if accept != "" && !strings.Contains(accept, app.config.OutputMimeType) && !strings.Contains(accept, "*/*") {
			context.RenderStatus(w, http.StatusNotAcceptable)
			return nil
		}

		accept = r.Header.Get("Accept-Charset")
		if accept != "" && !strings.Contains(accept, app.config.OutputCharset) && !strings.Contains(accept, "*") {
			context.RenderStatus(w, http.StatusNotAcceptable)
			return nil
		}
	}

	return &context.Context{
		Response:           w,
		Request:            r,
		OutputMimeType:     app.outputMimeType,
		OutputMimeTypeName: app.config.OutputMimeType,
		InputMimeType:      unmarshal,
		InputCharset:       inputCharset,
		OutputCharset:      app.outputCharset,
		OutputCharsetName:  app.config.OutputCharset,
	}
}
