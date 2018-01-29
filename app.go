// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"errors"
	"net/http"
	"path"
	"path/filepath"
	"strings"

	"github.com/issue9/logs"
	"github.com/issue9/middleware/compress"
	"github.com/issue9/mux"
	charset "golang.org/x/text/encoding"

	"github.com/issue9/web/context"
	"github.com/issue9/web/encoding"
	"github.com/issue9/web/internal/config"
	"github.com/issue9/web/internal/server"
)

const (
	logsFilename   = "logs.xml" // 日志配置文件的文件名。
	configFilename = "web.yaml" // 配置文件的文件名。
)

type app struct {
	configDir string
	config    *config.Config
	modules   []*Module

	middleware Middleware // 应用于全局所有路由项的中间件
	mux        *mux.Mux
	router     *mux.Prefix

	// 根据配置文件，获取相应的输出编码和字符集。
	outputEncoding encoding.MarshalFunc
	outputCharset  charset.Encoding
}

func newApp(configDir string, m Middleware) (*app, error) {
	dir, err := filepath.Abs(configDir)
	if err != nil {
		return nil, err
	}

	app := &app{
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

func (app *app) loadConfig() error {
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
func IsDebug() bool {
	return defaultApp.config.Debug
}

// File 获取配置目录下的文件名
func (app *app) File(path ...string) string {
	paths := make([]string, 0, len(path)+1)
	paths = append(paths, app.configDir)
	return filepath.Join(append(paths, path...)...)
}

// Run 加载各个模块的数据，运行路由，执行监听程序。
//
// 必须得保证在调用 Run() 时，logs 包的所有功能是可用的，
// 之后的好多操作，都会将日志输出 logs 中的相关通道中。
func (app *app) Run() error {
	if err := app.initDependency(); err != nil {
		return err
	}

	// 静态文件路由，在其它路由构建之前调用
	for url, dir := range app.config.Static {
		pattern := path.Join(app.config.Root, url+"{path}")
		fs := http.FileServer(http.Dir(dir))
		app.router.Get(pattern, http.StripPrefix(url, compress.New(fs, logs.ERROR())))
	}

	if app.middleware != nil {
		return server.Listen(app.middleware(app.mux), app.config)
	}
	return server.Listen(app.mux, app.config)
}

// Close 立即关闭服务
func (app *app) Close() error {
	return server.Close()
}

// Shutdown 关闭所有服务。
//
// 和 Close 的区别在于 Shutdown 会等待所有的服务完成之后才关闭，
// 等待时间由配置文件决定。
func (app *app) Shutdown() error {
	return server.Shutdown()
}

// URL 构建一条基于 app.url 的完整 URL
func (app *app) URL(path string) string {
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
func (app *app) NewContext(w http.ResponseWriter, r *http.Request) *context.Context {
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
