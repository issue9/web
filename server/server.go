// SPDX-License-Identifier: MIT

// Package server 提供与服务端实现相关的功能
//
// 目前实现了三种类型的服务端：
//   - [New] 构建普通的 HTTP 服务；
//   - [NewGateway] 构建微服务的网关服务；
//   - [NewService] 构建微服务；
//
// 服务的初始化主要依赖 [Options] 对象，用户可以通过以下几种方式初始化 [Options] 对象：
//   - &Options{} 最简单的方式，所有的 [Options] 均支持默认值；
//   - [LoadOptions] 从配置文件中初始化 [Options] 对象；
//
// # 配置文件
//
// 对于配置文件各个字段的定义，可参考当前目录下的 CONFIG.html。
// 配置文件中除了固定的字段之外，还提供了泛型变量 User 用于指定用户自定义的额外字段。
//
// # 注册函数
//
// 当前包提供大量的注册函数，以用将某些无法直接采用序列化的内容转换可序列化的。
// 比如通过 [RegisterCompression] 将 `gzip-default` 等字符串表示成压缩算法，
// 以便在配置文件进行指定。
//
// 所有的注册函数处理逻辑上都相似，碰上同名的会覆盖，否则是添加。
// 且默认情况下都提供了一些可选项，只有在用户需要额外添加自己的内容时才需要调用注册函数。
package server

//go:generate go run ./make_doc.go

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
