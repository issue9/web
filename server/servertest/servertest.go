// SPDX-License-Identifier: MIT

// Package servertest 针对 server 的测试用例
package servertest

import (
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/issue9/assert/v2"
	"github.com/issue9/assert/v2/rest"
	"github.com/issue9/mux/v6"

	"github.com/issue9/web/server"
)

type Tester struct {
	a        *assert.Assertion
	s        *server.Server
	hostname string
	wg       sync.WaitGroup
}

// NewTester 声明一个 server 实例
func NewTester(a *assert.Assertion, o *server.Options) *Tester {
	a.TB().Helper()

	s, o := newServer(a, o)
	a.NotNil(s).NotNil(o)

	return &Tester{
		s:        s,
		a:        a,
		hostname: "http://localhost" + o.Port,
	}
}

func (s *Tester) Server() *server.Server { return s.s }

func (s *Tester) GoServe() {
	s.wg.Add(1)
	go func() {
		s.a.TB().Helper()

		defer s.wg.Done()

		err := s.Server().Serve()
		s.a.Error(err).ErrorIs(err, http.ErrServerClosed, "错误信息为:%v", err)
	}()
}

// NewRouter 创建一个默认的路由
//
// 相当于：
//  s.Server().NewRouter("default", "http://localhost:8080/", nil)
//
// NOTE: 如果需要多个路由，请使用 Server().NewRouter 并指定正确的 group.Matcher 对象，
// 或是将 Tester.NewRouter 放在最后。
func (s *Tester) NewRouter(ms ...server.Middleware) *server.Router {
	s.a.TB().Helper()

	rs := s.Server().Routers()
	router := rs.New("default", nil, &server.RouterOptions{
		Middlewares: ms,
		URLDomain:   "http://localhost:8080/",
		RecoverFunc: mux.WriterRecovery(http.StatusInternalServerError, os.Stderr),
	})
	s.a.NotNil(router)
	return router
}

// Wait 等待 GoServe 退出
func (s *Tester) Wait() { s.wg.Wait() }

// NewRequest 发起新的请求
//
// path 为请求路径，如果没有 http:// 和 https:// 前缀，则会自动加上 http://localhost 作为其域名地址；
// client 如果为空，则采用 http.DefaultClient 作为默认值；
func (s *Tester) NewRequest(method, path string, client *http.Client) *rest.Request {
	if !strings.HasPrefix(path, "http://") && !strings.HasPrefix(path, "https://") {
		path = s.hostname + path
	}

	if client == nil {
		client = &http.Client{}
	}

	return rest.NewRequest(s.a, method, path).Client(client)
}

func (s *Tester) Get(path string) *rest.Request {
	return s.NewRequest(http.MethodGet, path, nil)
}

func (s *Tester) Delete(path string) *rest.Request {
	return s.NewRequest(http.MethodDelete, path, nil)
}

func (s *Tester) Close(shutdown time.Duration) {
	s.a.NotError(s.Server().Close(shutdown))
}

// BuildHandler 生成以 code 作为状态码和内容输出的路由处理函数
func BuildHandler(code int) server.HandlerFunc {
	return func(next *server.Context) server.Responser {
		return next.Object(code, []byte(strconv.Itoa(code)))
	}
}
