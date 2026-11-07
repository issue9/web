// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package web

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/issue9/assert/v6"
	"github.com/issue9/mux/v10/header"
	"github.com/issue9/mux/v10/types"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
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
		a.Contains(srv.logBuf.String(), "problem_test.go:35"). // NOTE: 此测试依赖上一行的行号
									Contains(srv.logBuf.String(), "log1 log2").
									Contains(srv.logBuf.String(), header.XRequestID). // 包含 x-request-id 值
									Equal(w.Code, http.StatusInternalServerError)

		// problemError

		srv.logBuf.Reset()
		w = httptest.NewRecorder()
		r = httptest.NewRequest(http.MethodGet, "/path", nil)
		ctx = srv.NewContext(w, r, types.NewRoute())
		ctx.Log(NewProblemError(ProblemBadRequest, errors.New("log1 log2")), "").Apply(ctx)
		a.Contains(srv.logBuf.String(), "problem_test.go:47"). // NOTE: 此测试依赖上一行的行号
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
		a.Contains(srv.logBuf.String(), "problem_test.go:60"). // NOTE: 此测试依赖上一行的行号
									Contains(srv.logBuf.String(), "log1 log2").
									Contains(srv.logBuf.String(), header.XRequestID). // 包含 x-request-id 值
									Equal(w.Code, 411)

		// ProblemError

		srv.logBuf.Reset()
		w = httptest.NewRecorder()
		r = httptest.NewRequest(http.MethodGet, "/path", nil)
		ctx = srv.NewContext(w, r, types.NewRoute())
		ctx.Log(NewProblemError(ProblemBadRequest, errors.New("log1 log2")), "41110").Apply(ctx)
		a.Contains(srv.logBuf.String(), "problem_test.go:72") // NOTE: 此测试依赖上一行的行号
	})
}

func TestProblems_Add(t *testing.T) {
	a := assert.New(t, false)

	ps := newProblems()
	a.NotNil(ps)
	l := len(ps.problems)

	a.False(ps.Exists("40010")).
		False(ps.Exists("40011")).
		True(ps.Exists(ProblemNotFound))

	ps.Add("", []*LocaleProblem{
		{Status: 400, ID: "40010", Title: Phrase("title"), Detail: Phrase("detail")},
		{Status: 400, ID: "40011", Title: Phrase("title"), Detail: Phrase("detail")},
	}...)
	a.True(ps.Exists("40010")).
		True(ps.Exists("40011")).
		Equal(l+2, len(ps.problems)).
		Length(slices.Collect(ps.All()), l+2)

	a.PanicString(func() {
		ps.Add("", []*LocaleProblem{
			{Status: 400, ID: "40010", Title: Phrase("title"), Detail: Phrase("detail")},
		}...)
	}, "该 ID 值 40010 已经存在")

	a.PanicString(func() {
		ps.Add("", []*LocaleProblem{
			{Status: 99, ID: "40012", Title: Phrase("title"), Detail: Phrase("detail")},
		}...)
	}, "status 必须是一个有效的状态码")

	a.PanicString(func() {
		ps.Add("", []*LocaleProblem{
			{Status: 400, ID: "40013", Title: nil, Detail: nil},
		}...)
	}, "title 不能为空")
}

func TestProblems_initProblem(t *testing.T) {
	a := assert.New(t, false)
	p := message.NewPrinter(language.SimplifiedChinese)

	ps := newProblems()
	a.NotNil(ps)
	ps.Add("", &LocaleProblem{Status: 400, ID: "40010", Title: Phrase("title"), Detail: Phrase("detail")})
	pp := &Problem{}
	ps.initProblem(pp, "40010", p)
	a.Equal(pp.Type, aboutBlank)

	ps = newProblems()
	a.NotNil(ps)
	ps.Add("", &LocaleProblem{Status: 400, ID: "40011", Type: "https://example.com/qa#40011", Title: Phrase("title"), Detail: Phrase("detail")})
	ps.Add("https://example.com/qa#", &LocaleProblem{Status: 400, ID: "40012", Type: "40012", Title: Phrase("title"), Detail: Phrase("detail")})
	pp = &Problem{}
	ps.initProblem(pp, "40011", p)
	a.Equal(pp.Type, "https://example.com/qa#40011")
	pp = &Problem{}
	ps.initProblem(pp, "40012", p)
	a.Equal(pp.Type, "https://example.com/qa#40012")

	a.PanicString(func() {
		ps.initProblem(pp, "not-exists", p)
	}, "未找到有关 not-exists 的定义")
}
