// SPDX-License-Identifier: MIT

package server

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/assert/v3/rest"
	"github.com/issue9/localeutil"
	"golang.org/x/text/language"

	"github.com/issue9/web/filter"
	"github.com/issue9/web/internal/errs"
	"github.com/issue9/web/logs"
)

var _ BuildProblemFunc = RFC7807Builder

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
	errLog := new(bytes.Buffer)

	srv := newTestServer(a, &Options{
		Logs: &logs.Options{Handler: logs.NewTextHandler("20060102-15:04:05", errLog), Caller: true, Created: true},
	})
	errLog.Reset()

	t.Run("id=empty", func(t *testing.T) {
		a := assert.New(t, false)
		errLog.Reset()
		w := httptest.NewRecorder()
		r := rest.Get(a, "/path").Request()
		ctx := srv.newContext(w, r, nil)
		ctx.Error(errors.New("log1 log2"), "").Apply(ctx)
		a.Contains(errLog.String(), "problem_test.go:56") // NOTE: 此测试依赖上一行的行号
		a.Contains(errLog.String(), "log1 log2")
		a.Contains(errLog.String(), srv.requestIDKey) // 包含 x-request-id 值
		a.Equal(w.Code, http.StatusInternalServerError)

		// errs.HTTP

		errLog.Reset()
		w = httptest.NewRecorder()
		r = rest.Get(a, "/path").Request()
		ctx = srv.newContext(w, r, nil)
		ctx.Error(errs.NewHTTPError(http.StatusBadRequest, errors.New("log1 log2")), "").Apply(ctx)
		a.Contains(errLog.String(), "problem_test.go:68") // NOTE: 此测试依赖上一行的行号
		a.Contains(errLog.String(), "log1 log2")
		a.Contains(errLog.String(), srv.requestIDKey) // 包含 x-request-id 值
		a.Equal(w.Code, http.StatusBadRequest)
	})

	t.Run("id=41110", func(t *testing.T) {
		a := assert.New(t, false)
		errLog.Reset()
		w := httptest.NewRecorder()
		r := rest.Get(a, "/path").Request()
		ctx := srv.newContext(w, r, nil)
		ctx.Error(errors.New("log1 log2"), "41110").Apply(ctx)
		a.Contains(errLog.String(), "problem_test.go:81") // NOTE: 此测试依赖上一行的行号
		a.Contains(errLog.String(), "log1 log2")
		a.Contains(errLog.String(), srv.requestIDKey) // 包含 x-request-id 值
		a.Equal(w.Code, 411)

		// errs.HTTP

		errLog.Reset()
		w = httptest.NewRecorder()
		r = rest.Get(a, "/path").Request()
		ctx = srv.newContext(w, r, nil)
		ctx.Error(errs.NewHTTPError(http.StatusBadRequest, errors.New("log1 log2")), "41110").Apply(ctx)
		a.Contains(errLog.String(), "problem_test.go:93") // NOTE: 此测试依赖上一行的行号
		a.Contains(errLog.String(), "log1 log2")
		a.Contains(errLog.String(), srv.requestIDKey) // 包含 x-request-id 值
		a.Equal(w.Code, 411)
	})
}

func TestContext_Problem(t *testing.T) {
	a := assert.New(t, false)
	srv := newTestServer(a, nil)
	a.NotError(srv.CatalogBuilder().SetString(language.Und, "lang", "und"))
	a.NotError(srv.CatalogBuilder().SetString(language.SimplifiedChinese, "lang", "hans"))
	srv.AddProblem("40000", 400, localeutil.Phrase("lang"), localeutil.Phrase("lang")) // lang 有翻译

	// 能正常翻译错误信息
	w := httptest.NewRecorder()
	r := rest.Get(a, "/path").
		Header("accept-language", language.SimplifiedChinese.String()).
		Header("accept", "application/json").
		Request()
	ctx := srv.newContext(w, r, nil)
	resp := ctx.Problem("40000")
	resp.Apply(ctx)
	a.Equal(w.Body.String(), `{"type":"40000","title":"hans","status":400,"detail":"hans"}`)

	// 未指定 accept-language，采用默认的 und
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").
		Header("accept", "application/json").
		Request()
	ctx = srv.newContext(w, r, nil)
	resp = ctx.Problem("40000")
	resp.Apply(ctx)
	a.Equal(w.Body.String(), `{"type":"40000","title":"und","status":400,"detail":"und"}`)

	// 不存在的本地化信息，采用默认的 und
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").
		Header("accept-language", "en-US").
		Header("accept", "application/json").
		Request()
	ctx = srv.newContext(w, r, nil)
	resp = ctx.Problem("40000")
	resp.With("with", "abc")
	resp.Apply(ctx)
	a.Equal(w.Body.String(), `{"type":"40000","title":"und","status":400,"detail":"und","with":"abc"}`)

	// 不存在
	a.Panic(func() { ctx.Problem("not-exists") })
	a.Panic(func() { ctx.Problem("50000") })

	// with field

	r = rest.Post(a, "/path", []byte("123")).
		Header("Accept", "application/json").
		Header("Content-Type", "application/json").
		Request()
	w = httptest.NewRecorder()
	ctx = newTestServer(a, nil).newContext(w, r, nil)
	ctx.Server().AddProblem("40010", http.StatusBadRequest, localeutil.Phrase("40010"), localeutil.Phrase("40010")).
		AddProblem("40011", http.StatusBadRequest, localeutil.Phrase("40011"), localeutil.Phrase("40011"))

	resp = ctx.Problem("40010")
	resp.With("detail1", "40010")
	resp.AddParam("k1", "v1")

	resp.Apply(ctx)
	a.Equal(w.Body.String(), `{"type":"40010","title":"40010","status":400,"detail":"40010","detail1":"40010","params":[{"name":"k1","reason":"v1"}]}`)
}

func TestContext_NewFilterProblem(t *testing.T) {
	a := assert.New(t, false)
	s := newTestServer(a, nil)
	w := httptest.NewRecorder()
	r := rest.Get(a, "/path").Request()
	ctx := s.newContext(w, r, nil)

	min_2 := filter.NewRule(min(-2), localeutil.Phrase("-2"))
	min_3 := filter.NewRule(min(-3), localeutil.Phrase("-3"))
	max50 := filter.NewRule(max(50), localeutil.Phrase("50"))
	max_4 := filter.NewRule(max(-4), localeutil.Phrase("-4"))

	n100 := -100
	p100 := 100
	v := ctx.NewFilterProblem(false).
		AddFilter(filter.New(filter.NewRules(min_2, min_3))("f1", &n100)).
		AddFilter(filter.New(filter.NewRules(max50, max_4))("f2", &p100))
	a.Equal(v.keys, []string{"f1", "f2"}).
		Equal(v.reasons, []string{"-2", "50"})

	n100 = -100
	p100 = 100
	v = ctx.NewFilterProblem(true).
		AddFilter(filter.New(filter.NewRules(min_2, min_3))("f1", &n100)).
		AddFilter(filter.New(filter.NewRules(max50, max_4))("f2", &p100))
	a.Equal(v.keys, []string{"f1"}).
		Equal(v.reasons, []string{"-2"})
}

func TestFilter_When(t *testing.T) {
	a := assert.New(t, false)
	s := newTestServer(a, nil)
	w := httptest.NewRecorder()
	r := rest.Get(a, "/path").Request()
	ctx := s.newContext(w, r, nil)

	min18 := filter.NewRule(min(18), localeutil.Phrase("不能小于 18"))
	notEmpty := filter.NewRule(required[string], localeutil.Phrase("不能为空"))

	obj := &object{}
	v := ctx.NewFilterProblem(false).
		AddFilter(filter.New(min18)("obj/age", &obj.Age)).
		When(obj.Age > 0, func(v *FilterProblem) {
			v.AddFilter(filter.New(notEmpty)("obj/name", &obj.Name))
		})
	a.Equal(v.keys, []string{"obj/age"}).
		Equal(v.reasons, []string{"不能小于 18"})

	obj = &object{Age: 15}
	v = ctx.NewFilterProblem(false).
		AddFilter(filter.New(min18)("obj/age", &obj.Age)).
		When(obj.Age > 0, func(v *FilterProblem) {
			v.AddFilter(filter.New(notEmpty)("obj/name", &obj.Name))
		})
	a.Equal(v.keys, []string{"obj/age", "obj/name"}).
		Equal(v.reasons, []string{"不能小于 18", "不能为空"})
}
