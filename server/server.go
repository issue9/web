// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

//go:generate web htmldoc -lang=zh-CN -dir=./ -o=./CONFIG.html -object=configOf

// Package server 提供与服务端实现相关的功能
//
// 目前实现了三种类型的服务端：
//   - [New] 构建普通的 HTTP 服务；
//   - [NewGateway] 构建微服务的网关服务；
//   - [NewService] 构建微服务；
package server

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/issue9/web"
	"github.com/issue9/web/selector"
	"github.com/issue9/web/server/registry"
)

type (
	httpServer struct {
		*web.InternalServer
		hs     *http.Server
		state  web.State
		closed chan struct{}
	}

	service struct {
		*httpServer
		registry registry.Registry
		peer     selector.Peer
	}

	gateway struct {
		*httpServer
		registry registry.Registry
	}
)

func newHTTPServer(name, version string, o *Options, s web.Server) *httpServer {
	srv := &httpServer{
		hs:     o.HTTPServer,
		state:  web.Stopped,
		closed: make(chan struct{}, 1),
	}
	if s == nil {
		s = srv
	}

	srv.InternalServer = o.internalServer(name, version, s)
	srv.hs.Handler = srv

	for _, f := range o.Init { // NOTE: 需要保证在最后
		f(srv)
	}
	return srv
}

// New 新建 HTTP 服务
//
// name, version 表示服务的名称和版本号；
// o 指定了一些带有默认值的参数；
func New(name, version string, o *Options) (web.Server, error) {
	o, err := sanitizeOptions(o, typeHTTP)
	if err != nil {
		return nil, err.AddFieldParent("o")
	}

	return newHTTPServer(name, version, o, nil), nil
}

func (srv *httpServer) State() web.State { return srv.state }

func (srv *httpServer) Serve() (err error) {
	if srv.State() == web.Running {
		panic("当前已经处于运行状态")
	}
	srv.state = web.Running

	if c := srv.hs.TLSConfig; c != nil && (len(c.Certificates) > 0 || c.GetCertificate != nil) {
		err = srv.hs.ListenAndServeTLS("", "")
	} else {
		err = srv.hs.ListenAndServe()
	}

	if errors.Is(err, http.ErrServerClosed) {
		// 由 [Server.Close] 主动触发的关闭事件，才需要等待其执行完成。
		// 其它错误直接返回，否则一些内部错误会永远卡在此处无法返回。
		<-srv.closed
	}
	return err
}

func (srv *httpServer) Close(shutdownTimeout time.Duration) {
	if srv.State() != web.Running {
		return
	}
	srv.state = web.Stopped // 调用 Close 即设置状态

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)

	defer func() {
		srv.InternalServer.Close()
		cancel()

		// [http.Server.Shutdown] 会让 [http.Server.ListenAndServe] 等方法直接返回，
		// 所以由 srv.close 保证在当前函数返回之后再通知 [Server.Serve] 退出。
		srv.closed <- struct{}{}
	}()

	if err := srv.hs.Shutdown(ctx); err != nil && !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
		srv.Logs().ERROR().Error(err)
	}
}

// NewService 声明微服务节点
func NewService(name, version string, o *Options) (web.Server, error) {
	o, err := sanitizeOptions(o, typeService)
	if err != nil {
		return nil, err.AddFieldParent("o")
	}

	s := &service{
		registry: o.Registry,
		peer:     o.Peer,
	}
	s.httpServer = newHTTPServer(name, version, o, s)
	return s, nil
}

func (s *service) Serve() error {
	dreg, err := s.registry.Register(s.Name(), s.peer)
	if err != nil {
		return err
	}
	s.OnClose(func() error { return dreg() })

	return s.httpServer.Serve()
}

// NewGateway 声明微服务的网关
func NewGateway(name, version string, o *Options) (web.Server, error) {
	o, err := sanitizeOptions(o, typeGateway)
	if err != nil {
		return nil, err.AddFieldParent("o")
	}

	g := &gateway{registry: o.Registry}
	g.httpServer = newHTTPServer(name, version, o, g)

	for name, match := range o.Mapper {
		proxy := g.registry.ReverseProxy(name, g)
		g.Routers().New(name, match).Any("{path}", func(ctx *web.Context) web.Responser {
			proxy.ServeHTTP(ctx, ctx.Request())
			return nil
		})
	}

	return g, nil
}
