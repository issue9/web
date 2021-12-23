// SPDX-License-Identifier: MIT

package server

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/issue9/assert/v2"
	"github.com/issue9/assert/v2/rest"
	"github.com/issue9/localeutil"
	"github.com/issue9/mux/v5/group"
	"golang.org/x/text/language"
)

var (
	errLog      = new(bytes.Buffer)
	criticalLog = new(bytes.Buffer)
)

func TestServer_FileServer(t *testing.T) {
	a := assert.New(t, false)
	server := newServer(a, nil)

	r := server.NewRouter("host", "http://localhost:8081/root/", group.MatcherFunc(group.Any))
	a.NotNil(r)
	r.Get("/m1/test", f201)
	r.Get("/client/{path}", server.FileServer(os.DirFS("./testdata"), "path", "index.html"))

	srv := rest.NewServer(a, server.group, nil)

	srv.Get("/m1/test").
		Do(nil).
		Status(http.StatusCreated).
		Header("Content-Type", "text/html").
		StringBody("1234567890")

	// 定义的静态文件
	srv.Get("/client/file1.txt").
		Do(nil).
		Status(http.StatusOK).
		Header("Content-Type", "text/plain; charset=utf-8").
		StringBody("file1")

	srv.Get("/client/not-exists").
		Do(nil).
		Status(http.StatusNotFound)

	// 删除
	r.Remove("/client/{path}")
	srv.Get("/client/file1.txt").
		Do(nil).
		Status(http.StatusNotFound)

	// 带域名
	server = newServer(a, nil)
	host := group.NewHosts(false, "example.com")
	a.NotNil(host)
	r = server.NewRouter("example", "https://example.com/blog", host)
	a.NotNil(r)
	r.Get("/admin/{path}", server.FileServer(os.DirFS("./testdata"), "path", "index.html"))
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "https://example.com/admin/file1.txt", nil)
	server.group.ServeHTTP(w, req)
	a.Equal(w.Result().StatusCode, http.StatusOK)
}

func TestContext_Critical(t *testing.T) {
	a := assert.New(t, false)
	w := httptest.NewRecorder()
	criticalLog.Reset()
	b := newServer(a, nil)
	b.Logs().CRITICAL().SetOutput(criticalLog)
	b.Logs().CRITICAL().SetFlags(log.Llongfile)
	ctx := &Context{
		Response: w,
		server:   b,
	}

	ctx.renderResponser(ctx.Critical(http.StatusInternalServerError, "log1", "log2"))
	a.Contains(criticalLog.String(), "response_test.go:85") // NOTE: 此测试依赖上一行的行号
	a.Contains(criticalLog.String(), "log1 log2")
}

func TestContext_Errorf(t *testing.T) {
	a := assert.New(t, false)
	w := httptest.NewRecorder()
	b := newServer(a, nil)
	errLog.Reset()
	b.Logs().ERROR().SetOutput(errLog)
	ctx := &Context{
		Response: w,
		server:   b,
	}

	ctx.renderResponser(ctx.Errorf(http.StatusInternalServerError, "error @%s:%d", "file.go", 51))
	a.True(strings.HasPrefix(errLog.String(), "error @file.go:51"))
}

func TestContext_Criticalf(t *testing.T) {
	a := assert.New(t, false)
	w := httptest.NewRecorder()
	b := newServer(a, nil)
	criticalLog.Reset()
	b.Logs().CRITICAL().SetOutput(criticalLog)
	ctx := &Context{
		Response: w,
		server:   b,
	}

	ctx.renderResponser(ctx.Criticalf(http.StatusInternalServerError, "error @%s:%d", "file.go", 51))
	a.True(strings.HasPrefix(criticalLog.String(), "error @file.go:51"))
}

func TestContext_ResultWithFields(t *testing.T) {
	a := assert.New(t, false)

	r := httptest.NewRequest(http.MethodGet, "/path", bytes.NewBufferString("123"))
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
	r := httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Set("accept-language", language.SimplifiedChinese.String())
	r.Header.Set("accept", "application/json")
	ctx := srv.NewContext(w, r)
	resp := ctx.Result("40000", nil)
	ctx.renderResponser(resp)
	a.Equal(w.Body.String(), `{"message":"hans","code":"40000"}`)

	// 未指定 accept-language，采用默认的 und
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Set("accept", "application/json")
	ctx = srv.NewContext(w, r)
	resp = ctx.Result("40000", nil)
	ctx.renderResponser(resp)
	a.Equal(w.Body.String(), `{"message":"und","code":"40000"}`)

	// 不存在的本地化信息，采用默认的 und
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/path", nil)
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

	r := httptest.NewRequest(http.MethodGet, "/path", bytes.NewBufferString("123"))
	r.Header.Set("Accept", "application/json")
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx := newServer(a, nil).NewContext(w, r)
	a.Equal(ctx.Redirect(301, "https://example.com"), exited)
	a.Equal(w.Result().StatusCode, 301).
		Equal(w.Header().Get("Location"), "https://example.com")

}
