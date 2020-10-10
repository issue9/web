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
	"github.com/issue9/scheduled"

	"github.com/issue9/web/context"
	"github.com/issue9/web/context/mimetype"
	"github.com/issue9/web/context/mimetype/gob"
	"github.com/issue9/web/internal/version"
	"github.com/issue9/web/service"
)

// Version 当前框架的版本
const Version = version.Version

// Context 定义了在单个 HTTP 请求期间的上下文环境
//
// 是对 http.ResponseWriter 和 http.Request 的简单包装。
type Context = context.Context

// Result 定义了返回给用户的错误信息
type Result = context.Result

// Web 管理整个项目所有实例
type Web struct {
	isTLS     bool
	logs      *logs.Logs
	ctxServer *context.Server

	httpServer      *http.Server
	closed          chan struct{} // 当 shutdown 延时关闭时，通过此事件确定 Serve() 的返回时机。
	shutdownTimeout time.Duration

	// modules
	services  *service.Manager
	scheduled *scheduled.Server
	modules   []*Module
	inited    bool
}

type contextKey int

// ContextKeyWeb 可以从 http.Request 中获取 Web 实例的关键字
const ContextKeyWeb contextKey = 0

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

	return New(conf)
}

// New 根据内容进行初始化 Web 对象
func New(conf *Config) (web *Web, err error) {
	if err = conf.sanitize(); err != nil {
		return nil, err
	}

	l := logs.New()
	if conf.Logs != nil {
		if err = l.Init(conf.Logs); err != nil {
			return nil, err
		}
	}

	ctxServer, err := conf.toCTXServer(l)
	if err != nil {
		return nil, err
	}

	web = &Web{
		isTLS:     conf.isTLS,
		logs:      l,
		ctxServer: ctxServer,

		httpServer: &http.Server{
			Addr:              conf.addr,
			Handler:           ctxServer.Handler(),
			ReadTimeout:       conf.ReadTimeout.Duration(),
			ReadHeaderTimeout: conf.ReadHeaderTimeout.Duration(),
			WriteTimeout:      conf.WriteTimeout.Duration(),
			IdleTimeout:       conf.IdleTimeout.Duration(),
			MaxHeaderBytes:    conf.MaxHeaderBytes,
			ErrorLog:          l.ERROR(),
			TLSConfig:         conf.TLSConfig,
			BaseContext: func(net.Listener) ctx.Context {
				return ctx.WithValue(ctx.Background(), ContextKeyWeb, web)
			},
		},
		closed:          make(chan struct{}, 1),
		shutdownTimeout: conf.ShutdownTimeout.Duration(),

		services:  service.NewManager(),
		scheduled: scheduled.NewServer(conf.location),
		modules:   make([]*Module, 0, 10),
	}

	web.services.AddService(web.scheduledService, "计划任务")

	if conf.ShutdownSignal != nil {
		web.grace(conf.ShutdownSignal...)
	}

	if conf.Plugins != "" {
		if err := web.loadPlugins(conf.Plugins); err != nil {
			return nil, err
		}
	}

	return web, nil
}

func (conf *Config) toCTXServer(l *logs.Logs) (srv *context.Server, err error) {
	srv = context.NewServer(l, conf.DisableOptions, conf.DisableHead, conf.url)

	if conf.ResultBuilder != nil {
		srv.ResultBuilder = conf.ResultBuilder
	}

	srv.Location = conf.location

	for path, dir := range conf.Static {
		srv.AddStatic(path, dir)
	}

	if conf.Catalog != nil {
		srv.Catalog = conf.Catalog
	}

	if err = srv.AddMarshals(conf.Marshalers); err != nil {
		return nil, err
	}
	if err = srv.AddUnmarshals(conf.Unmarshalers); err != nil {
		return nil, err
	}

	for status, rslt := range conf.results {
		srv.AddMessages(status, rslt)
	}

	if conf.Debug != nil {
		srv.SetDebugger(conf.Debug.Pprof, conf.Debug.Vars)
	}

	if len(conf.Middlewares) > 0 {
		srv.AddMiddlewares(conf.Middlewares...)
	}
	if len(conf.Filters) > 0 {
		srv.AddFilters(conf.Filters...)
	}

	for _, h := range conf.ErrorHandlers {
		srv.SetErrorHandle(h.Handler, h.Status...)
	}

	return srv, nil
}

// Logs 返回日志实例
func (web *Web) Logs() *logs.Logs {
	return web.logs
}

// CTXServer 返回 context.Server 实例
func (web *Web) CTXServer() *context.Server {
	return web.ctxServer
}

// HTTPServer 返回 http.Server 实例
func (web *Web) HTTPServer() *http.Server {
	return web.httpServer
}

// Serve 运行 HTTP 服务
func (web *Web) Serve() (err error) {
	web.services.Run()

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
		web.services.Stop()
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
