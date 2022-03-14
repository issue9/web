// SPDX-License-Identifier: MIT

package server_test

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
	"github.com/issue9/mux/v6/muxutil"
	"golang.org/x/text/language"

	"github.com/issue9/web/serialization/text"
	"github.com/issue9/web/server"
	"github.com/issue9/web/server/servertest"
)

func buildMiddleware(a *assert.Assertion, v string) server.Middleware {
	return server.MiddlewareFunc(func(next server.HandlerFunc) server.HandlerFunc {
		return func(ctx *server.Context) *server.Response {
			resp := next(ctx)
			a.NotNil(resp)

			val, _ := resp.GetHeader("h")
			resp = resp.SetHeader("h", val+v)
			a.NotNil(resp)

			return resp
		}
	})
}

func TestMiddleware(t *testing.T) {
	a := assert.New(t, false)
	srv := servertest.NewTester(a, nil)

	router := srv.NewRouter(buildMiddleware(a, "b1"), buildMiddleware(a, "b2-"))
	a.NotNil(router)
	router.Get("/path", servertest.BuildHandler(201))
	prefix := router.Prefix("/p1", buildMiddleware(a, "p1"), buildMiddleware(a, "p2-"))
	a.NotNil(prefix)
	prefix.Get("/path", servertest.BuildHandler(201))

	srv.GoServe()

	srv.Get("/p1/path").
		Header("accept", text.Mimetype).
		Do(nil).
		Status(http.StatusCreated).
		Header("h", "b1b2-p1p2-").
		StringBody("201")

	srv.Get("/path").
		Header("accept", text.Mimetype).
		Do(nil).
		Status(http.StatusCreated).
		Header("h", "b1b2-").
		StringBody("201")

	srv.Close(0)
	srv.Wait()
}

func TestServer_Routers(t *testing.T) {
	a := assert.New(t, false)
	s := servertest.NewTester(a, nil)
	srv := s.Server()
	rs := srv.Routers()

	s.GoServe()

	ver := muxutil.NewHeaderVersion("ver", "v", log.Default(), "2")
	a.NotNil(ver)
	r1 := rs.New("ver", ver, &server.RouterOptions{
		URLDomain: "https://example.com",
	})
	a.NotNil(r1)

	uu, err := r1.URL(false, "/posts/1", nil)
	a.NotError(err).Equal("https://example.com/posts/1", uu)

	r1.Prefix("/p1").Delete("/path", servertest.BuildHandler(http.StatusCreated))
	s.Delete("/p1/path").Header("Accept", "text/plain;v=2").Do(nil).Status(http.StatusCreated)

	r2 := rs.Router("ver")
	a.Equal(r2, r1)
	a.Equal(1, len(rs.Routers())).
		Equal(rs.Routers()[0].Name(), "ver")

	// 删除整个路由
	rs.Remove("ver")
	a.Equal(0, len(rs.Routers()))
	s.Delete("/p1/path").
		Header("Accept", "text/plain;v=2").
		Do(nil).
		Status(http.StatusNotFound)

	s.Close(0)
	s.Wait()
}

func TestServer_FileServer(t *testing.T) {
	a := assert.New(t, false)
	s := servertest.NewTester(a, nil)
	rs := s.Server().Routers()

	s.GoServe()

	// 带版本

	ver := muxutil.NewHeaderVersion("ver", "vv", log.Default(), "2")
	a.NotNil(ver)
	r := rs.New("ver", ver, &server.RouterOptions{
		URLDomain: "https://example.com/version",
	})
	r.Get("/ver/{path}", s.Server().FileServer(os.DirFS("./testdata"), "path", "index.html"))

	s.Get("/ver/file1.txt").
		Header("Accept", "text/plain;vv=2").
		Do(nil).
		Status(http.StatusOK).
		StringBody("file1")

	p := muxutil.NewPathVersion("vv", "v2")
	a.NotNil(p)
	r = rs.New("path", p, &server.RouterOptions{
		URLDomain: "https://example.com/path",
	})
	r.Get("/path/{path}", s.Server().FileServer(os.DirFS("./testdata"), "path", "index.html"))

	s.Get("/v2/path/file1.txt").
		Do(nil).
		Status(http.StatusOK).
		StringBody("file1")

	r = s.NewRouter()
	r.Get("/m1/test", servertest.BuildHandler(201))
	r.Get("/client/{path}", s.Server().FileServer(os.DirFS("./testdata"), "path", "index.html"))

	s.Get("/m1/test").
		Header("accept", text.Mimetype).
		Do(nil).
		Status(http.StatusCreated).
		StringBody("201")

	// 定义的静态文件
	s.Get("/client/file1.txt").
		Do(nil).
		Status(http.StatusOK).
		Header("Content-Type", "text/plain; charset=utf-8").
		StringBody("file1")

	s.Get("/client/not-exists").
		Do(nil).
		Status(http.StatusNotFound)

	// 删除
	r.Remove("/client/{path}")
	s.Get("/client/file1.txt").
		Do(nil).
		Status(http.StatusNotFound)

	s.Close(0)
	s.Wait()
}

func TestContext_Error(t *testing.T) {
	a := assert.New(t, false)
	errLog := new(bytes.Buffer)

	srv := servertest.NewServer(a, nil)
	errLog.Reset()
	srv.Logs().ERROR().SetOutput(errLog)
	srv.Logs().ERROR().SetFlags(log.Llongfile)

	a.Run("Error", func(a *assert.Assertion) {
		w := httptest.NewRecorder()
		r := rest.Get(a, "/path").Request()
		ctx := srv.NewContext(w, r)
		ctx.Render(ctx.Error(http.StatusNotImplemented, "log1", "log2"))
		a.Contains(errLog.String(), "router_test.go:187") // NOTE: 此测试依赖上一行的行号
		a.Contains(errLog.String(), "log1 log2")
		a.Equal(w.Code, http.StatusNotImplemented)
	})

	a.Run("InternalServerError", func(a *assert.Assertion) {
		errLog.Reset()
		w := httptest.NewRecorder()
		r := rest.Get(a, "/path").Request()
		ctx := srv.NewContext(w, r)
		ctx.Render(ctx.InternalServerError("log1", "log2"))
		a.Contains(errLog.String(), "router_test.go:198") // NOTE: 此测试依赖上一行的行号
		a.Contains(errLog.String(), "log1 log2")
		a.Equal(w.Code, http.StatusInternalServerError)
	})

	srv.Logs().ERROR().SetFlags(0)

	a.Run("Errorf", func(a *assert.Assertion) {
		errLog.Reset()
		w := httptest.NewRecorder()
		r := rest.Get(a, "/path").Request()
		ctx := srv.NewContext(w, r)
		ctx.Render(ctx.Errorf(http.StatusNotImplemented, "error @%s:%d", "file.go", 51))
		a.True(strings.HasPrefix(errLog.String(), "error @file.go:51"))
		a.Equal(w.Code, http.StatusNotImplemented)
	})

	a.Run("InternalServerErrorf", func(a *assert.Assertion) {
		errLog.Reset()
		w := httptest.NewRecorder()
		r := rest.Get(a, "/path").Request()
		ctx := srv.NewContext(w, r)
		ctx.Render(ctx.InternalServerErrorf("error @%s:%d", "file.go", 51))
		a.True(strings.HasPrefix(errLog.String(), "error @file.go:51"))
		a.Equal(w.Code, http.StatusInternalServerError)
	})
}

func TestResp(t *testing.T) {
	a := assert.New(t, false)

	a.Panic(func() {
		server.Resp(50)
	})

	a.Panic(func() {
		server.Resp(600)
	})

	s := server.Resp(201)
	a.NotNil(s)

	s.SetHeader("h", "1")
	v, found := s.GetHeader("h")
	a.True(found).Equal(v, "1")
	s.SetHeader("h", "2")
	v, found = s.GetHeader("h")
	a.True(found).Equal(v, "2")
	s.DelHeader("h")
	v, found = s.GetHeader("h")
	a.False(found).Equal(v, "")

	srv := servertest.NewServer(a, nil)

	r := rest.Get(a, "/path").Request()
	w := httptest.NewRecorder()
	ctx := srv.NewContext(w, r)
	ctx.Render(s)
	a.Equal(w.Code, 201)

	r = rest.Get(a, "/path").Request()
	w = httptest.NewRecorder()
	ctx = srv.NewContext(w, r)
	ctx.Render(nil)
	a.Equal(w.Code, 200) // 默认值 200
}

func TestContext_ResultWithFields(t *testing.T) {
	a := assert.New(t, false)

	r := rest.Post(a, "/path", []byte("123")).
		Header("Accept", "application/json").
		Header("Content-Type", "application/json").
		Request()
	w := httptest.NewRecorder()
	ctx := servertest.NewServer(a, nil).NewContext(w, r)
	ctx.Server().AddResult(http.StatusBadRequest, "40010", localeutil.Phrase("40010"))
	ctx.Server().AddResult(http.StatusBadRequest, "40011", localeutil.Phrase("40011"))

	resp := ctx.Result("40010", server.ResultFields{
		"k1": []string{"v1", "v2"},
	})

	ctx.Render(resp)
	a.Equal(w.Body.String(), `{"message":"40010","code":"40010","fields":[{"name":"k1","message":["v1","v2"]}]}`)
}

func TestContext_Result(t *testing.T) {
	a := assert.New(t, false)
	srv := servertest.NewServer(a, nil)
	a.NotError(srv.Locale().Builder().SetString(language.Und, "lang", "und"))
	a.NotError(srv.Locale().Builder().SetString(language.SimplifiedChinese, "lang", "hans"))

	srv.AddResult(400, "40000", localeutil.Phrase("lang")) // lang 有翻译
	w := httptest.NewRecorder()

	// 能正常翻译错误信息
	r := rest.Get(a, "/path").
		Header("accept-language", language.SimplifiedChinese.String()).
		Header("accept", "application/json").
		Request()
	ctx := srv.NewContext(w, r)
	resp := ctx.Result("40000", nil)
	ctx.Render(resp)
	a.Equal(w.Body.String(), `{"message":"hans","code":"40000"}`)

	// 未指定 accept-language，采用默认的 und
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").
		Header("accept", "application/json").
		Request()
	ctx = srv.NewContext(w, r)
	resp = ctx.Result("40000", nil)
	ctx.Render(resp)
	a.Equal(w.Body.String(), `{"message":"und","code":"40000"}`)

	// 不存在的本地化信息，采用默认的 und
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").
		Header("accept-language", "en-US").
		Header("accept", "application/json").
		Request()
	ctx = srv.NewContext(w, r)
	resp = ctx.Result("40000", nil)
	ctx.Render(resp)
	a.Equal(w.Body.String(), `{"message":"und","code":"40000"}`)

	// 不存在
	a.Panic(func() { ctx.Result("400", nil) })
	a.Panic(func() { ctx.Result("50000", nil) })
}

func TestContext_Redirect(t *testing.T) {
	a := assert.New(t, false)

	r := rest.Post(a, "/path", []byte("123")).
		Header("Accept", "application/json").
		Header("Content-Type", "application/json").
		Request()
	w := httptest.NewRecorder()
	ctx := servertest.NewServer(a, nil).NewContext(w, r)
	ctx.Render(ctx.Redirect(301, "https://example.com"))

	a.Equal(w.Result().StatusCode, 301).
		Equal(w.Header().Get("Location"), "https://example.com")
}
