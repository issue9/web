// SPDX-FileCopyrightText: 2026 caixw
//
// SPDX-License-Identifier: MIT

package servertest

import (
	"context"
	"net/http"

	"github.com/issue9/assert/v5"
	"github.com/issue9/assert/v5/rest"

	"github.com/issue9/web"
)

// TestStartup 测试用的启动器
//
// 可与 server.NewHTTP 等初始化方法的 Options.Startup 字段使用。
type TestStartup struct {
	rs      *rest.Server
	creator func(h http.Handler) *rest.Server
}

// NewTestServerStartup 声明基于 [httptest.NewTestServer] 的启动器
func NewTestServerStartup(a *assert.Assertion) *TestStartup {
	return &TestStartup{
		creator: func(h http.Handler) *rest.Server { return rest.NewTestServer(a, h) },
	}
}

// NewServerStartup 声明基于 [httptest.NewServer] 的启动器
func NewServerStartup(a *assert.Assertion) *TestStartup {
	return &TestStartup{
		creator: func(h http.Handler) *rest.Server { return rest.NewServer(a, h) },
	}
}

// NewTLSServerStartup 声明基于 [httptest.NewTLSServer] 的启动器
func NewTLSServerStartup(a *assert.Assertion) *TestStartup {
	return &TestStartup{
		creator: func(h http.Handler) *rest.Server { return rest.NewTLSServer(a, h) },
	}
}

func (s *TestStartup) SetHandler(h http.Handler) { s.rs = s.creator(h) }

func (b *TestStartup) Serve(s *web.Server) error { return nil }

func (s *TestStartup) Close(ctx context.Context) error {
	s.rs.Close()
	return nil
}

func (s *TestStartup) NewRequest(method, path string) *rest.Request {
	return rest.NewRequest(s.rs.Assertion(), method, path).Client(s.rs.Server().Client())
}

func (s *TestStartup) Get(path string) *rest.Request { return s.NewRequest(http.MethodGet, path) }

func (s *TestStartup) Patch(path string) *rest.Request { return s.NewRequest(http.MethodPatch, path) }

func (s *TestStartup) Put(path string) *rest.Request { return s.NewRequest(http.MethodPut, path) }

func (s *TestStartup) Post(path string) *rest.Request { return s.NewRequest(http.MethodPost, path) }

func (s *TestStartup) Delete(path string) *rest.Request { return s.NewRequest(http.MethodDelete, path) }
