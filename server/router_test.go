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

	"github.com/issue9/assert/v2"
	"github.com/issue9/assert/v2/rest"
	"github.com/issue9/mux/v5/group"
)

var f204 = func(ctx *Context) Responser { return Status(http.StatusNoContent) }

func TestRouter(t *testing.T) {
	a := assert.New(t, false)
	server := newServer(a, nil)
	srv := rest.NewServer(a, server.group, nil)
	router := server.NewRouter("default", "https://localhost:8088/root", group.MatcherFunc(group.Any))
	a.NotNil(router)
	a.Equal(server.Routers(), []*Router{router})

	path := "/path"
	router.Handle(path, f204, http.MethodGet, http.MethodDelete)
	srv.Get(path).Do().Status(http.StatusNoContent)
	srv.Delete(path).Do().Status(http.StatusNoContent)
	srv.Post(path, nil).Do().Status(http.StatusMethodNotAllowed)

	// 不指定请求方法，表示所有请求方法
	path = "/path1"
	router.Handle(path, f204)
	srv.Delete(path).Do().Status(http.StatusNoContent)
	srv.Patch(path, nil).Do().Status(http.StatusNoContent)

	path = "/path2"

	srv.Delete(path).Do().Status(http.StatusNotFound)

	router.Delete(path, f204)
	srv.Delete(path).Do().Status(http.StatusNoContent)

	router.Get(path, f204)
	srv.Get(path).Do().Status(http.StatusNoContent)

	router.Post(path, f204)
	srv.Post(path, nil).Do().Status(http.StatusNoContent)

	router.Patch(path, f204)
	srv.Patch(path, nil).Do().Status(http.StatusNoContent)

	router.Put(path, f204)
	srv.Put(path, nil).Do().Status(http.StatusNoContent)

	srv.NewRequest(http.MethodOptions, path).
		Do().
		Status(http.StatusOK).
		Header("Allow", "DELETE, GET, HEAD, OPTIONS, PATCH, POST, PUT")
}

func TestRouter_SetDebugger(t *testing.T) {
	a := assert.New(t, false)
	server := newServer(a, nil)
	srv := rest.NewServer(a, server.group, nil)
	r := server.NewRouter("default", "http://localhost:8081", group.MatcherFunc(group.Any))
	a.NotNil(r)

	srv.Get("/d/pprof/").Do().Status(http.StatusNotFound)
	srv.Get("/d/vars").Do().Status(http.StatusNotFound)
	r.SetDebugger("/d/pprof/", "/vars")
	srv.Get("/d/pprof/").Do().Status(http.StatusOK) // 相对于 server.Root
	srv.Get("/vars").Do().Status(http.StatusOK)
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
	a := assert.New(t, false)
	server := newServer(a, nil)
	srv := rest.NewServer(a, server.group, nil)
	router := server.NewRouter("host", "http://localhost:8081/root/", group.MatcherFunc(group.Any))
	a.NotNil(router)

	p := router.Prefix("/p")
	a.NotNil(p)

	path := "/path"
	p.Handle(path, f204, http.MethodGet, http.MethodDelete)
	srv.Get("/p" + path).Do().Status(http.StatusNoContent)
	srv.Delete("/p" + path).Do().Status(http.StatusNoContent)
	srv.Post("/p"+path, nil).Do().Status(http.StatusMethodNotAllowed)

	p.Post(path, f204)
	srv.Post("/p"+path, nil).Do().Status(http.StatusNoContent)

	p.Patch(path, f204)
	srv.Patch("/p"+path, nil).Do().Status(http.StatusNoContent)

	srv.NewRequest(http.MethodOptions, "/p"+path).
		Do().
		Status(http.StatusOK).
		Header("allow", "DELETE, GET, HEAD, OPTIONS, PATCH, POST")

	p.Remove(path, http.MethodDelete)
	srv.Delete("/p" + path).Do().Status(http.StatusMethodNotAllowed)
}

func TestRouter_Static(t *testing.T) {
	a := assert.New(t, false)
	server := newServer(a, nil)
	server.SetErrorHandle(func(w io.Writer, status int) {
		_, err := w.Write([]byte("error handler test"))
		a.NotError(err)
	}, http.StatusNotFound)
	r := server.NewRouter("host", "http://localhost:8081/root/", group.MatcherFunc(group.Any))
	a.NotNil(r)

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

	srv := rest.NewServer(a, server.group, nil)

	srv.Get("/m1/test").
		Header("Accept-Encoding", "gzip,deflate;q=0.8").
		Do().
		Status(http.StatusCreated).
		Header("Content-Type", "text/html").
		Header("Content-Encoding", "gzip").
		Header("Vary", "Content-Encoding").
		BodyFunc(func(a *assert.Assertion, body []byte) {
			buf := bytes.NewBuffer(body)
			reader, err := gzip.NewReader(buf)
			a.NotError(err).NotNil(reader)
			data, err := ioutil.ReadAll(reader)
			a.NotError(err).NotNil(data)
			a.Equal(string(data), "1234567890")
		})

	// not found
	// 返回 ErrorHandler 内容
	srv.Get("/not-exists.txt").
		Do().
		Status(http.StatusNotFound).
		StringBody("error handler test")

	// 定义的静态文件
	srv.Get("/client/file1.txt").
		Header("Accept-Encoding", "gzip,deflate;q=0.8").
		Do().
		Status(http.StatusOK).
		Header("Content-Type", "text/plain; charset=utf-8").
		Header("Content-Encoding", "gzip").
		Header("Vary", "Content-Encoding").
		BodyFunc(func(a *assert.Assertion, body []byte) {
			buf := bytes.NewBuffer(body)
			reader, err := gzip.NewReader(buf)
			a.NotError(err).NotNil(reader)
			data, err := ioutil.ReadAll(reader)
			a.NotError(err).NotNil(data)
			a.Equal(string(data), "file1")
		})

	// 删除
	r.Remove("/client/{path}")
	srv.Get("/client/file1.txt").
		Do().
		Status(http.StatusNotFound)

	// 带域名
	server = newServer(a, nil)
	host := group.NewHosts(false, "example.com")
	a.NotNil(host)
	r = server.NewRouter("example", "https://example.com/blog", host)
	a.NotNil(r)
	r.Static("/admin/{path}", "./testdata", "index.html")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "https://example.com/admin/file1.txt", nil)
	server.group.ServeHTTP(w, req)
	a.Equal(w.Result().StatusCode, http.StatusOK)
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
