// SPDX-FileCopyrightText: 2026 caixw
//
// SPDX-License-Identifier: MIT

package startuptest

import (
	"context"
	"net/http"

	"github.com/issue9/assert/v6"
	"github.com/issue9/assert/v6/rest"

	"github.com/issue9/web"
)

type testStartup struct {
	rs      *rest.Server
	creator func(h http.Handler) *rest.Server
}

// NewTestServerStartup 声明基于 [httptest.NewTestServer] 的启动器
func NewTestServerStartup(a *assert.Assertion) Startup {
	return &testStartup{
		creator: func(h http.Handler) *rest.Server { return rest.NewTestServer(a, h) },
	}
}

// NewServerStartup 声明基于 [httptest.NewServer] 的启动器
func NewServerStartup(a *assert.Assertion) Startup {
	return &testStartup{
		creator: func(h http.Handler) *rest.Server { return rest.NewServer(a, h) },
	}
}

// NewTLSServerStartup 声明基于 [httptest.NewTLSServer] 的启动器
func NewTLSServerStartup(a *assert.Assertion) Startup {
	return &testStartup{
		creator: func(h http.Handler) *rest.Server { return rest.NewTLSServer(a, h) },
	}
}

func (b *testStartup) Init(h http.Handler) { b.rs = b.creator(h) }

func (b *testStartup) Serve(s *web.Server) error { return nil }

func (s *testStartup) Close(ctx context.Context) error {
	s.rs.Close()
	return nil
}

func (s *testStartup) URL() string { return s.rs.URL() }

func (s *testStartup) Client() *http.Client { return s.rs.Server().Client() }

func (s *testStartup) Assertion() *assert.Assertion { return s.rs.Assertion() }

func (s *testStartup) NewRequest(method, path string) *rest.Request {
	return s.rs.NewRequest(method, path)
}

func (s *testStartup) Get(path string) *rest.Request { return s.NewRequest(http.MethodGet, path) }

func (s *testStartup) Put(path string) *rest.Request { return s.NewRequest(http.MethodPut, path) }

func (s *testStartup) Post(path string) *rest.Request { return s.NewRequest(http.MethodPost, path) }

func (s *testStartup) Patch(path string) *rest.Request { return s.NewRequest(http.MethodPatch, path) }

func (s *testStartup) Delete(path string) *rest.Request { return s.NewRequest(http.MethodDelete, path) }
