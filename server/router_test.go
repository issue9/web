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
		a.Contains(errLog.String(), "router_test.go:36") // NOTE: 此测试依赖上一行的行号
		a.Contains(errLog.String(), "log1 log2")
		a.Equal(w.Code, http.StatusNotImplemented)
	})

	a.Run("InternalServerError", func(a *assert.Assertion) {
		errLog.Reset()
		w := httptest.NewRecorder()
		r := rest.Get(a, "/path").Request()
		ctx := srv.newContext(w, r, nil)
		ctx.InternalServerError(errors.New("log1 log2")).Apply(ctx)
		a.Contains(errLog.String(), "router_test.go:47") // NOTE: 此测试依赖上一行的行号
		a.Contains(errLog.String(), "log1 log2")
		a.Equal(w.Code, http.StatusInternalServerError)
	})
}

func TestContext_Result(t *testing.T) {
	a := assert.New(t, false)
	srv := newServer(a, nil)
	a.NotError(srv.CatalogBuilder().SetString(language.Und, "lang", "und"))
	a.NotError(srv.CatalogBuilder().SetString(language.SimplifiedChinese, "lang", "hans"))

	srv.AddErrInfo(400, "40000", localeutil.Phrase("lang")) // lang 有翻译
	w := httptest.NewRecorder()

	// 能正常翻译错误信息
	r := rest.Get(a, "/path").
		Header("accept-language", language.SimplifiedChinese.String()).
		Header("accept", "application/json").
		Request()
	ctx := srv.newContext(w, r, nil)
	resp := ctx.Result("40000", nil)
	resp.Apply(ctx)
	a.Equal(w.Body.String(), `{"message":"hans","code":"40000"}`)

	// 未指定 accept-language，采用默认的 und
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").
		Header("accept", "application/json").
		Request()
	ctx = srv.newContext(w, r, nil)
	resp = ctx.Result("40000", nil)
	resp.Apply(ctx)
	a.Equal(w.Body.String(), `{"message":"und","code":"40000"}`)

	// 不存在的本地化信息，采用默认的 und
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").
		Header("accept-language", "en-US").
		Header("accept", "application/json").
		Request()
	ctx = srv.newContext(w, r, nil)
	resp = ctx.Result("40000", nil)
	resp.Apply(ctx)
	a.Equal(w.Body.String(), `{"message":"und","code":"40000"}`)

	// 不存在
	a.Panic(func() { ctx.Result("400", nil) })
	a.Panic(func() { ctx.Result("50000", nil) })

	// with field

	r = rest.Post(a, "/path", []byte("123")).
		Header("Accept", "application/json").
		Header("Content-Type", "application/json").
		Request()
	w = httptest.NewRecorder()
	ctx = newServer(a, nil).newContext(w, r, nil)
	ctx.Server().AddErrInfo(http.StatusBadRequest, "40010", localeutil.Phrase("40010"))
	ctx.Server().AddErrInfo(http.StatusBadRequest, "40011", localeutil.Phrase("40011"))

	resp = ctx.Result("40010", FieldErrs{
		"k1": []string{"v1", "v2"},
	})

	resp.Apply(ctx)
	a.Equal(w.Body.String(), `{"message":"40010","code":"40010","fields":[{"name":"k1","message":["v1","v2"]}]}`)
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
