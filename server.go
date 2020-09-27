// SPDX-License-Identifier: MIT

package web

import (
	ctx "context"
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"

	"github.com/issue9/logs/v2"
	"github.com/issue9/web/context"
	"github.com/issue9/web/module"
)

type contextKey int

// ContextKeyWeb 可以从 http.Request 中获取 Web 实例的关键字
const ContextKeyWeb contextKey = 0

// CTXServer 返回 context.Server 实例
func (web *Web) CTXServer() *context.Server {
	return web.ctxServer
}

// HTTPServer 返回 http.Server 实例
func (web *Web) HTTPServer() *http.Server {
	return web.httpServer
}

// MODServer 返回 *module.MODServer 实例
func (web *Web) MODServer() *module.Server {
	return web.modServer
}

// Serve 运行 HTTP 服务
func (web *Web) Serve() (err error) {
	web.modServer.RunServices()

	if web.isTLS {
		err = web.httpServer.ListenAndServeTLS("", "")
	} else {
		err = web.httpServer.ListenAndServe()
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

	if web.ShutdownTimeout == 0 {
		return web.httpServer.Close()
	}

	c, cancel := ctx.WithTimeout(ctx.Background(), web.ShutdownTimeout.Duration())
	defer cancel()
	if err := web.httpServer.Shutdown(c); err != nil && !errors.Is(err, ctx.DeadlineExceeded) {
		return err
	}
	return nil
}

// Init 根据内容进行初始化相关信息
//
// 当调用此函数之后，会根据 Web 的字段初始化所有的私有实例，之后再改变字段内容，不会再启作用。
// 但是可以通过 web.CTXServer() 返回的实例直接修改相关的参数。
func (web *Web) Init() (err error) {
	if err = web.sanitize(); err != nil {
		return err
	}

	web.logs = logs.New()
	if web.LogsConfig != nil {
		if err = web.logs.Init(web.LogsConfig); err != nil {
			return err
		}
	}

	web.ctxServer, err = context.NewServer(web.logs, web.ResultBuilder, web.DisableOptions, web.DisableHead, web.url.Path)
	if err != nil {
		return err
	}
	web.ctxServer.Debug = web.Debug
	web.ctxServer.Location = web.Location
	web.ctxServer.AllowedDomain(web.AllowedDomains...)
	for path, dir := range web.Static {
		web.ctxServer.AddStatic(path, dir)
	}
	for k, v := range web.Headers {
		web.ctxServer.SetHeader(k, v)
	}
	web.ctxServer.Interceptor = web.ContextInterceptor
	if err = web.ctxServer.AddMarshals(web.Marshalers); err != nil {
		return err
	}
	if err = web.ctxServer.AddUnmarshals(web.Unmarshalers); err != nil {
		return err
	}

	web.httpServer = &http.Server{
		Addr:              web.addr,
		Handler:           web.ctxServer.Handler(),
		ReadTimeout:       web.ReadTimeout.Duration(),
		ReadHeaderTimeout: web.ReadHeaderTimeout.Duration(),
		WriteTimeout:      web.WriteTimeout.Duration(),
		IdleTimeout:       web.IdleTimeout.Duration(),
		MaxHeaderBytes:    web.MaxHeaderBytes,
		ErrorLog:          web.logs.ERROR(),
		BaseContext: func(net.Listener) ctx.Context {
			return ctx.WithValue(ctx.Background(), ContextKeyWeb, web)
		},
	}

	if web.isTLS {
		web.isTLS = true
		web.httpServer.TLSConfig, err = web.toTLSConfig()
		if err != nil {
			return err
		}
	}

	web.closed = make(chan struct{}, 1)

	if web.modServer, err = module.NewServer(web.ctxServer, web.Plugins); err != nil {
		return err
	}

	if web.ShutdownSignal != nil {
		web.grace(web.ShutdownSignal...)
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
