// SPDX-License-Identifier: MIT

// Package web 一个微型的 RESTful API 框架
package web

import (
	ctx "context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/issue9/logs/v2"
	lc "github.com/issue9/logs/v2/config"
	"golang.org/x/text/message"

	"github.com/issue9/web/config"
	"github.com/issue9/web/context"
	"github.com/issue9/web/context/contentype"
	"github.com/issue9/web/context/contentype/gob"
	"github.com/issue9/web/internal/version"
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
	logs      *logs.Logs
	ctxServer *context.Server

	httpServer      *http.Server
	closed          chan struct{} // 当 shutdown 延时关闭时，通过此事件确定 Serve() 的返回时机。
	shutdownTimeout time.Duration

	// modules
	modules []*Module
	inited  bool
}

type contextKey int

// ContextKeyWeb 可以从 http.Request 中获取 Web 实例的关键字
const ContextKeyWeb contextKey = 0

// GetWeb 从 ctx 中获取 *Web 实例
func GetWeb(ctx *Context) *Web {
	return ctx.Server().Vars[ContextKeyWeb].(*Web)
}

// Classic 返回一个开箱即用的 Web 实例
func Classic(logConfigFile, configFile string) (*Web, error) {
	logConf := &lc.Config{}
	if err := config.LoadFile(logConfigFile, logConf); err != nil {
		return nil, err
	}
	if err := logConf.Sanitize(); err != nil {
		return nil, err
	}

	l := logs.New()
	if err := l.Init(logConf); err != nil {
		return nil, err
	}

	conf := &Config{}
	if err := config.LoadFile(configFile, conf); err != nil {
		return nil, err
	}

	conf.Marshalers = map[string]contentype.MarshalFunc{
		"application/json":         json.Marshal,
		"application/xml":          xml.Marshal,
		contentype.DefaultMimetype: gob.Marshal,
	}

	conf.Unmarshalers = map[string]contentype.UnmarshalFunc{
		"application/json":         json.Unmarshal,
		"application/xml":          xml.Unmarshal,
		contentype.DefaultMimetype: gob.Unmarshal,
	}

	conf.Results = map[int]message.Reference{
		40001: "无效的报头",
		40002: "无效的地址",
		40003: "无效的查询参数",
		40004: "无效的报文",
	}

	return New(l, conf)
}

// New 返回 Web 对象
func New(l *logs.Logs, conf *Config) (web *Web, err error) {
	if err = conf.sanitize(); err != nil {
		return nil, err
	}

	ctxServer, err := conf.toCTXServer(l)
	if err != nil {
		return nil, err
	}

	web = &Web{
		logs:      l,
		ctxServer: ctxServer,

		httpServer: &http.Server{
			Addr:              conf.addr,
			ReadTimeout:       conf.ReadTimeout.Duration(),
			ReadHeaderTimeout: conf.ReadHeaderTimeout.Duration(),
			WriteTimeout:      conf.WriteTimeout.Duration(),
			IdleTimeout:       conf.IdleTimeout.Duration(),
			MaxHeaderBytes:    conf.MaxHeaderBytes,
			ErrorLog:          l.ERROR(),
			TLSConfig:         conf.TLSConfig,
		},
		closed:          make(chan struct{}, 1),
		shutdownTimeout: conf.ShutdownTimeout.Duration(),

		modules: make([]*Module, 0, 10),
	}
	ctxServer.Vars[ContextKeyWeb] = web

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
	srv = context.NewServer(l, conf.Cache, conf.DisableOptions, conf.DisableHead, conf.url)

	if conf.ResultBuilder != nil {
		srv.ResultBuilder = conf.ResultBuilder
	}

	srv.Location = conf.location

	for path, dir := range conf.Static {
		if err := srv.Router().Static(path, dir); err != nil {
			return nil, err
		}
	}

	if conf.Catalog != nil {
		srv.Catalog = conf.Catalog
	}

	if err = srv.Mimetypes().AddMarshals(conf.Marshalers); err != nil {
		return nil, err
	}
	if err = srv.Mimetypes().AddUnmarshals(conf.Unmarshalers); err != nil {
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
	err = web.CTXServer().Serve(web.HTTPServer())

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
		web.CTXServer().Close()
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
