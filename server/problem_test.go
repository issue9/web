// SPDX-License-Identifier: MIT

package server

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/assert/v3/rest"
	"github.com/issue9/localeutil"
	"github.com/issue9/logs/v4"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// 此函数放最前，内有依赖行数的测试，心意减少其行数的变化。
func TestContext_Log(t *testing.T) {
	a := assert.New(t, false)
	errLog := new(bytes.Buffer)

	srv := newServer(a, &Options{
		Logs: logs.New(logs.NewTextWriter("20060102-15:04:05", errLog), logs.Created, logs.Caller),
	})
	errLog.Reset()

	t.Run("InternalServerError", func(t *testing.T) {
		a := assert.New(t, false)
		w := httptest.NewRecorder()
		r := rest.Get(a, "/path").Request()
		ctx := srv.newContext(w, r, nil)
		ctx.InternalServerError(errors.New("log1 log2")).Apply(ctx)
		a.Contains(errLog.String(), "problem_test.go:35") // NOTE: 此测试依赖上一行的行号
		a.Contains(errLog.String(), "log1 log2")
		a.Equal(w.Code, 500)
	})

	t.Run("Log", func(t *testing.T) {
		a := assert.New(t, false)
		errLog.Reset()
		w := httptest.NewRecorder()
		r := rest.Get(a, "/path").Request()
		ctx := srv.newContext(w, r, nil)
		ctx.Error("41110", logs.LevelError, errors.New("log1 log2")).Apply(ctx)
		a.Contains(errLog.String(), "problem_test.go:47") // NOTE: 此测试依赖上一行的行号
		a.Contains(errLog.String(), "log1 log2")
		a.Equal(w.Code, 411)
	})
}

func TestProblems_Add(t *testing.T) {
	a := assert.New(t, false)

	ps := newProblems(RFC7807Builder)
	a.NotNil(ps)
	a.NotZero(ps.Count())
	l := ps.Count()

	ps.Add("40010", 400, localeutil.Phrase("title"), localeutil.Phrase("detail"))
	a.Equal(l+1, ps.Count()).True(ps.Exists("40010"))

	ps.Add("40011", 400, localeutil.Phrase("title"), localeutil.Phrase("detail"))
	a.Equal(l+2, ps.Count())

	a.PanicString(func() {
		ps.Add("40010", 400, localeutil.Phrase("title"), localeutil.Phrase("detail"))
	}, "存在相同值的 id 参数")
	a.Equal(l+2, ps.Count())
}

func TestProblems_Visit(t *testing.T) {
	a := assert.New(t, false)

	ps := newProblems(RFC7807Builder)
	cnt := 0
	ps.Visit(func(id string, status int, title, detail localeutil.LocaleStringer) bool {
		cnt++
		return true
	})
	a.Equal(ps.Count(), cnt)
	l := ps.Count()

	ps.Add("40010", 400, localeutil.Phrase("title"), localeutil.Phrase("detail"))
	cnt = 0
	ps.Visit(func(id string, status int, title, detail localeutil.LocaleStringer) bool {
		cnt++
		return true
	})
	a.Equal(l+1, cnt)

	ps.Add("40011", 400, localeutil.Phrase("title"), localeutil.Phrase("detail"))
	ps.Add("40012", 400, localeutil.Phrase("title"), localeutil.Phrase("detail"))
	cnt = 0
	ps.Visit(func(id string, status int, title, detail localeutil.LocaleStringer) bool {
		cnt++
		return true
	})
	a.Equal(l+3, cnt)

	cnt = 0
	ps.Visit(func(id string, status int, title, detail localeutil.LocaleStringer) bool {
		cnt++
		return false // 中断
	})
	a.Equal(1, cnt)
}

func TestProblems_Mimetype(t *testing.T) {
	a := assert.New(t, false)
	ps := newProblems(RFC7807Builder)
	a.NotNil(ps)

	a.Equal(ps.mimetype("application/json"), "application/json")
	ps.AddMimetype("application/json", "application/problem+json")
	a.Equal(ps.mimetype("application/json"), "application/problem+json")
	a.PanicString(func() {
		ps.AddMimetype("application/json", "application/problem")
	}, "已经存在的 mimetype")
}

func TestProblems_Problem(t *testing.T) {
	a := assert.New(t, false)
	s := newServer(a, nil)
	ps := s.Problems()

	a.PanicString(func() {
		ps.Problem(message.NewPrinter(language.Und), "not-exists")
	}, "未找到有关 not-exists 的定义")

	w := httptest.NewRecorder()
	r := rest.Get(a, "/path").
		Header("Accept", "application/json").
		Header("accept-language", language.SimplifiedChinese.String()).
		Request()
	ctx := s.newContext(w, r, nil)

	p := ps.Problem(ctx.LocalePrinter(), "41110")
	a.NotNil(p)
	p.Apply(ctx)
	a.Equal(w.Body.String(), `{"type":"41110","title":"hans","status":411}`).
		Equal(w.Result().StatusCode, 411)
}

func TestContext_Problem(t *testing.T) {
	a := assert.New(t, false)
	srv := newServer(a, nil)
	a.NotError(srv.CatalogBuilder().SetString(language.Und, "lang", "und"))
	a.NotError(srv.CatalogBuilder().SetString(language.SimplifiedChinese, "lang", "hans"))
	srv.Problems().Add("40000", 400, localeutil.Phrase("lang"), localeutil.Phrase("lang")) // lang 有翻译

	// 能正常翻译错误信息
	w := httptest.NewRecorder()
	r := rest.Get(a, "/path").
		Header("accept-language", language.SimplifiedChinese.String()).
		Header("accept", "application/json").
		Request()
	ctx := srv.newContext(w, r, nil)
	resp := ctx.Problem("40000")
	resp.Apply(ctx)
	a.Equal(w.Body.String(), `{"type":"40000","title":"hans","status":400}`)

	// 未指定 accept-language，采用默认的 und
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").
		Header("accept", "application/json").
		Request()
	ctx = srv.newContext(w, r, nil)
	resp = ctx.Problem("40000")
	resp.Apply(ctx)
	a.Equal(w.Body.String(), `{"type":"40000","title":"und","status":400}`)

	// 不存在的本地化信息，采用默认的 und
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").
		Header("accept-language", "en-US").
		Header("accept", "application/json").
		Request()
	ctx = srv.newContext(w, r, nil)
	resp = ctx.Problem("40000").With("with", "abc")
	resp.Apply(ctx)
	a.Equal(w.Body.String(), `{"type":"40000","title":"und","status":400,"with":"abc"}`)

	// 不存在
	a.Panic(func() { ctx.Problem("not-exists") })
	a.Panic(func() { ctx.Problem("50000") })

	// with field

	r = rest.Post(a, "/path", []byte("123")).
		Header("Accept", "application/json").
		Header("Content-Type", "application/json").
		Request()
	w = httptest.NewRecorder()
	ctx = newServer(a, nil).newContext(w, r, nil)
	ctx.Server().Problems().Add("40010", http.StatusBadRequest, localeutil.Phrase("40010"), localeutil.Phrase("40010"))
	ctx.Server().Problems().Add("40011", http.StatusBadRequest, localeutil.Phrase("40011"), localeutil.Phrase("40011"))

	resp = ctx.Problem("40010").With("detail", "40010")
	resp.AddParam("k1", "v1")

	resp.Apply(ctx)
	a.Equal(w.Body.String(), `{"type":"40010","title":"40010","status":400,"detail":"40010","params":[{"name":"k1","reason":"v1"}]}`)
}
