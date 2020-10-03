// SPDX-License-Identifier: MIT

package web

import (
	ctx "context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/issue9/logs/v2"
	"golang.org/x/text/message"

	"github.com/issue9/web/context"
	"github.com/issue9/web/context/mimetype"
	"github.com/issue9/web/context/mimetype/gob"
	"github.com/issue9/web/module"
)

type contextKey int

// ContextKeyWeb 可以从 http.Request 中获取 Web 实例的关键字
const ContextKeyWeb contextKey = 0

// Web 管理整个项目所有实例
type Web struct {
	isTLS           bool
	logs            *logs.Logs
	httpServer      *http.Server
	ctxServer       *context.Server
	modServer       *module.Server
	closed          chan struct{} // 当 shutdown 延时关闭时，通过此事件确定 Serve() 的返回时机。
	shutdownTimeout time.Duration
}

// GetWeb 从 ctx 中获取 *Web 实例
func GetWeb(ctx *Context) *Web {
	return ctx.Request.Context().Value(ContextKeyWeb).(*Web)
}

// Classic 返回一个开箱即用的 Web 实例
//
// 会加载 dir 目录下的 web.yaml 和 logs.xml 两个配置文件内容，并用于初始化 Web 实例。
func Classic(dir string) (*Web, error) {
	conf, err := LoadConfig(dir)
	if err != nil {
		return nil, err
	}

	conf.Marshalers = map[string]mimetype.MarshalFunc{
		"application/json":       json.Marshal,
		"application/xml":        xml.Marshal,
		mimetype.DefaultMimetype: gob.Marshal,
	}

	conf.Unmarshalers = map[string]mimetype.UnmarshalFunc{
		"application/json":       json.Unmarshal,
		"application/xml":        xml.Unmarshal,
		mimetype.DefaultMimetype: gob.Unmarshal,
	}

	conf.Results = map[int]string{
		40001: "无效的报头",
		40002: "无效的地址",
		40003: "无效的查询参数",
		40004: "无效的报文",
	}

	if conf.ResultBuilder == nil {
		conf.ResultBuilder = context.DefaultResultBuilder
	}

	if conf.Catalog == nil {
		conf.Catalog = message.DefaultCatalog
	}

	return New(conf)
}

// New 根据内容进行初始化 Web 对象
func New(conf *Config) (web *Web, err error) {
	if err = conf.sanitize(); err != nil {
		return nil, err
	}

	web = &Web{}

	web.logs = logs.New()
	if conf.Logs != nil {
		if err = web.logs.Init(conf.Logs); err != nil {
			return nil, err
		}
	}

	web.ctxServer = context.NewServer(web.logs, conf.DisableOptions, conf.DisableHead, conf.url)
	if conf.ResultBuilder != nil {
		web.ctxServer.ResultBuilder = conf.ResultBuilder
	}
	web.ctxServer.Interceptor = conf.ContextInterceptor
	web.ctxServer.Location = conf.location
	for path, dir := range conf.Static {
		web.ctxServer.AddStatic(path, dir)
	}
	web.ctxServer.Catalog = conf.Catalog
	if err = web.ctxServer.AddMarshals(conf.Marshalers); err != nil {
		return nil, err
	}
	if err = web.ctxServer.AddUnmarshals(conf.Unmarshalers); err != nil {
		return nil, err
	}
	for status, rslt := range conf.results {
		web.ctxServer.AddMessages(status, rslt)
	}
	if conf.Debug != nil {
		web.ctxServer.SetDebugger(conf.Debug.Pprof, conf.Debug.Vars)
	}
	web.ctxServer.AddMiddlewares(conf.Middlewares...)

	web.httpServer = &http.Server{
		Addr:              conf.addr,
		Handler:           web.ctxServer.Handler(),
		ReadTimeout:       conf.ReadTimeout.Duration(),
		ReadHeaderTimeout: conf.ReadHeaderTimeout.Duration(),
		WriteTimeout:      conf.WriteTimeout.Duration(),
		IdleTimeout:       conf.IdleTimeout.Duration(),
		MaxHeaderBytes:    conf.MaxHeaderBytes,
		ErrorLog:          web.logs.ERROR(),
		BaseContext: func(net.Listener) ctx.Context {
			return ctx.WithValue(ctx.Background(), ContextKeyWeb, web)
		},
	}

	if conf.isTLS {
		web.isTLS = true
		web.httpServer.TLSConfig = conf.TLSConfig
	}

	web.closed = make(chan struct{}, 1)
	web.shutdownTimeout = conf.ShutdownTimeout.Duration()

	if web.modServer, err = module.NewServer(web.ctxServer, conf.Plugins); err != nil {
		return nil, err
	}

	if conf.ShutdownSignal != nil {
		web.grace(conf.ShutdownSignal...)
	}

	return web, nil
}

// CTXServer 返回 context.Server 实例
func (web *Web) CTXServer() *context.Server {
	return web.ctxServer
}

// HTTPServer 返回 http.Server 实例
func (web *Web) HTTPServer() *http.Server {
	return web.httpServer
}

// Modules 返回模块列表
func (web *Web) Modules() []*Module {
	return web.modServer.Modules()
}

// Services 返回服务列表
func (web *Web) Services() []*Service {
	return web.modServer.Services()
}

// Tags 返回所有标签列表
//
// 键名为模块名称，键值为该模块下的标签列表。
func (web *Web) Tags() map[string][]string {
	return web.modServer.Tags()
}

// InitModules 初始化模块
func (web *Web) InitModules(tag string) error {
	return web.modServer.Init(tag, web.logs.INFO())
}

// Serve 运行 HTTP 服务
func (web *Web) Serve() (err error) {
	web.modServer.RunServices()

	if web.isTLS {
		err = web.HTTPServer().ListenAndServeTLS("", "")
	} else {
		err = web.HTTPServer().ListenAndServe()
	}

	// 由 Shutdown() 或 Close() 主动触发的关闭事件，才需要等待其执行完成，
	// 其它错误直接返回，否则一些内部错误会永远卡在此处无法返回。
	if errors.Is(err, http.ErrServerClosed) {
		<-web.closed
	}
	return err
}

// Close 关闭服务
func (web *Web) Close() error {
	defer func() {
		web.modServer.StopServices()
		web.closed <- struct{}{}
	}()

	if web.shutdownTimeout == 0 {
		return web.HTTPServer().Close()
	}

	c, cancel := ctx.WithTimeout(ctx.Background(), web.shutdownTimeout)
	defer cancel()
	if err := web.HTTPServer().Shutdown(c); err != nil && !errors.Is(err, ctx.DeadlineExceeded) {
		return err
	}
	return nil
}

func (web *Web) grace(sig ...os.Signal) {
	go func() {
		signalChannel := make(chan os.Signal)
		signal.Notify(signalChannel, sig...)

		<-signalChannel
		signal.Stop(signalChannel)
		close(signalChannel)

		if err := web.Close(); err != nil {
			web.logs.Error(err)
		}
		web.logs.Flush() // 保证内容会被正常输出到日志。
	}()
}
