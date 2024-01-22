// SPDX-License-Identifier: MIT

// Package server 服务端实现
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
		httpServer *http.Server
		state      web.State
		closed     chan struct{}
	}

	service struct {
		*httpServer
		registry registry.Registry
		peer     selector.Peer
	}

	gateway struct {
		*httpServer
		registry registry.Registry
		mapper   registry.Mapper
	}
)

func newHTTPServer(name, version string, o *Options, s web.Server) *httpServer {
	srv := &httpServer{
		httpServer: o.HTTPServer,
		state:      web.Stopped,
		closed:     make(chan struct{}, 1),
	}
	if s == nil {
		s = srv
	}

	srv.InternalServer = o.internalServer(name, version, s)
	srv.httpServer.Handler = srv

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
		err.Path = "Options"
		return nil, err
	}

	return newHTTPServer(name, version, o, nil), nil
}

func (srv *httpServer) State() web.State { return srv.state }

func (srv *httpServer) Serve() (err error) {
	if srv.State() == web.Running {
		panic("当前已经处于运行状态")
	}
	srv.state = web.Running

	if c := srv.httpServer.TLSConfig; c != nil && (len(c.Certificates) > 0 || c.GetCertificate != nil) {
		err = srv.httpServer.ListenAndServeTLS("", "")
	} else {
		err = srv.httpServer.ListenAndServe()
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

	defer func() {
		srv.InternalServer.Close()
		srv.state = web.Stopped
		srv.closed <- struct{}{} // NOTE: 保证最后执行
	}()

	if shutdownTimeout == 0 {
		if err := srv.httpServer.Close(); err != nil {
			srv.Logs().ERROR().Error(err)
		}
		return
	}

	c, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.httpServer.Shutdown(c); err != nil && !errors.Is(err, context.DeadlineExceeded) {
		srv.Logs().ERROR().Error(err)
	}
}

// NewService 将 [web.Server] 作为微服务节点
func NewService(name, version string, o *Options) (web.Server, error) {
	o, err := sanitizeOptions(o, typeService)
	if err != nil {
		err.Path = "Options"
		return nil, err
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
		err.Path = "Options"
		return nil, err
	}

	g := &gateway{
		registry: o.Registry,
		mapper:   o.Mapper,
	}
	g.httpServer = newHTTPServer(name, version, o, g)
	return g, nil
}

func (g *gateway) Serve() error {
	proxy := g.registry.ReverseProxy(g.mapper, g)

	r := g.Routers().New("proxy", nil)
	r.Any("{path}", func(ctx *web.Context) web.Responser {
		proxy.ServeHTTP(ctx, ctx.Request())
		return nil
	})

	return g.httpServer.Serve()
}
