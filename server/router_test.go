// SPDX-License-Identifier: MIT

package server

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/assert/rest"
	"github.com/issue9/localeutil"
	"github.com/issue9/mux/v5/group"
)

var f204 = func(ctx *Context) Responser { return Status(http.StatusNoContent) }

func TestRouter(t *testing.T) {
	a := assert.New(t)
	server := newServer(a, nil)
	srv := rest.NewServer(t, server.group, nil)
	router, err := server.NewRouter("default", "https://localhost:8088/root", group.MatcherFunc(group.Any))
	a.NotError(err).NotNil(router)

	path := "/path"
	a.NotError(router.Handle(path, f204, http.MethodGet, http.MethodDelete))
	srv.Get("/root" + path).Do().Status(http.StatusNoContent)
	srv.Delete("/root" + path).Do().Status(http.StatusNoContent)
	srv.Post("/root"+path, nil).Do().Status(http.StatusMethodNotAllowed)

	// 不指定请求方法，表示所有请求方法
	path = "/path1"
	a.NotError(router.Handle(path, f204))
	srv.Delete("/root" + path).Do().Status(http.StatusNoContent)
	srv.Patch("/root"+path, nil).Do().Status(http.StatusNoContent)

	path = "/path2"

	srv.Delete("/root" + path).Do().Status(http.StatusNotFound)

	router.Delete(path, f204)
	srv.Delete("/root" + path).Do().Status(http.StatusNoContent)

	router.Get(path, f204)
	srv.Get("/root" + path).Do().Status(http.StatusNoContent)

	router.Post(path, f204)
	srv.Post("/root"+path, nil).Do().Status(http.StatusNoContent)

	router.Patch(path, f204)
	srv.Patch("/root"+path, nil).Do().Status(http.StatusNoContent)

	router.Put(path, f204)
	srv.Put("/root"+path, nil).Do().Status(http.StatusNoContent)

	srv.NewRequest(http.MethodOptions, "/root"+path).
		Do().
		Status(http.StatusOK).
		Header("Allow", "DELETE, GET, HEAD, OPTIONS, PATCH, POST, PUT")
}

func TestRouter_SetDebugger(t *testing.T) {
	a := assert.New(t)
	server := newServer(a, nil)
	srv := rest.NewServer(t, server.group, nil)
	defer srv.Close()
	r, err := server.NewRouter("default", "http://localhost:8081/root", group.MatcherFunc(group.Any))
	a.NotError(err).NotNil(r)

	srv.Get("/d/pprof/").Do().Status(http.StatusNotFound)
	srv.Get("/d/vars").Do().Status(http.StatusNotFound)
	a.NotError(r.SetDebugger("/d/pprof/", "/vars"))
	srv.Get("/root/d/pprof/").Do().Status(http.StatusOK) // 相对于 server.Root
	srv.Get("/root/vars").Do().Status(http.StatusOK)
}

func TestRouter_URL(t *testing.T) {
	a := assert.New(t)

	data := []*struct {
		root      string            // 项目根路径
		input     string            // 输入的内容
		params    map[string]string // 输入路径中带的参数
		url, path string            // 输出内容
	}{
		{
			root:  "",
			input: "/abc",
			url:   "/abc",
			path:  "/abc",
		},

		{
			root:  "/",
			input: "/",
			url:   "/",
			path:  "/",
		},

		{
			root:  "/",
			input: "/abc/def",
			url:   "/abc/def",
			path:  "/abc/def",
		},

		{
			root:  "https://localhost/",
			input: "/abc/def",
			url:   "https://localhost/abc/def",
			path:  "/abc/def",
		},
		{
			root:  "https://localhost",
			input: "/abc/def",
			url:   "https://localhost/abc/def",
			path:  "/abc/def",
		},
		{
			root:  "https://localhost",
			input: "abc/def",
			url:   "https://localhost/abc/def",
			path:  "/abc/def",
		},

		{
			root:  "https://example.com:8080/def/",
			input: "abc",
			url:   "https://example.com:8080/def/abc",
			path:  "/def/abc",
		},

		{
			root:   "https://example.com:8080/blog",
			input:  "/posts/{id}/content",
			params: map[string]string{"id": "5"},
			url:    "https://example.com:8080/blog/posts/5/content",
			path:   "/blog/posts/5/content",
		},
	}

	srv := newServer(a, nil)
	for i, item := range data {
		router, err := srv.NewRouter("test-router", item.root, group.MatcherFunc(group.Any))
		a.NotError(err).NotNil(router)
		router.Get(item.input, f204)

		uu, err := router.URL(item.input, item.params)
		a.NotError(err)
		a.Equal(uu, item.url, "url not equal @%d,v1=%s,v2=%s", i, uu, item.url)
		path, err := router.Path(item.input, item.params)
		a.NotError(err)
		a.Equal(path, item.path, "path not equal @%d,v1=%s,v2=%s", i, path, item.path)

		srv.RemoveRouter("test-router")
	}

	r, err := srv.NewRouter("test-router", "https://example.com/blog", group.MatcherFunc(group.Any))
	a.NotError(err).NotNil(r)
	uu, err := r.URL("", nil)
	a.NotError(err).Equal(uu, "https://example.com/blog")

	p, err := r.Path("", nil)
	a.NotError(err).Equal(p, "/blog")
}

func TestRouter_NewRouter(t *testing.T) {
	a := assert.New(t)
	srv := newServer(a, nil)
	host := group.NewHosts(false, "example.com")
	a.NotNil(host)

	router, err := srv.NewRouter("host", "https://example.com", host)
	a.NotError(err).NotNil(router)

	uu, err := router.URL("/posts/1", nil)
	a.NotError(err).Equal("https://example.com/posts/1", uu)
	path, err := router.Path("/posts/1", nil)
	a.NotError(err).Equal("/posts/1", path)

	router.Prefix("/p1").Delete("/path", f204)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodDelete, "https://example.com:88/p1/path", nil)
	srv.group.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, http.StatusNoContent)

	// 删除整个路由
	srv.RemoveRouter("host")
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodDelete, "https://example.com:88/p1/path", nil)
	srv.group.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, http.StatusNotFound)
}

func TestRouter_Prefix(t *testing.T) {
	a := assert.New(t)
	server := newServer(a, nil)
	srv := rest.NewServer(t, server.group, nil)
	router, err := server.NewRouter("host", "http://localhost:8081/root/", group.MatcherFunc(group.Any))
	a.NotError(err).NotNil(router)

	p := router.Prefix("/p")
	a.NotNil(p)

	path := "/path"
	a.NotError(p.Handle(path, f204, http.MethodGet, http.MethodDelete))
	srv.Get("/root/p" + path).Do().Status(http.StatusNoContent)
	srv.Delete("/root/p" + path).Do().Status(http.StatusNoContent)
	srv.Post("/root/p"+path, nil).Do().Status(http.StatusMethodNotAllowed)

	p.Post(path, f204)
	srv.Post("/root/p"+path, nil).Do().Status(http.StatusNoContent)

	p.Patch(path, f204)
	srv.Patch("/root/p"+path, nil).Do().Status(http.StatusNoContent)

	srv.NewRequest(http.MethodOptions, "/root/p"+path).
		Do().
		Status(http.StatusOK).
		Header("allow", "DELETE, GET, HEAD, OPTIONS, PATCH, POST")

	p.Remove(path, http.MethodDelete)
	srv.Delete("/root/p" + path).Do().Status(http.StatusMethodNotAllowed)
}

func TestRouter_Static(t *testing.T) {
	a := assert.New(t)
	server := newServer(a, nil)
	server.SetErrorHandle(func(w io.Writer, status int) {
		_, err := w.Write([]byte("error handler test"))
		a.NotError(err)
	}, http.StatusNotFound)
	r, err := server.NewRouter("host", "http://localhost:8081/root/", group.MatcherFunc(group.Any))
	a.NotError(err).NotNil(r)

	r.Get("/m1/test", f201)
	a.Panic(func() {
		r.Static("/path", "./testdata", "index.html") // 不包含命名参数
	})
	a.Panic(func() {
		r.Static("/path/{abc", "./testdata", "index.html") // 格式无效
	})
	a.Panic(func() {
		r.Static("/path/abc}", "./testdata", "index.html") // 格式无效
	})
	a.Panic(func() {
		r.Static("/path/{}", "./testdata", "index.html") // 命名参数未指定名称
	})
	a.Panic(func() {
		r.Static("/path/{}}", "./testdata", "index.html") // 格式无效
	})

	r.Static("/client/{path}", "./testdata/", "index.html")
	server.SetErrorHandle(func(w io.Writer, status int) {
		_, err := w.Write([]byte("error handler test"))
		a.NotError(err)
	}, http.StatusNotFound)

	srv := rest.NewServer(t, server.group, nil)
	defer srv.Close()

	buf := new(bytes.Buffer)
	srv.Get("/root/m1/test").
		Header("Accept-Encoding", "gzip,deflate;q=0.8").
		Do().
		Status(http.StatusCreated).
		ReadBody(buf).
		Header("Content-Type", "text/html").
		Header("Content-Encoding", "gzip").
		Header("Vary", "Content-Encoding")
	reader, err := gzip.NewReader(buf)
	a.NotError(err).NotNil(reader)
	data, err := ioutil.ReadAll(reader)
	a.NotError(err).NotNil(data)
	a.Equal(string(data), "1234567890")

	// not found
	// 返回 ErrorHandler 内容
	srv.Get("/not-exists.txt").
		Do().
		Status(http.StatusNotFound).
		StringBody("error handler test")

	// 定义的静态文件
	buf.Reset()
	srv.Get("/root/client/file1.txt").
		Header("Accept-Encoding", "gzip,deflate;q=0.8").
		Do().
		Status(http.StatusOK).
		ReadBody(buf).
		Header("Content-Type", "text/plain; charset=utf-8").
		Header("Content-Encoding", "gzip").
		Header("Vary", "Content-Encoding")
	reader, err = gzip.NewReader(buf)
	a.NotError(err).NotNil(reader)
	data, err = ioutil.ReadAll(reader)
	a.NotError(err).NotNil(data)
	a.Equal(string(data), "file1")

	// 删除
	r.Remove("/client/{path}")
	srv.Get("/root/client/file1.txt").
		Do().
		Status(http.StatusNotFound)

	// 带域名
	server = newServer(a, nil)
	host := group.NewHosts(false, "example.com")
	a.NotNil(host)
	r, err = server.NewRouter("example", "https://example.com/blog", host)
	a.NotError(err).NotNil(r)
	r.Static("/admin/{path}", "./testdata", "index.html")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "https://example.com/blog/admin/file1.txt", nil)
	server.group.ServeHTTP(w, req)
	a.Equal(w.Result().StatusCode, http.StatusOK)
}

func TestServer_Router(t *testing.T) {
	a := assert.New(t)
	srv := newServer(a, nil)

	r, err := srv.NewRouter("host", "http://localhost:8081/root/", group.MatcherFunc(group.Any))
	a.NotError(err).NotNil(r)
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

func TestAction_AddRoutes(t *testing.T) {
	a := assert.New(t)

	defRouters := map[string][]string{"*": {http.MethodOptions}}

	srv := newServer(a, nil)
	r, err := srv.NewRouter("host", "http://localhost:8081/root/", group.MatcherFunc(group.Any))
	a.NotError(err).NotNil(r)

	m := srv.NewModule("m1", "v1", localeutil.Phrase("m1 desc"))
	a.NotNil(m)
	m.Action("install").AddRoutes(func(router *Router) {}, "")
	a.Equal(r.MuxRouter().Routes(), defRouters) // 未初始化
	a.Error(srv.initModules(false, "install"))
	a.Equal(r.MuxRouter().Routes(), defRouters) // 已初始化，但是未指定正常的路由名称

	srv = newServer(a, nil)
	r, err = srv.NewRouter("host", "http://localhost:8081/root/", group.MatcherFunc(group.Any))
	a.NotError(err).NotNil(r)
	m = srv.NewModule("m2", "v2", localeutil.Phrase("m2 desc"))
	a.NotNil(m)
	m.Action("install").AddRoutes(func(router *Router) {
		a.Equal(r, router)
		router.Get("p1", f201)
	}, "host")

	a.Equal(r.MuxRouter().Routes(), defRouters) // 未初始化
	a.NotError(srv.initModules(false, "install"))
	a.Equal(2, len(r.MuxRouter().Routes())) // 已初始化，包含一个默认的 OPTIONS *
}
