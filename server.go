// SPDX-License-Identifier: MIT

package web

import (
	ctx "context"
	"errors"
	"net/http"

	"github.com/issue9/web/context"
	"github.com/issue9/web/module"
)

// CTXServer 返回 context.Server 实例
func (web *Web) CTXServer() *context.Server {
	return web.ctxServer
}

// HTTPServer 返回 http.Server 实例
func (web *Web) HTTPServer() *http.Server {
	return web.httpServer
}

// Modules 返回 *module.Modules 实例
func (web *Web) Modules() *module.Modules {
	return web.modules
}

// Serve 运行 HTTP 服务
func (web *Web) Serve() (err error) {
	web.modules.Run()

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
//
// 无论配置文件如果设置，此函数都是直接关闭服务，不会等待。
func (web *Web) Close() error {
	defer func() {
		web.modules.Stop()
		web.closed <- struct{}{}
	}()

	return web.httpServer.Close()
}

// Shutdown 等待完成所有请求并关闭服务
//
// 根据配置文件中的配置项，决定当前是直接关闭还是延时之后关闭。
func (web *Web) Shutdown(c ctx.Context) error {
	defer func() {
		web.modules.Stop()
		web.closed <- struct{}{}
	}()

	if err := web.httpServer.Shutdown(c); err != nil && !errors.Is(err, ctx.DeadlineExceeded) {
		return err
	}
	return nil
}

// Init 根据内容进行初始化相关信息
func (web *Web) Init() (err error) {
	if err = web.sanitize(); err != nil {
		return err
	}

	web.ctxServer, err = context.NewServer(web.Logs, web.ResultBuilder, web.DisableOptions, web.DisableHead, web.url.Path)
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

	web.httpServer = &http.Server{
		Addr:              web.addr,
		Handler:           web.ctxServer.Handler(),
		ReadTimeout:       web.ReadTimeout.Duration(),
		ReadHeaderTimeout: web.ReadHeaderTimeout.Duration(),
		WriteTimeout:      web.WriteTimeout.Duration(),
		IdleTimeout:       web.IdleTimeout.Duration(),
		MaxHeaderBytes:    web.MaxHeaderBytes,
		ErrorLog:          web.Logs.ERROR(),
	}

	if web.isTLS {
		web.isTLS = true
		web.httpServer.TLSConfig, err = web.toTLSConfig()
		if err != nil {
			return err
		}
	}

	web.closed = make(chan struct{}, 1)

	web.modules, err = module.NewModules(web.ctxServer, web.config, web.Plugins)
	return err
}
