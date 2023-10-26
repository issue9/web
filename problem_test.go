// SPDX-License-Identifier: MIT

package web

import (
	"errors"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert/v3"

	"github.com/issue9/web/internal/header"
)

var (
	_ Problem = &RFC7807{}
	_ error   = &RFC7807{}
)

type object struct {
	Name string
	Age  int
}

// 此函数放最前，内有依赖行数的测试，心意减少其行数的变化。
func TestContext_Error(t *testing.T) {
	a := assert.New(t, false)

	srv := newTestServer(a)

	t.Run("id=empty", func(t *testing.T) {
		a := assert.New(t, false)
		srv.logBuf.Reset()
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/path", nil)
		ctx := srv.NewContext(w, r)
		ctx.Error(errors.New("log1 log2"), "").Apply(ctx)
		a.Contains(srv.logBuf.String(), "problem_test.go:39") // NOTE: 此测试依赖上一行的行号
		a.Contains(srv.logBuf.String(), "log1 log2")
		a.Contains(srv.logBuf.String(), header.RequestIDKey) // 包含 x-request-id 值
		a.Equal(w.Code, http.StatusInternalServerError)

		// errs.HTTP

		srv.logBuf.Reset()
		w = httptest.NewRecorder()
		r = httptest.NewRequest(http.MethodGet, "/path", nil)
		ctx = srv.NewContext(w, r)
		ctx.Error(NewError(http.StatusBadRequest, errors.New("log1 log2")), "").Apply(ctx)
		a.Contains(srv.logBuf.String(), "problem_test.go:51") // NOTE: 此测试依赖上一行的行号
		a.Contains(srv.logBuf.String(), "log1 log2")
		a.Contains(srv.logBuf.String(), header.RequestIDKey) // 包含 x-request-id 值
		a.Equal(w.Code, http.StatusBadRequest)

		// fs.ErrPermission

		srv.logBuf.Reset()
		w = httptest.NewRecorder()
		r = httptest.NewRequest(http.MethodGet, "/path", nil)
		ctx = srv.NewContext(w, r)
		ctx.Error(fs.ErrPermission, "").Apply(ctx)
		a.Equal(w.Code, http.StatusForbidden)

		// fs.ErrNotExist

		srv.logBuf.Reset()
		w = httptest.NewRecorder()
		r = httptest.NewRequest(http.MethodGet, "/path", nil)
		ctx = srv.NewContext(w, r)
		ctx.Error(fs.ErrNotExist, "").Apply(ctx)
		a.Equal(w.Code, http.StatusNotFound)
	})

	t.Run("id=41110", func(t *testing.T) {
		a := assert.New(t, false)
		srv.logBuf.Reset()
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/path", nil)
		ctx := srv.NewContext(w, r)
		ctx.Error(errors.New("log1 log2"), "41110").Apply(ctx)
		a.Contains(srv.logBuf.String(), "problem_test.go:82") // NOTE: 此测试依赖上一行的行号
		a.Contains(srv.logBuf.String(), "log1 log2")
		a.Contains(srv.logBuf.String(), header.RequestIDKey) // 包含 x-request-id 值
		a.Equal(w.Code, 411)

		// errs.HTTP

		srv.logBuf.Reset()
		w = httptest.NewRecorder()
		r = httptest.NewRequest(http.MethodGet, "/path", nil)
		ctx = srv.NewContext(w, r)
		ctx.Error(NewError(http.StatusBadRequest, errors.New("log1 log2")), "41110").Apply(ctx)
		a.Contains(srv.logBuf.String(), "problem_test.go:94") // NOTE: 此测试依赖上一行的行号
		a.Contains(srv.logBuf.String(), "log1 log2")
		a.Contains(srv.logBuf.String(), header.RequestIDKey) // 包含 x-request-id 值
		a.Equal(w.Code, 411)
	})
}
