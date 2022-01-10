// SPDX-License-Identifier: MIT

package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/issue9/assert/v2"
	"github.com/issue9/assert/v2/rest"
	"github.com/issue9/mux/v6/group"

	"github.com/issue9/web/serialization/text"
)

func buildMiddleware(a *assert.Assertion, v string) MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) Responser {
			_, err := ctx.Response.Write([]byte(v))
			a.NotError(err)
			return next(ctx)
		}
	}
}

func TestMiddleware(t *testing.T) {
	a := assert.New(t, false)
	server := newServer(a, nil)

	router := server.NewRouter("default", "https://localhost:8088/root", group.MatcherFunc(group.Any), buildMiddleware(a, "b1"), buildMiddleware(a, "b2-"))
	a.NotNil(router)
	router.Get("/path", f201)
	prefix := router.Prefix("/p1", buildMiddleware(a, "p1"), buildMiddleware(a, "p2-"))
	a.NotNil(prefix)
	prefix.Get("/path", f201)

	srv := rest.NewServer(a, server.group, nil)
	srv.Get("/p1/path").
		Header("accept", text.Mimetype).
		Do(nil).
		Status(http.StatusOK). // 在 WriteHeader 之前有内容输出了
		StringBody("p2-p1b2-b11234567890")

	srv.Get("/path").
		Header("accept", text.Mimetype).
		Do(nil).
		Status(http.StatusOK). // 在 WriteHeader 之前有内容输出了
		StringBody("b2-b11234567890")
}

func TestRouter(t *testing.T) {
	a := assert.New(t, false)
	srv := newServer(a, nil)
	host := group.NewHosts(false, "example.com")
	a.NotNil(host)

	router := srv.NewRouter("host", "https://example.com", host)
	a.NotNil(router)

	uu, err := router.URL(false, "/posts/1", nil)
	a.NotError(err).Equal("https://example.com/posts/1", uu)

	router.Prefix("/p1").Delete("/path", f204)
	w := httptest.NewRecorder()
	r, err := http.NewRequest(http.MethodDelete, "https://example.com:88/p1/path", nil)
	a.NotError(err).NotNil(r)
	srv.group.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, http.StatusNoContent)

	rr := srv.Router("host")
	a.Equal(rr, router)
	a.Equal(1, len(srv.Routers())).
		Equal(srv.Routers()[0].Name(), "host")

	// 删除整个路由
	srv.RemoveRouter("host")
	a.Equal(0, len(srv.Routers()))
	w = httptest.NewRecorder()
	r, err = http.NewRequest(http.MethodDelete, "https://example.com:88/p1/path", nil)
	a.NotError(err).NotNil(r)
	srv.group.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, http.StatusNotFound)
}

func TestServer_FileServer(t *testing.T) {
	a := assert.New(t, false)
	server := newServer(a, nil)

	r := server.NewRouter("host", "http://localhost:8081/root/", group.MatcherFunc(group.Any))
	a.NotNil(r)
	r.Get("/m1/test", f201)
	r.Get("/client/{path}", server.FileServer(os.DirFS("./testdata"), "path", "index.html"))

	srv := rest.NewServer(a, server.group, nil)

	srv.Get("/m1/test").
		Header("accept", text.Mimetype).
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
	req, err := http.NewRequest(http.MethodGet, "https://example.com/admin/file1.txt", nil)
	a.NotError(err).NotNil(req)
	server.group.ServeHTTP(w, req)
	a.Equal(w.Result().StatusCode, http.StatusOK)
}
