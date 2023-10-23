// SPDX-License-Identifier: MIT

package web

import (
	"errors"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/assert/v3/rest"

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
		r := rest.Get(a, "/path").Request()
		ctx := srv.NewContext(w, r)
		ctx.Error(errors.New("log1 log2"), "").Apply(ctx)
		a.Contains(srv.logBuf.String(), "problem_test.go:40") // NOTE: 此测试依赖上一行的行号
		a.Contains(srv.logBuf.String(), "log1 log2")
		a.Contains(srv.logBuf.String(), header.RequestIDKey) // 包含 x-request-id 值
		a.Equal(w.Code, http.StatusInternalServerError)

		// errs.HTTP

		srv.logBuf.Reset()
		w = httptest.NewRecorder()
		r = rest.Get(a, "/path").Request()
		ctx = srv.NewContext(w, r)
		ctx.Error(NewError(http.StatusBadRequest, errors.New("log1 log2")), "").Apply(ctx)
		a.Contains(srv.logBuf.String(), "problem_test.go:52") // NOTE: 此测试依赖上一行的行号
		a.Contains(srv.logBuf.String(), "log1 log2")
		a.Contains(srv.logBuf.String(), header.RequestIDKey) // 包含 x-request-id 值
		a.Equal(w.Code, http.StatusBadRequest)

		// fs.ErrPermission

		srv.logBuf.Reset()
		w = httptest.NewRecorder()
		r = rest.Get(a, "/path").Request()
		ctx = srv.NewContext(w, r)
		ctx.Error(fs.ErrPermission, "").Apply(ctx)
		a.Equal(w.Code, http.StatusForbidden)

		// fs.ErrNotExist

		srv.logBuf.Reset()
		w = httptest.NewRecorder()
		r = rest.Get(a, "/path").Request()
		ctx = srv.NewContext(w, r)
		ctx.Error(fs.ErrNotExist, "").Apply(ctx)
		a.Equal(w.Code, http.StatusNotFound)
	})

	t.Run("id=41110", func(t *testing.T) {
		a := assert.New(t, false)
		srv.logBuf.Reset()
		w := httptest.NewRecorder()
		r := rest.Get(a, "/path").Request()
		ctx := srv.NewContext(w, r)
		ctx.Error(errors.New("log1 log2"), "41110").Apply(ctx)
		a.Contains(srv.logBuf.String(), "problem_test.go:83") // NOTE: 此测试依赖上一行的行号
		a.Contains(srv.logBuf.String(), "log1 log2")
		a.Contains(srv.logBuf.String(), header.RequestIDKey) // 包含 x-request-id 值
		a.Equal(w.Code, 411)

		// errs.HTTP

		srv.logBuf.Reset()
		w = httptest.NewRecorder()
		r = rest.Get(a, "/path").Request()
		ctx = srv.NewContext(w, r)
		ctx.Error(NewError(http.StatusBadRequest, errors.New("log1 log2")), "41110").Apply(ctx)
		a.Contains(srv.logBuf.String(), "problem_test.go:95") // NOTE: 此测试依赖上一行的行号
		a.Contains(srv.logBuf.String(), "log1 log2")
		a.Contains(srv.logBuf.String(), header.RequestIDKey) // 包含 x-request-id 值
		a.Equal(w.Code, 411)
	})
}

func TestFilterProblem(t *testing.T) {
	a := assert.New(t, false)
	s := newTestServer(a)
	w := httptest.NewRecorder()
	r := rest.Get(a, "/path").Request()
	ctx := s.NewContext(w, r)

	min_2 := NewRule(min(-2), Phrase("-2"))
	min_3 := NewRule(min(-3), Phrase("-3"))
	max50 := NewRule(max(50), Phrase("50"))
	max_4 := NewRule(max(-4), Phrase("-4"))

	n100 := -100
	p100 := 100
	v := ctx.newFilterProblem(false).
		AddFilter(NewFilter(NewRules(min_2, min_3))("f1", &n100)).
		AddFilter(NewFilter(NewRules(max50, max_4))("f2", &p100))
	a.Equal(v.p.Params, []RFC7807Param{
		{Name: "f1", Reason: "-2"},
		{Name: "f2", Reason: "50"},
	})

	n100 = -100
	p100 = 100
	v = ctx.newFilterProblem(true).
		AddFilter(NewFilter(NewRules(min_2, min_3))("f1", &n100)).
		AddFilter(NewFilter(NewRules(max50, max_4))("f2", &p100))
	a.Equal(v.p.Params, []RFC7807Param{
		{Name: "f1", Reason: "-2"},
	})
}

func TestFilterProblem_New(t *testing.T) {
	a := assert.New(t, false)
	s := newTestServer(a)
	w := httptest.NewRecorder()
	r := rest.Get(a, "/path").Request()
	ctx := s.NewContext(w, r)

	v := ctx.newFilterProblem(false)
	v1 := v.New("v1.", func(f *FilterProblem) {
		f.Add("f1", StringPhrase("s1"))
		v2 := f.New("v2.", func(f *FilterProblem) {
			f.AddError("f2", errors.New("s2"))
		})
		a.Equal(v2, f)
	})
	a.Equal(v1, v)

	a.Equal(v.p.Params, []RFC7807Param{
		{Name: "v1.f1", Reason: "s1"},
		{Name: "v1.v2.f2", Reason: "s2"},
	})
}

func TestFilter_When(t *testing.T) {
	a := assert.New(t, false)
	s := newTestServer(a)
	w := httptest.NewRecorder()
	r := rest.Get(a, "/path").Request()
	ctx := s.NewContext(w, r)

	min18 := NewRule(min(18), Phrase("不能小于 18"))
	notEmpty := NewRule(required[string], Phrase("不能为空"))

	obj := &object{}
	v := ctx.newFilterProblem(false).
		AddFilter(NewFilter(min18)("obj/age", &obj.Age)).
		When(obj.Age > 0, func(v *FilterProblem) {
			v.AddFilter(NewFilter(notEmpty)("obj/name", &obj.Name))
		})
	a.Equal(v.p.Params, []RFC7807Param{
		{Name: "obj/age", Reason: "不能小于 18"},
	})

	obj = &object{Age: 15}
	v = ctx.newFilterProblem(false).
		AddFilter(NewFilter(min18)("obj/age", &obj.Age)).
		When(obj.Age > 0, func(v *FilterProblem) {
			v.AddFilter(NewFilter(notEmpty)("obj/name", &obj.Name))
		})
	a.Equal(v.p.Params, []RFC7807Param{
		{Name: "obj/age", Reason: "不能小于 18"},
		{Name: "obj/name", Reason: "不能为空"},
	})
}
