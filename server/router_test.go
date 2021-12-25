// SPDX-License-Identifier: MIT

package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert/v2"
	"github.com/issue9/assert/v2/rest"
	"github.com/issue9/mux/v5/group"
)

var f204 = func(ctx *Context) Responser { return Status(http.StatusNoContent) }

func buildFilter(a *assert.Assertion, v string) Filter {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) Responser {
			_, err := ctx.Response.Write([]byte(v))
			a.NotError(err)
			return next(ctx)
		}
	}
}

func TestFilter(t *testing.T) {
	a := assert.New(t, false)
	server := newServer(a, nil)

	router := server.NewRouter("default", "https://localhost:8088/root", group.MatcherFunc(group.Any), buildFilter(a, "b1"), buildFilter(a, "b2-"))
	a.NotNil(router)
	router.Get("/path", f201)
	prefix := router.Prefix("/p1", buildFilter(a, "p1"), buildFilter(a, "p2-"))
	a.NotNil(prefix)
	prefix.Get("/path", f201)

	srv := rest.NewServer(a, server.group, nil)
	srv.Get("/p1/path").
		Do(nil).
		Status(http.StatusOK). // 在 WriteHeader 之前有内容输出了
		StringBody("b1b2-p1p2-1234567890")

	srv.Get("/path").
		Do(nil).
		Status(http.StatusOK). // 在 WriteHeader 之前有内容输出了
		StringBody("b1b2-1234567890")
}

func TestRouter(t *testing.T) {
	a := assert.New(t, false)
	server := newServer(a, nil)
	srv := rest.NewServer(a, server.group, nil)
	router := server.NewRouter("default", "https://localhost:8088/root", group.MatcherFunc(group.Any))
	a.NotNil(router)
	a.Equal(server.Routers(), []*Router{router})

	path := "/path"
	router.Handle(path, f204, http.MethodGet, http.MethodDelete)
	srv.Get(path).Do(nil).Status(http.StatusNoContent)
	srv.Delete(path).Do(nil).Status(http.StatusNoContent)
	srv.Post(path, nil).Do(nil).Status(http.StatusMethodNotAllowed)

	// 不指定请求方法，表示所有请求方法
	path = "/path1"
	router.Handle(path, f204)
	srv.Delete(path).Do(nil).Status(http.StatusNoContent)
	srv.Patch(path, nil).Do(nil).Status(http.StatusNoContent)

	path = "/path2"

	srv.Delete(path).Do(nil).Status(http.StatusNotFound)

	router.Delete(path, f204)
	srv.Delete(path).Do(nil).Status(http.StatusNoContent)

	router.Get(path, f204)
	srv.Get(path).Do(nil).Status(http.StatusNoContent)

	router.Post(path, f204)
	srv.Post(path, nil).Do(nil).Status(http.StatusNoContent)

	router.Patch(path, f204)
	srv.Patch(path, nil).Do(nil).Status(http.StatusNoContent)

	router.Put(path, f204)
	srv.Put(path, nil).Do(nil).Status(http.StatusNoContent)

	srv.NewRequest(http.MethodOptions, path).
		Do(nil).
		Status(http.StatusOK).
		Header("Allow", "DELETE, GET, HEAD, OPTIONS, PATCH, POST, PUT")
}

func TestRouter_URL(t *testing.T) {
	a := assert.New(t, false)

	data := []*struct {
		root   string            // 项目根路径
		input  string            // 输入的内容
		params map[string]string // 输入路径中带的参数
		url    string            // 输出内容
	}{
		{
			root:  "",
			input: "/abc",
			url:   "/abc",
		},

		{
			root:  "/",
			input: "/",
			url:   "/",
		},

		{
			root:  "/",
			input: "/abc/def",
			url:   "/abc/def",
		},

		{
			root:  "https://localhost/",
			input: "/abc/def",
			url:   "https://localhost/abc/def",
		},
		{
			root:  "https://localhost",
			input: "/abc/def",
			url:   "https://localhost/abc/def",
		},
		{
			root:  "https://localhost",
			input: "/abc/def",
			url:   "https://localhost/abc/def",
		},

		{
			root:  "https://example.com:8080/def/",
			input: "/abc",
			url:   "https://example.com:8080/def/abc",
		},

		{
			root:   "https://example.com:8080/blog",
			input:  "/posts/{id}/content",
			params: map[string]string{"id": "5"},
			url:    "https://example.com:8080/blog/posts/5/content",
		},
	}

	srv := newServer(a, nil)
	for i, item := range data {
		router := srv.NewRouter("test-router", item.root, group.MatcherFunc(group.Any))
		a.NotNil(router)
		router.Get(item.input, f204)

		uu, err := router.URL(false, item.input, item.params)
		a.NotError(err)
		a.Equal(uu, item.url, "url not equal @%d,v1=%s,v2=%s", i, uu, item.url)

		srv.RemoveRouter("test-router")
	}

	r := srv.NewRouter("test-router", "https://example.com/blog", group.MatcherFunc(group.Any))
	a.NotNil(r)
	uu, err := r.URL(false, "", nil)
	a.NotError(err).Equal(uu, "https://example.com/blog")
}

func TestRouter_NewRouter(t *testing.T) {
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

	// 删除整个路由
	srv.RemoveRouter("host")
	w = httptest.NewRecorder()
	r, err = http.NewRequest(http.MethodDelete, "https://example.com:88/p1/path", nil)
	a.NotError(err).NotNil(r)
	srv.group.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, http.StatusNotFound)
}

func TestRouter_Prefix(t *testing.T) {
	a := assert.New(t, false)
	server := newServer(a, nil)
	srv := rest.NewServer(a, server.group, nil)
	router := server.NewRouter("host", "http://localhost:8081/root/", group.MatcherFunc(group.Any))
	a.NotNil(router)

	p := router.Prefix("/p")
	a.NotNil(p)

	path := "/path"
	p.Handle(path, f204, http.MethodGet, http.MethodDelete)
	srv.Get("/p" + path).Do(nil).Status(http.StatusNoContent)
	srv.Delete("/p" + path).Do(nil).Status(http.StatusNoContent)
	srv.Post("/p"+path, nil).Do(nil).Status(http.StatusMethodNotAllowed)

	p.Post(path, f204)
	srv.Post("/p"+path, nil).Do(nil).Status(http.StatusNoContent)

	p.Patch(path, f204)
	srv.Patch("/p"+path, nil).Do(nil).Status(http.StatusNoContent)

	srv.NewRequest(http.MethodOptions, "/p"+path).
		Do(nil).
		Status(http.StatusOK).
		Header("allow", "DELETE, GET, HEAD, OPTIONS, PATCH, POST")

	p.Remove(path, http.MethodDelete)
	srv.Delete("/p" + path).Do(nil).Status(http.StatusMethodNotAllowed)
}

func TestServer_Router(t *testing.T) {
	a := assert.New(t, false)
	srv := newServer(a, nil)

	r := srv.NewRouter("host", "http://localhost:8081/root/", group.MatcherFunc(group.Any))
	a.NotNil(r)
	a.Equal(srv.Router("host"), r)

	// 同值，不同类型
	srv.Vars().Store("host", 123)
	a.Equal(srv.Router("host"), r)
	v, found := srv.Vars().Load("host")
	a.True(found).Equal(v, 123)

	srv.RemoveRouter("host")
	a.Nil(srv.Router("host"))
	v, found = srv.Vars().Load("host")
	a.True(found).Equal(v, 123)
}
