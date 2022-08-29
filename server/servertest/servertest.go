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

	"github.com/issue9/assert/v3"
	"github.com/issue9/assert/v3/rest"
	"github.com/issue9/mux/v7"

	"github.com/issue9/web/server"
)

type Tester struct {
	a        *assert.Assertion
	s        *server.Server
	hostname string
	wg       sync.WaitGroup
}

// NewTester 声明一个 [Tester] 实例
func NewTester(a *assert.Assertion, o *server.Options) *Tester {
	a.TB().Helper()

	s, o := newServer(a, o)
	a.NotNil(s).NotNil(o)

	return &Tester{
		s:        s,
		a:        a,
		hostname: "http://localhost" + o.HTTPServer.Addr,
	}
}

func (s *Tester) Server() *server.Server { return s.s }

func (s *Tester) GoServe() {
	s.wg.Add(1)
	ok := make(chan struct{}, 1)

	go func() {
		s.a.TB().Helper()

		defer s.wg.Done()

		ok <- struct{}{} // 最起码等待协程启动

		err := s.Server().Serve()
		s.a.Error(err).ErrorIs(err, http.ErrServerClosed, "错误信息为:%v", err)
	}()

	<-ok
}

// NewRouter 创建一个默认的路由
//
// 相当于：
//
//	s.Server().NewRouter("default", "http://localhost:8080/", nil)
//
// NOTE: 如果需要多个路由，请使用 Server().Routers().NewRouter() 并指定正确的 group.Matcher 对象，
// 或是将 Tester.NewRouter 放在最后。
func (s *Tester) NewRouter(ms ...server.Middleware) *server.Router {
	s.a.TB().Helper()

	rs := s.Server().Routers()
	router := rs.New("default", nil, mux.URLDomain("http://localhost:8080/"), mux.WriterRecovery(http.StatusInternalServerError, os.Stderr))
	s.a.NotNil(router)
	router.Use(ms...)
	return router
}

// NewRequest 发起新的请求
//
// path 为请求路径，如果没有 `http://` 和 `https://` 前缀，则会自动加上 `http://localhost“ 作为其域名地址；
// client 如果为空，则采用 &http.Client{} 作为默认值；
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

// Close 关闭测试服务
//
// NOTE: 会等待所有请求都退出之后，才会返回。
func (s *Tester) Close(shutdown time.Duration) {
	// NOTE: Tester 主要用于第三方测试，
	// 所以不主动将 Close 注册至 a.TB().Cleanup，由调用方决定何时调用。
	s.a.NotError(s.Server().Close(shutdown))
	s.wg.Wait()
}

// BuildHandler 生成以 code 作为状态码和内容输出的路由处理函数
func BuildHandler(code int) server.HandlerFunc {
	return func(ctx *server.Context) server.Responser {
		return server.ResponserFunc(func(ctx *server.Context) {
			if err := ctx.Marshal(code, []byte(strconv.Itoa(code)), false); err != nil {
				ctx.Logs().ERROR().Error(err)
			}
		})
	}
}
