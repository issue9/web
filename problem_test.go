// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package web

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert/v6"
	"github.com/issue9/mux/v10/header"
	"github.com/issue9/mux/v10/types"
	"github.com/issue9/web/problems"
)

var _ Responser = &Problem{}

// 此函数放最前，内有依赖行数的测试，心意减少其行数的变化。
func TestContext_Log(t *testing.T) {
	a := assert.New(t, false)

	srv := newTestServer(a, false)

	t.Run("problemID=empty", func(t *testing.T) {
		a := assert.New(t, false)
		srv.logBuf.Reset()
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/path", nil)
		ctx := srv.NewContext(w, r, types.NewRoute())
		ctx.Log(errors.New("log1 log2"), "").Apply(ctx)
		a.Contains(srv.logBuf.String(), "problem_test.go:33"). // NOTE: 此测试依赖上一行的行号
									Contains(srv.logBuf.String(), "log1 log2").
									Contains(srv.logBuf.String(), header.XRequestID). // 包含 x-request-id 值
									Equal(w.Code, http.StatusInternalServerError)

		// problemError

		srv.logBuf.Reset()
		w = httptest.NewRecorder()
		r = httptest.NewRequest(http.MethodGet, "/path", nil)
		ctx = srv.NewContext(w, r, types.NewRoute())
		ctx.Log(NewProblemError(problems.BadRequest, errors.New("log1 log2")), "").Apply(ctx)
		a.Contains(srv.logBuf.String(), "problem_test.go:45"). // NOTE: 此测试依赖上一行的行号
									Contains(srv.logBuf.String(), "log1 log2").
									Contains(srv.logBuf.String(), header.XRequestID). // 包含 x-request-id 值
									Equal(w.Code, http.StatusBadRequest)
	})

	t.Run("problemID=41110", func(t *testing.T) {
		a := assert.New(t, false)
		srv.logBuf.Reset()
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/path", nil)
		ctx := srv.NewContext(w, r, types.NewRoute())
		ctx.Log(errors.New("log1 log2"), "41110").Apply(ctx)
		a.Contains(srv.logBuf.String(), "problem_test.go:58"). // NOTE: 此测试依赖上一行的行号
									Contains(srv.logBuf.String(), "log1 log2").
									Contains(srv.logBuf.String(), header.XRequestID). // 包含 x-request-id 值
									Equal(w.Code, 411)

		// ProblemError

		srv.logBuf.Reset()
		w = httptest.NewRecorder()
		r = httptest.NewRequest(http.MethodGet, "/path", nil)
		ctx = srv.NewContext(w, r, types.NewRoute())
		ctx.Log(NewProblemError(problems.BadRequest, errors.New("log1 log2")), "41110").Apply(ctx)
		a.Contains(srv.logBuf.String(), "problem_test.go:70") // NOTE: 此测试依赖上一行的行号
	})
}
