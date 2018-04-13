// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package app 核心功能的实现
package app

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/issue9/logs"
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

	middleware module.Middleware // 应用于全局路由项的中间件
	mux        *mux.Mux
	router     *mux.Prefix
	server     *http.Server

	modules []*module.Module

	// 根据配置文件，获取相应的输出编码和字符集。
	outputMimeType encoding.MarshalFunc
	outputCharset  xencoding.Encoding

	// 当 shutdown 延时关闭时，通过此事件确定 Run() 的返回时机。
	closed chan bool
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
		closed:     make(chan bool, 1),
	}

	if err = logs.InitFromXMLFile(app.File(logsFilename)); err != nil {
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

	app.modules = make([]*module.Module, 0, 50)

	return nil
}

// Modules 获取所有的模块信息
func (app *App) Modules() []*module.Module {
	return app.modules
}

// NewModule 声明一个新的模块
func (app *App) NewModule(name, desc string, deps ...string) (*module.Module, error) {
	for _, m := range app.modules {
		if m.Name == name {
			return nil, fmt.Errorf("模块 %s 已经存在", name)
		}
	}

	m := module.New(name, desc, deps...)
	app.modules = append(app.modules, m)

	return m, nil
}

// File 获取配置目录下的文件名
func (app *App) File(path ...string) string {
	paths := make([]string, 0, len(path)+1)
	paths = append(paths, app.configDir)
	return filepath.Join(append(paths, path...)...)
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

// NewContext 根据当前配置，生成 context.Context 对象，若是出错则 panic
func (app *App) NewContext(w http.ResponseWriter, r *http.Request) *context.Context {
	encName, charsetName, err := encoding.ParseContentType(r.Header.Get("Content-Type"))

	if err != nil {
		context.Exit(http.StatusUnsupportedMediaType)
	}

	unmarshal := encoding.Unmarshal(encName)
	if unmarshal == nil {
		context.Exit(http.StatusUnsupportedMediaType)
	}

	inputCharset := encoding.Charset(charsetName)
	if inputCharset == nil {
		context.Exit(http.StatusUnsupportedMediaType)
	}

	if app.config.Strict {
		accept := r.Header.Get("Accept")
		if accept != "" && !strings.Contains(accept, app.config.OutputMimeType) && !strings.Contains(accept, "*/*") {
			context.Exit(http.StatusNotAcceptable)
		}

		accept = r.Header.Get("Accept-Charset")
		if accept != "" && !strings.Contains(accept, app.config.OutputCharset) && !strings.Contains(accept, "*") {
			context.Exit(http.StatusNotAcceptable)
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
