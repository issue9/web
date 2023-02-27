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
	"golang.org/x/text/language"

	"github.com/issue9/web/logs"
)

var _ BuildProblemFunc = RFC7807Builder

// 此函数放最前，内有依赖行数的测试，心意减少其行数的变化。
func TestContext_Log(t *testing.T) {
	a := assert.New(t, false)
	errLog := new(bytes.Buffer)

	srv := newTestServer(a, &Options{
		Logs: &logs.Options{Writer: logs.NewTextWriter("20060102-15:04:05", errLog), Caller: true, Created: true},
	})
	errLog.Reset()

	t.Run("InternalServerError", func(t *testing.T) {
		a := assert.New(t, false)
		w := httptest.NewRecorder()
		r := rest.Get(a, "/path").Request()
		ctx := srv.newContext(w, r, nil)
		ctx.InternalServerError(errors.New("log1 log2")).Apply(ctx)
		a.Contains(errLog.String(), "problem_test.go:37") // NOTE: 此测试依赖上一行的行号
		a.Contains(errLog.String(), "log1 log2")
		a.Equal(w.Code, 500)
	})

	t.Run("Log", func(t *testing.T) {
		a := assert.New(t, false)
		errLog.Reset()
		w := httptest.NewRecorder()
		r := rest.Get(a, "/path").Request()
		ctx := srv.newContext(w, r, nil)
		ctx.Error("41110", logs.Error, errors.New("log1 log2")).Apply(ctx)
		a.Contains(errLog.String(), "problem_test.go:49") // NOTE: 此测试依赖上一行的行号
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
	resp = ctx.Problem("40000")
	resp.With("with", "abc")
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
	ctx = newTestServer(a, nil).newContext(w, r, nil)
	ctx.Server().AddProblem("40010", http.StatusBadRequest, localeutil.Phrase("40010"), localeutil.Phrase("40010")).
		AddProblem("40011", http.StatusBadRequest, localeutil.Phrase("40011"), localeutil.Phrase("40011"))

	resp = ctx.Problem("40010")
	resp.With("detail", "40010")
	resp.AddParam("k1", "v1")

	resp.Apply(ctx)
	a.Equal(w.Body.String(), `{"type":"40010","title":"40010","status":400,"detail":"40010","params":[{"name":"k1","reason":"v1"}]}`)
}
