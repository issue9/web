// SPDX-License-Identifier: MIT

package web

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/assert/v3/rest"
	"golang.org/x/text/language"

	"github.com/issue9/web/filter"
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

func required[T any](v T) bool { return !reflect.ValueOf(v).IsZero() }

func min(v int) func(int) bool {
	return func(a int) bool { return a >= v }
}

func max(v int) func(int) bool {
	return func(a int) bool { return a < v }
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
		a.Contains(srv.logBuf.String(), "problem_test.go:52") // NOTE: 此测试依赖上一行的行号
		a.Contains(srv.logBuf.String(), "log1 log2")
		a.Contains(srv.logBuf.String(), requestIDKey) // 包含 x-request-id 值
		a.Equal(w.Code, http.StatusInternalServerError)

		// errs.HTTP

		srv.logBuf.Reset()
		w = httptest.NewRecorder()
		r = rest.Get(a, "/path").Request()
		ctx = srv.NewContext(w, r)
		ctx.Error(NewError(http.StatusBadRequest, errors.New("log1 log2")), "").Apply(ctx)
		a.Contains(srv.logBuf.String(), "problem_test.go:64") // NOTE: 此测试依赖上一行的行号
		a.Contains(srv.logBuf.String(), "log1 log2")
		a.Contains(srv.logBuf.String(), requestIDKey) // 包含 x-request-id 值
		a.Equal(w.Code, http.StatusBadRequest)
	})

	t.Run("id=41110", func(t *testing.T) {
		a := assert.New(t, false)
		srv.logBuf.Reset()
		w := httptest.NewRecorder()
		r := rest.Get(a, "/path").Request()
		ctx := srv.NewContext(w, r)
		ctx.Error(errors.New("log1 log2"), "41110").Apply(ctx)
		a.Contains(srv.logBuf.String(), "problem_test.go:77") // NOTE: 此测试依赖上一行的行号
		a.Contains(srv.logBuf.String(), "log1 log2")
		a.Contains(srv.logBuf.String(), requestIDKey) // 包含 x-request-id 值
		a.Equal(w.Code, 411)

		// errs.HTTP

		srv.logBuf.Reset()
		w = httptest.NewRecorder()
		r = rest.Get(a, "/path").Request()
		ctx = srv.NewContext(w, r)
		ctx.Error(NewError(http.StatusBadRequest, errors.New("log1 log2")), "41110").Apply(ctx)
		a.Contains(srv.logBuf.String(), "problem_test.go:89") // NOTE: 此测试依赖上一行的行号
		a.Contains(srv.logBuf.String(), "log1 log2")
		a.Contains(srv.logBuf.String(), requestIDKey) // 包含 x-request-id 值
		a.Equal(w.Code, 411)
	})
}

// TODO
func testContext_Problem(t *testing.T) {
	a := assert.New(t, false)
	srv := newTestServer(a)
	// a.NotError(srv.Catalog().SetString(language.Und, "lang", "und"))
	// a.NotError(srv.Catalog().SetString(language.SimplifiedChinese, "lang", "hans"))
	// srv.AddProblem("40000", 400, Phrase("lang"), Phrase("lang")) // lang 有翻译

	// 能正常翻译错误信息
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Set(header.AcceptLang, language.SimplifiedChinese.String())
	r.Header.Set(header.Accept, "application/json")
	ctx := srv.NewContext(w, r)
	resp := ctx.Problem("41110")
	resp.WithInstance("111")
	resp.Apply(ctx)
	a.Equal(w.Body.String(), `{"type":"41110","title":"hans","detail":"hans","instance":"111","status":400}`)

	// 未指定 accept-language，采用默认的 und
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").
		Header("accept", "application/json").
		Request()
	ctx = srv.NewContext(w, r)
	resp = ctx.Problem("40000")
	resp.WithInstance("111")
	resp.Apply(ctx)
	a.Equal(w.Body.String(), `{"type":"40000","title":"und","detail":"und","instance":"111","status":400}`)

	// 不存在的本地化信息，采用默认的 und
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").
		Header("accept-language", "en-US").
		Header("accept", "application/json").
		Request()
	ctx = srv.NewContext(w, r)
	resp = ctx.Problem("40000")
	resp.WithExtensions(&object{Age: 11})
	resp.WithInstance("111")
	resp.Apply(ctx)
	a.Equal(w.Body.String(), `{"type":"40000","title":"und","detail":"und","instance":"111","status":400,"extensions":{"Name":"","Age":11}}`)

	// 不存在
	a.Panic(func() { ctx.Problem("not-exists") })
	a.Panic(func() { ctx.Problem("50000") })

	// with field

	r = rest.Post(a, "/path", []byte("123")).
		Header("Accept", "application/json").
		Header("Content-Type", "application/json").
		Request()
	w = httptest.NewRecorder()
	ctx = newTestServer(a).NewContext(w, r)
	ctx.Server().AddProblem("40010", http.StatusBadRequest, Phrase("40010"), Phrase("40010"))
	ctx.Server().AddProblem("40011", http.StatusBadRequest, Phrase("40011"), Phrase("40011"))

	resp = ctx.Problem("40010")
	resp.WithParam("k1", "v1")
	resp.WithInstance("111")

	resp.Apply(ctx)
	a.Equal(w.Body.String(), `{"type":"40010","title":"40010","detail":"40010","instance":"111","status":400,"params":[{"name":"k1","reason":"v1"}]}`)
}

func TestContext_NewFilterProblem(t *testing.T) {
	a := assert.New(t, false)
	s := newTestServer(a)
	w := httptest.NewRecorder()
	r := rest.Get(a, "/path").Request()
	ctx := s.NewContext(w, r)

	min_2 := filter.NewRule(min(-2), Phrase("-2"))
	min_3 := filter.NewRule(min(-3), Phrase("-3"))
	max50 := filter.NewRule(max(50), Phrase("50"))
	max_4 := filter.NewRule(max(-4), Phrase("-4"))

	n100 := -100
	p100 := 100
	v := ctx.NewFilterProblem(false).
		AddFilter(filter.New(filter.NewRules(min_2, min_3))("f1", &n100)).
		AddFilter(filter.New(filter.NewRules(max50, max_4))("f2", &p100))
	a.Equal(v.p.Params, []RFC7807Param{
		{Name: "f1", Reason: "-2"},
		{Name: "f2", Reason: "50"},
	})

	n100 = -100
	p100 = 100
	v = ctx.NewFilterProblem(true).
		AddFilter(filter.New(filter.NewRules(min_2, min_3))("f1", &n100)).
		AddFilter(filter.New(filter.NewRules(max50, max_4))("f2", &p100))
	a.Equal(v.p.Params, []RFC7807Param{
		{Name: "f1", Reason: "-2"},
	})
}

func TestFilter_When(t *testing.T) {
	a := assert.New(t, false)
	s := newTestServer(a)
	w := httptest.NewRecorder()
	r := rest.Get(a, "/path").Request()
	ctx := s.NewContext(w, r)

	min18 := filter.NewRule(min(18), Phrase("不能小于 18"))
	notEmpty := filter.NewRule(required[string], Phrase("不能为空"))

	obj := &object{}
	v := ctx.NewFilterProblem(false).
		AddFilter(filter.New(min18)("obj/age", &obj.Age)).
		When(obj.Age > 0, func(v *FilterProblem) {
			v.AddFilter(filter.New(notEmpty)("obj/name", &obj.Name))
		})
	a.Equal(v.p.Params, []RFC7807Param{
		{Name: "obj/age", Reason: "不能小于 18"},
	})

	obj = &object{Age: 15}
	v = ctx.NewFilterProblem(false).
		AddFilter(filter.New(min18)("obj/age", &obj.Age)).
		When(obj.Age > 0, func(v *FilterProblem) {
			v.AddFilter(filter.New(notEmpty)("obj/name", &obj.Name))
		})
	a.Equal(v.p.Params, []RFC7807Param{
		{Name: "obj/age", Reason: "不能小于 18"},
		{Name: "obj/name", Reason: "不能为空"},
	})
}
