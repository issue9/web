// SPDX-License-Identifier: MIT

package server

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/issue9/assert/v2"
	"github.com/issue9/assert/v2/rest"
	"github.com/issue9/localeutil"
	"github.com/issue9/logs/v4"
	"golang.org/x/text/language"

	"github.com/issue9/web/problem"
	"github.com/issue9/web/serializer/text"
	"github.com/issue9/web/serializer/text/testobject"
)

func TestContext_Error(t *testing.T) {
	a := assert.New(t, false)
	errLog := new(bytes.Buffer)

	srv := newServer(a, &Options{
		Logs: logs.New(logs.NewTextWriter("20060102-15:04:05", errLog), logs.Created, logs.Caller),
	})
	errLog.Reset()

	a.Run("Error", func(a *assert.Assertion) {
		w := httptest.NewRecorder()
		r := rest.Get(a, "/path").Request()
		ctx := srv.newContext(w, r, nil)
		ctx.Error(http.StatusNotImplemented, errors.New("log1 log2")).Apply(ctx)
		a.Contains(errLog.String(), "router_test.go:37") // NOTE: 此测试依赖上一行的行号
		a.Contains(errLog.String(), "log1 log2")
		a.Equal(w.Code, http.StatusNotImplemented)
	})

	a.Run("InternalServerError", func(a *assert.Assertion) {
		errLog.Reset()
		w := httptest.NewRecorder()
		r := rest.Get(a, "/path").Request()
		ctx := srv.newContext(w, r, nil)
		ctx.InternalServerError(errors.New("log1 log2")).Apply(ctx)
		a.Contains(errLog.String(), "router_test.go:48") // NOTE: 此测试依赖上一行的行号
		a.Contains(errLog.String(), "log1 log2")
		a.Equal(w.Code, http.StatusInternalServerError)
	})
}

func TestContext_Result(t *testing.T) {
	a := assert.New(t, false)
	srv := newServer(a, nil)
	a.NotError(srv.CatalogBuilder().SetString(language.Und, "lang", "und"))
	a.NotError(srv.CatalogBuilder().SetString(language.SimplifiedChinese, "lang", "hans"))

	srv.Problems().Add("40000", 400, localeutil.Phrase("lang"), localeutil.Phrase("lang")) // lang 有翻译
	w := httptest.NewRecorder()

	// 能正常翻译错误信息
	r := rest.Get(a, "/path").
		Header("accept-language", language.SimplifiedChinese.String()).
		Header("accept", "application/json").
		Request()
	ctx := srv.newContext(w, r, nil)
	resp := ctx.Problem(nil, "40000")
	resp.Apply(ctx)
	a.Equal(w.Body.String(), `{"type":"40000","title":"hans","detail":"hans","status":400}`)

	// 未指定 accept-language，采用默认的 und
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").
		Header("accept", "application/json").
		Request()
	ctx = srv.newContext(w, r, nil)
	resp = ctx.Problem(nil, "40000")
	resp.Apply(ctx)
	a.Equal(w.Body.String(), `{"type":"40000","title":"und","detail":"und","status":400}`)

	// 不存在的本地化信息，采用默认的 und
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").
		Header("accept-language", "en-US").
		Header("accept", "application/json").
		Request()
	ctx = srv.newContext(w, r, nil)
	resp = ctx.Problem(nil, "40000")
	resp.Apply(ctx)
	a.Equal(w.Body.String(), `{"type":"40000","title":"und","detail":"und","status":400}`)

	// 不存在
	a.Panic(func() { ctx.Problem(nil, "400") })
	a.Panic(func() { ctx.Problem(nil, "50000") })

	// with field

	r = rest.Post(a, "/path", []byte("123")).
		Header("Accept", "application/json").
		Header("Content-Type", "application/json").
		Request()
	w = httptest.NewRecorder()
	ctx = newServer(a, nil).newContext(w, r, nil)
	ctx.Server().Problems().Add("40010", http.StatusBadRequest, localeutil.Phrase("40010"), localeutil.Phrase("40010"))
	ctx.Server().Problems().Add("40011", http.StatusBadRequest, localeutil.Phrase("40011"), localeutil.Phrase("40011"))

	resp = ctx.Problem(problem.NewInvalidParamsProblem(FieldErrs{
		"k1": []string{"v1", "v2"},
	}), "40010")

	resp.Apply(ctx)
	a.Equal(w.Body.String(), `{"type":"40010","title":"40010","detail":"40010","status":400,"invalid-params":[{"name":"k1","reason":"v1"},{"name":"k1","reason":"v2"}]}`)
}

func TestContext_Created(t *testing.T) {
	a := assert.New(t, false)
	s := newServer(a, nil)

	w := httptest.NewRecorder()
	r := rest.Post(a, "/path", nil).
		Header("Accept", text.Mimetype).
		Header("content-type", text.Mimetype).
		Request()
	ctx := s.newContext(w, r, nil)
	resp := ctx.Created(&testobject.TextObject{Name: "test", Age: 123}, "")
	resp.Apply(ctx)
	a.Equal(w.Code, http.StatusCreated).
		Equal(w.Body.String(), `test,123`)

	w = httptest.NewRecorder()
	r = rest.Post(a, "/path", nil).
		Header("Accept", text.Mimetype).
		Header("content-type", text.Mimetype).
		Request()
	ctx = s.newContext(w, r, nil)
	resp = ctx.Created(&testobject.TextObject{Name: "test", Age: 123}, "/test")
	resp.Apply(ctx)
	a.Equal(w.Code, http.StatusCreated).
		Equal(w.Body.String(), `test,123`).
		Equal(w.Header().Get("Location"), "/test")
}

func TestContext_RetryAfter(t *testing.T) {
	a := assert.New(t, false)
	s := newServer(a, nil)

	w := httptest.NewRecorder()
	r := rest.Post(a, "/path", nil).
		Header("Accept", text.Mimetype).
		Header("content-type", text.Mimetype).
		Request()
	ctx := s.newContext(w, r, nil)
	resp := ctx.NotImplemented()
	resp.Apply(ctx)
	a.Equal(w.Code, http.StatusNotImplemented)

	// Retry-After
	w = httptest.NewRecorder()
	r = rest.Post(a, "/path", nil).
		Header("Accept", text.Mimetype).
		Header("content-type", text.Mimetype).
		Request()
	ctx = s.newContext(w, r, nil)
	resp = ctx.RetryAfter(http.StatusServiceUnavailable, 120)
	resp.Apply(ctx)
	a.Equal(w.Code, http.StatusServiceUnavailable).
		Empty(w.Body.String()).
		Equal(w.Header().Get("Retry-After"), "120")

	// Retry-After
	now := time.Now()
	w = httptest.NewRecorder()
	r = rest.Post(a, "/path", nil).
		Header("Accept", text.Mimetype).
		Header("content-type", text.Mimetype).
		Request()
	ctx = s.newContext(w, r, nil)
	resp = ctx.RetryAt(http.StatusMovedPermanently, now)
	resp.Apply(ctx)
	a.Equal(w.Code, http.StatusMovedPermanently).
		Empty(w.Body.String()).
		Contains(w.Header().Get("Retry-After"), "GMT")
}

func TestContext_Redirect(t *testing.T) {
	a := assert.New(t, false)

	r := rest.Post(a, "/path", []byte("123")).
		Header("Accept", "application/json").
		Header("Content-Type", "application/json").
		Request()
	w := httptest.NewRecorder()
	ctx := newServer(a, nil).newContext(w, r, nil)
	resp := ctx.Redirect(301, "https://example.com")
	resp.Apply(ctx)

	a.Equal(w.Result().StatusCode, 301).
		Equal(w.Header().Get("Location"), "https://example.com")
}
