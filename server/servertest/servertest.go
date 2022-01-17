// SPDX-License-Identifier: MIT

// Package servertest 针对 server 的测试用例
package servertest

import (
	"net/http"
	"time"

	"github.com/issue9/assert/v2"
	"github.com/issue9/assert/v2/rest"
	"github.com/issue9/mux/v6/group"

	"github.com/issue9/web/server"
)

type Server struct {
	a    *assert.Assertion
	s    *server.Server
	exit chan struct{}
}

// NewServer 声明一个 server 实例
func NewServer(a *assert.Assertion, s *server.Server) *Server {
	a.TB().Helper()
	a.NotNil(s)

	return &Server{
		s:    s,
		a:    a,
		exit: make(chan struct{}, 1),
	}
}

func (s *Server) Server() *server.Server { return s.s }

func (s *Server) GoServe() {
	go func() {
		s.a.TB().Helper()

		err := s.s.Serve()
		s.a.Error(err).ErrorIs(err, http.ErrServerClosed, "错误信息为:%v", err)
		s.exit <- struct{}{}
	}()

	// 等待 srv.Serve() 启动完毕，不同机器可能需要的时间会不同
	time.Sleep(5000 * time.Microsecond)
}

// NewRouter 创建一个默认的路由
//
// 相当于：
//  s.Server().NewRouter("default", "http://localhost:8080/", group.MatcherFunc(group.Any))
func (s *Server) NewRouter() *server.Router {
	s.a.TB().Helper()

	router := s.Server().NewRouter("default", "http://localhost:8080/", group.MatcherFunc(group.Any))
	s.a.NotNil(router)
	return router
}

// Wait 等待 GoServe 退出
func (s *Server) Wait() { <-s.exit }

func (s *Server) NewRequest(method, path string) *rest.Request {
	return rest.NewRequest(s.a, method, path).Client(http.DefaultClient)
}

func (s *Server) Get(path string) *rest.Request {
	return s.NewRequest(http.MethodGet, path)
}
