// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

// Package servertest 为测试 [web.Server] 提供一些简便的功能
package servertest

import (
	"net/http"

	"github.com/issue9/assert/v4"
	"github.com/issue9/assert/v4/rest"

	"github.com/issue9/web"
)

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

func Patch(a *assert.Assertion, path string, body []byte) *rest.Request {
	return NewRequest(a, http.MethodPatch, path).Body(body)
}

func Delete(a *assert.Assertion, path string) *rest.Request {
	return NewRequest(a, http.MethodDelete, path)
}

// Run 运行服务内容并返回等待退出的方法
func Run(a *assert.Assertion, s web.Server) func() {
	ok := make(chan struct{}, 1)
	exit := make(chan struct{}, 1)

	a.Go(func(a *assert.Assertion) {
		defer func() { exit <- struct{}{} }()

		ok <- struct{}{} // 最起码等待协程启动
		a.ErrorIs(s.Serve(), http.ErrServerClosed)
	})

	<-ok
	return func() { <-exit }
}
