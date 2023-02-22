// SPDX-License-Identifier: MIT

// Package servertest 针对 server 的测试用例
package servertest

import (
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/issue9/assert/v3"
	"github.com/issue9/assert/v3/rest"
	"github.com/issue9/mux/v7"

	"github.com/issue9/web/server"
)

type Server struct {
	a        *assert.Assertion
	s        *server.Server
	r        *server.Router
	hostname string

	// 采用 sync.WaitGroup 而不是 chan，
	// 可以保证用户在不调用 GoServe 的情况下调用 Close 可以正常关闭。
	wg sync.WaitGroup
}

func NewServer(a *assert.Assertion, o *server.Options) *Server {
	a.TB().Helper()

	s, o := newServer(a, o)
	a.NotNil(s).NotNil(o)

	return &Server{
		s:        s,
		a:        a,
		hostname: "http://localhost" + o.HTTPServer.Addr,
	}
}

func (s *Server) Server() *server.Server { return s.s }

func (s *Server) GoServe() {
	ok := make(chan struct{}, 1)
	s.wg.Add(1)

	go func() {
		s.a.TB().Helper()
		defer s.wg.Done()

		ok <- struct{}{} // 最起码等待协程启动

		err := s.Server().Serve()
		s.a.Error(err).ErrorIs(err, http.ErrServerClosed, "错误信息为:%v", err)
	}()

	<-ok
}

// Router 返回默认的路由
//
// 相当于：
//
//	s.Server().NewRouter("default", s.hostname, nil)
//
// NOTE: 如果需要多个路由，请使用 [Server.Server.Routers.New] 并指定正确的 group.Matcher 对象，
// 或是将 [Server.Router] 放在最后调用。
//
// 第一次调用时创建实例，多次调用返回同一个实例。
func (s *Server) Router() *server.Router {
	if s.r == nil {
		s.a.TB().Helper()
		rs := s.Server().Routers()
		s.r = rs.New("default", nil, mux.URLDomain(s.hostname), mux.WriterRecovery(http.StatusInternalServerError, os.Stderr))
		s.a.NotNil(s.r)
	}
	return s.r
}

// NewRequest 发起新的请求
//
// path 为请求路径，如果没有 `http://` 和 `https://` 前缀，则会自动加上 Server.hostname 作为其域名地址；
// client 如果为空，则采用 &http.Client{} 作为默认值；
func (s *Server) NewRequest(method, path string, client *http.Client) *rest.Request {
	if !strings.HasPrefix(path, "http://") && !strings.HasPrefix(path, "https://") {
		path = s.hostname + path
	}

	if client == nil {
		client = &http.Client{}
	}

	return rest.NewRequest(s.a, method, path).Client(client)
}

func (s *Server) Get(path string) *rest.Request {
	return s.NewRequest(http.MethodGet, path, nil)
}

func (s *Server) Delete(path string) *rest.Request {
	return s.NewRequest(http.MethodDelete, path, nil)
}

// Close 关闭测试服务
//
// NOTE: 会等待所有请求都退出之后，才会返回。
func (s *Server) Close(shutdown time.Duration) {
	// NOTE: Server 主要用于第三方测试，
	// 所以不主动将 Close 注册至 a.TB().Cleanup，由调用方决定何时调用。
	s.a.NotError(s.Server().Close(shutdown))
	s.wg.Wait()
}

// BuildHandler 生成以 code 作为状态码和内容输出的路由处理函数
func BuildHandler(code int) server.HandlerFunc {
	return func(ctx *server.Context) server.Responser {
		return server.ResponserFunc(func(ctx *server.Context) {
			ctx.Marshal(code, code, false)
		})
	}
}
