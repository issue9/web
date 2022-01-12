// SPDX-License-Identifier: MIT

package server

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/issue9/assert/v2"
	"github.com/issue9/assert/v2/rest"
	"github.com/issue9/localeutil"
	"golang.org/x/text/language"
)

var (
	errLog      = new(bytes.Buffer)
	criticalLog = new(bytes.Buffer)
)

func TestContext_Critical(t *testing.T) {
	a := assert.New(t, false)
	srv := newServer(a, nil)
	criticalLog.Reset()
	srv.Logs().CRITICAL().SetOutput(criticalLog)
	srv.Logs().CRITICAL().SetFlags(log.Llongfile)

	w := httptest.NewRecorder()
	r := rest.Get(a, "/path").Request()
	ctx := srv.NewContext(w, r)

	ctx.renderResponser(ctx.Critical(http.StatusInternalServerError, "log1", "log2"))
	a.Contains(criticalLog.String(), "response_test.go:35") // NOTE: 此测试依赖上一行的行号
	a.Contains(criticalLog.String(), "log1 log2")
}

func TestContext_Errorf(t *testing.T) {
	a := assert.New(t, false)
	srv := newServer(a, nil)
	errLog.Reset()
	srv.Logs().ERROR().SetOutput(errLog)

	w := httptest.NewRecorder()
	r := rest.Get(a, "/path").Request()
	ctx := srv.NewContext(w, r)

	ctx.renderResponser(ctx.Errorf(http.StatusInternalServerError, "error @%s:%d", "file.go", 51))
	a.True(strings.HasPrefix(errLog.String(), "error @file.go:51"))
}

func TestContext_Criticalf(t *testing.T) {
	a := assert.New(t, false)
	srv := newServer(a, nil)
	criticalLog.Reset()
	srv.Logs().CRITICAL().SetOutput(criticalLog)

	w := httptest.NewRecorder()
	r := rest.Get(a, "/path").Request()
	ctx := srv.NewContext(w, r)

	ctx.renderResponser(ctx.Criticalf(http.StatusInternalServerError, "error @%s:%d", "file.go", 51))
	a.True(strings.HasPrefix(criticalLog.String(), "error @file.go:51"))
}

func TestStatus(t *testing.T) {
	a := assert.New(t, false)

	a.Panic(func() {
		Status(50)
	})

	a.Panic(func() {
		Status(600)
	})

	s := Status(201)
	a.NotNil(s)

	srv := newServer(a, nil)

	r, err := http.NewRequest(http.MethodGet, "/path", nil)
	a.NotError(err).NotNil(r)
	w := httptest.NewRecorder()
	ctx := srv.NewContext(w, r)
	ctx.renderResponser(s)
	a.Equal(w.Code, 201)

	r, err = http.NewRequest(http.MethodGet, "/path", nil)
	a.NotError(err).NotNil(r)
	w = httptest.NewRecorder()
	ctx = srv.NewContext(w, r)
	ctx.renderResponser(Exit())
	a.Equal(w.Code, 200) // 默认值 200
}

func TestContext_ResultWithFields(t *testing.T) {
	a := assert.New(t, false)

	r, err := http.NewRequest(http.MethodGet, "/path", bytes.NewBufferString("123"))
	a.NotError(err).NotNil(r)
	r.Header.Set("Accept", "application/json")
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx := newServer(a, nil).NewContext(w, r)
	ctx.server.AddResult(http.StatusBadRequest, "40010", localeutil.Phrase("40010"))
	ctx.server.AddResult(http.StatusBadRequest, "40011", localeutil.Phrase("40011"))

	resp := ctx.Result("40010", ResultFields{
		"k1": []string{"v1", "v2"},
	})

	ctx.renderResponser(resp)
	a.Equal(w.Body.String(), `{"message":"40010","code":"40010","fields":[{"name":"k1","message":["v1","v2"]}]}`)
}

func TestContext_Result(t *testing.T) {
	a := assert.New(t, false)
	srv := newServer(a, nil)
	a.NotError(srv.Locale().Builder().SetString(language.Und, "lang", "und"))
	a.NotError(srv.Locale().Builder().SetString(language.SimplifiedChinese, "lang", "hans"))

	srv.AddResult(400, "40000", localeutil.Phrase("lang")) // lang 有翻译
	w := httptest.NewRecorder()

	// 能正常翻译错误信息
	r, err := http.NewRequest(http.MethodGet, "/path", nil)
	a.NotError(err).NotNil(r)
	r.Header.Set("accept-language", language.SimplifiedChinese.String())
	r.Header.Set("accept", "application/json")
	ctx := srv.NewContext(w, r)
	resp := ctx.Result("40000", nil)
	ctx.renderResponser(resp)
	a.Equal(w.Body.String(), `{"message":"hans","code":"40000"}`)

	// 未指定 accept-language，采用默认的 und
	w = httptest.NewRecorder()
	r, err = http.NewRequest(http.MethodGet, "/path", nil)
	a.NotError(err).NotNil(r)
	r.Header.Set("accept", "application/json")
	ctx = srv.NewContext(w, r)
	resp = ctx.Result("40000", nil)
	ctx.renderResponser(resp)
	a.Equal(w.Body.String(), `{"message":"und","code":"40000"}`)

	// 不存在的本地化信息，采用默认的 und
	w = httptest.NewRecorder()
	r, err = http.NewRequest(http.MethodGet, "/path", nil)
	a.NotError(err).NotNil(r)
	r.Header.Set("accept-language", "en-US")
	r.Header.Set("accept", "application/json")
	ctx = srv.NewContext(w, r)
	resp = ctx.Result("40000", nil)
	ctx.renderResponser(resp)
	a.Equal(w.Body.String(), `{"message":"und","code":"40000"}`)

	// 不存在
	a.Panic(func() { ctx.Result("400", nil) })
	a.Panic(func() { ctx.Result("50000", nil) })
}

func TestContext_Redirect(t *testing.T) {
	a := assert.New(t, false)

	r, err := http.NewRequest(http.MethodGet, "/path", bytes.NewBufferString("123"))
	a.NotError(err).NotNil(r)
	r.Header.Set("Accept", "application/json")
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx := newServer(a, nil).NewContext(w, r)
	a.Equal(ctx.Redirect(301, "https://example.com"), exited)
	a.Equal(w.Result().StatusCode, 301).
		Equal(w.Header().Get("Location"), "https://example.com")

}
