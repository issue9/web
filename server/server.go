// SPDX-License-Identifier: MIT

// Package server 服务端实现
package server

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/issue9/web"
)

type httpServer struct {
	*web.InternalServer
	httpServer *http.Server
	state      web.State
	closed     chan struct{}
}

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
