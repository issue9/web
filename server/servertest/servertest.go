// SPDX-License-Identifier: MIT

// Package servertest 为 server 提供一些简便的测试方
package servertest

import (
	"net/http"

	"github.com/issue9/assert/v3"
	"github.com/issue9/assert/v3/rest"
)

type Server interface {
	Serve() error
}

// NewRequest 发起测试请求
//
// 功能与 [rest.NewRequest] 相同，默认指定了 client 参数。
func NewRequest(a *assert.Assertion, method, path string) *rest.Request {
	return rest.NewRequest(a, method, path).Client(&http.Client{})
}

func Get(a *assert.Assertion, path string) *rest.Request {
	return NewRequest(a, http.MethodGet, path)
}

func Post(a *assert.Assertion, path string, body []byte) *rest.Request {
	return NewRequest(a, http.MethodPost, path).Body(body)
}

func Delete(a *assert.Assertion, path string) *rest.Request {
	return NewRequest(a, http.MethodDelete, path)
}

// Run 运行服务内容并返回等待退出的方法
func Run(a *assert.Assertion, s Server) func() {
	ok := make(chan struct{}, 1)
	exit := make(chan struct{}, 1)

	go func() {
		a.TB().Helper()

		defer func() { exit <- struct{}{} }()

		ok <- struct{}{} // 最起码等待协程启动，在 s.Serve 之前运行。
		err := s.Serve()
		a.ErrorIs(err, http.ErrServerClosed, "错误信息为:%v", err)
	}()

	<-ok
	return func() { <-exit }
}
