// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package web

import (
	"errors"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert/v4"
	"github.com/issue9/mux/v8/header"
	"github.com/issue9/mux/v8/types"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var (
	_ Responser = &Problem{}
	_ error     = &Problem{}
)

// 此函数放最前，内有依赖行数的测试，心意减少其行数的变化。
func TestContext_Error(t *testing.T) {
	a := assert.New(t, false)

	srv := newTestServer(a)

	t.Run("id=empty", func(t *testing.T) {
		a := assert.New(t, false)
		srv.logBuf.Reset()
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/path", nil)
		ctx := srv.NewContext(w, r, types.NewContext())
		ctx.Error(errors.New("log1 log2"), "").Apply(ctx)
		a.Contains(srv.logBuf.String(), "problem_test.go:38"). // NOTE: 此测试依赖上一行的行号
									Contains(srv.logBuf.String(), "log1 log2").
									Contains(srv.logBuf.String(), header.XRequestID). // 包含 x-request-id 值
									Equal(w.Code, http.StatusInternalServerError)

		// errs.HTTP

		srv.logBuf.Reset()
		w = httptest.NewRecorder()
		r = httptest.NewRequest(http.MethodGet, "/path", nil)
		ctx = srv.NewContext(w, r, types.NewContext())
		ctx.Error(NewError(http.StatusBadRequest, errors.New("log1 log2")), "").Apply(ctx)
		a.Contains(srv.logBuf.String(), "problem_test.go:50"). // NOTE: 此测试依赖上一行的行号
									Contains(srv.logBuf.String(), "log1 log2").
									Contains(srv.logBuf.String(), header.XRequestID). // 包含 x-request-id 值
									Equal(w.Code, http.StatusBadRequest)

		// fs.ErrPermission

		srv.logBuf.Reset()
		w = httptest.NewRecorder()
		r = httptest.NewRequest(http.MethodGet, "/path", nil)
		ctx = srv.NewContext(w, r, types.NewContext())
		ctx.Error(fs.ErrPermission, "").Apply(ctx)
		a.Equal(w.Code, http.StatusForbidden)

		// fs.ErrNotExist

		srv.logBuf.Reset()
		w = httptest.NewRecorder()
		r = httptest.NewRequest(http.MethodGet, "/path", nil)
		ctx = srv.NewContext(w, r, types.NewContext())
		ctx.Error(fs.ErrNotExist, "").Apply(ctx)
		a.Equal(w.Code, http.StatusNotFound)
	})

	t.Run("id=41110", func(t *testing.T) {
		a := assert.New(t, false)
		srv.logBuf.Reset()
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/path", nil)
		ctx := srv.NewContext(w, r, types.NewContext())
		ctx.Error(errors.New("log1 log2"), "41110").Apply(ctx)
		a.Contains(srv.logBuf.String(), "problem_test.go:81"). // NOTE: 此测试依赖上一行的行号
									Contains(srv.logBuf.String(), "log1 log2").
									Contains(srv.logBuf.String(), header.XRequestID). // 包含 x-request-id 值
									Equal(w.Code, 411)

		// errs.HTTP

		srv.logBuf.Reset()
		w = httptest.NewRecorder()
		r = httptest.NewRequest(http.MethodGet, "/path", nil)
		ctx = srv.NewContext(w, r, types.NewContext())
		ctx.Error(NewError(http.StatusBadRequest, errors.New("log1 log2")), "41110").Apply(ctx)
		a.Contains(srv.logBuf.String(), "problem_test.go:93"). // NOTE: 此测试依赖上一行的行号
									Contains(srv.logBuf.String(), "log1 log2").
									Contains(srv.logBuf.String(), header.XRequestID). // 包含 x-request-id 值
									Equal(w.Code, 411)
	})
}

func TestProblems_Add(t *testing.T) {
	a := assert.New(t, false)

	ps := newProblems("")
	a.NotNil(ps)
	l := len(ps.problems)

	a.False(ps.exists("40010")).
		False(ps.exists("40011")).
		True(ps.exists(ProblemNotFound))

	ps.Add(400, []*LocaleProblem{
		{ID: "40010", Title: Phrase("title"), Detail: Phrase("detail")},
		{ID: "40011", Title: Phrase("title"), Detail: Phrase("detail")},
	}...)
	a.True(ps.exists("40010")).
		True(ps.exists("40011")).
		Equal(l+2, len(ps.problems))

	a.PanicString(func() {
		ps.Add(400, []*LocaleProblem{
			{ID: "40010", Title: Phrase("title"), Detail: Phrase("detail")},
		}...)
	}, "存在相同值的 id 参数")

	a.PanicString(func() {
		ps.Add(99, []*LocaleProblem{
			{ID: "40012", Title: Phrase("title"), Detail: Phrase("detail")},
		}...)
	}, "status 必须是一个有效的状态码")

	a.PanicString(func() {
		ps.Add(412, []*LocaleProblem{
			{ID: "40013", Title: nil, Detail: nil},
		}...)
	}, "title 不能为空")
}

func TestProblems_initProblem(t *testing.T) {
	a := assert.New(t, false)
	p := message.NewPrinter(language.SimplifiedChinese)

	ps := newProblems("")
	a.NotNil(ps)
	ps.Add(400, &LocaleProblem{ID: "40010", Title: Phrase("title"), Detail: Phrase("detail")})
	pp := &Problem{}
	ps.initProblem(pp, "40010", p)
	a.Equal(pp.Type, "40010")

	ps = newProblems("https://example.com/qa#")
	a.NotNil(ps)
	ps.Add(400, &LocaleProblem{ID: "40011", Title: Phrase("title"), Detail: Phrase("detail")})
	pp = &Problem{}
	ps.initProblem(pp, "40011", p)
	a.Equal(pp.Type, "https://example.com/qa#40011").
		Equal(ps.Prefix(), "https://example.com/qa#")

	ps = newProblems(ProblemAboutBlank)
	a.NotNil(ps)
	ps.Add(400, &LocaleProblem{ID: "40012", Title: Phrase("title"), Detail: Phrase("detail")})
	pp = &Problem{}
	ps.initProblem(pp, "40012", p)
	a.Equal(pp.Type, ProblemAboutBlank)

	a.PanicString(func() {
		ps.initProblem(pp, "not-exists", p)
	}, "未找到有关 not-exists 的定义")
}
