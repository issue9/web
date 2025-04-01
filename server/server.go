// SPDX-FileCopyrightText: 2018-2025 caixw
//
// SPDX-License-Identifier: MIT

// Package server 提供与服务端实现相关的功能
//
// 目前实现了三种类型的服务端：
//   - [NewHTTP] 构建普通的 HTTP 服务；
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
		hs    *http.Server
		state web.State
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

func newHTTPServer(id, version string, o *Options, s web.Server) *httpServer {
	srv := &httpServer{
		hs:    o.HTTPServer,
		state: web.Stopped,
	}
	if s == nil {
		s = srv
	}

	srv.InternalServer = o.internalServer(id, version, s)
	srv.hs.Handler = srv

	for _, plugin := range o.Plugins { // NOTE: 需要保证在最后
		plugin.Plugin(srv)
	}
	return srv
}

// NewHTTP 新建 HTTP 服务
//
// id, version 表示服务的 ID 和版本号；
// o 指定了一些带有默认值的参数；
func NewHTTP(id, version string, o *Options) (web.Server, error) {
	o, err := sanitizeOptions(o, typeHTTP)
	if err != nil {
		return nil, err.AddFieldParent("o")
	}

	return newHTTPServer(id, version, o, nil), nil
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

	<-srv.Done()
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
	}()

	if err := srv.hs.Shutdown(ctx); err != nil && !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
		srv.Logs().ERROR().Error(err)
	}
}

// NewService 声明微服务节点
//
// [Options.Registry] 和 [Options.Peer] 不能为空。
func NewService(id, version string, o *Options) (web.Server, error) {
	o, err := sanitizeOptions(o, typeService)
	if err != nil {
		return nil, err.AddFieldParent("o")
	}

	s := &service{
		registry: o.Registry,
		peer:     o.Peer,
	}
	s.httpServer = newHTTPServer(id, version, o, s)
	return s, nil
}

func (s *service) Serve() error {
	dreg, err := s.registry.Register(s.ID(), s.peer)
	if err != nil {
		return err
	}
	s.OnClose(func() error { return dreg() })

	return s.httpServer.Serve()
}

// NewGateway 声明微服务的网关
//
// [Options.Mapper] 和 [Options.Peer] 不能为空。
func NewGateway(id, version string, o *Options) (web.Server, error) {
	o, err := sanitizeOptions(o, typeGateway)
	if err != nil {
		return nil, err.AddFieldParent("o")
	}

	g := &gateway{registry: o.Registry}
	g.httpServer = newHTTPServer(id, version, o, g)

	for name, match := range o.Mapper {
		proxy := g.registry.ReverseProxy(name, g)
		g.Routers().New(name, match).Any("{path}", func(ctx *web.Context) web.Responser {
			proxy.ServeHTTP(ctx, ctx.Request())
			return nil
		})
	}

	return g, nil
}
