// SPDX-License-Identifier: MIT

package context

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/assert/rest"
	"github.com/issue9/mux/v3"
)

var f1 = func(ctx *Context) { ctx.Render(http.StatusOK, nil, nil) }

func TestRouter(t *testing.T) {
	a := assert.New(t)
	server := newServer(a)
	srv := rest.NewServer(t, server.Handler(), nil)
	router := server.Router()

	path := "/path"
	a.NotError(router.Handle(path, f1, http.MethodGet, http.MethodDelete))
	srv.Get("/root" + path).Do().Status(http.StatusOK)
	srv.Delete("/root" + path).Do().Status(http.StatusOK)
	srv.Post("/root"+path, nil).Do().Status(http.StatusMethodNotAllowed)

	// 不指定请求方法，表示所有请求方法
	path = "/path1"
	a.NotError(router.Handle(path, f1))
	srv.Delete("/root" + path).Do().Status(http.StatusOK)
	srv.Patch("/root"+path, nil).Do().Status(http.StatusOK)

	path = "/path2"

	srv.Delete("/root" + path).Do().Status(http.StatusNotFound)

	router.Delete(path, f1)
	srv.Delete("/root" + path).Do().Status(http.StatusOK)

	router.Get(path, f1)
	srv.Get("/root" + path).Do().Status(http.StatusOK)

	router.Post(path, f1)
	srv.Post("/root"+path, nil).Do().Status(http.StatusOK)

	router.Patch(path, f1)
	srv.Patch("/root"+path, nil).Do().Status(http.StatusOK)

	router.Put(path, f1)
	srv.Put("/root"+path, nil).Do().Status(http.StatusOK)

	srv.NewRequest(http.MethodOptions, "/root"+path).
		Do().
		Status(http.StatusOK).
		Header("Allow", "DELETE, GET, HEAD, OPTIONS, PATCH, POST, PUT")

	// 自定义 options
	router.Options(path, "abc")
	srv.NewRequest(http.MethodOptions, "/root"+path).
		Do().
		Status(http.StatusOK).
		Header("Allow", "abc")

	router.Remove(path, http.MethodOptions)
	srv.NewRequest(http.MethodOptions, "/root"+path).
		Do().
		Status(http.StatusMethodNotAllowed)
}

func TestRouter_URL_Path(t *testing.T) {
	a := assert.New(t)

	data := []*struct {
		root, input, url, path string
	}{
		{},

		{
			root:  "",
			input: "/abc",
			url:   "/abc",
			path:  "/abc",
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
			root:  "https://localhost/",
			input: "",
			url:   "https://localhost",
			path:  "",
		},

		{
			root:  "https://example.com:8080/def/",
			input: "",
			url:   "https://example.com:8080/def",
			path:  "/def",
		},

		{
			root:  "https://example.com:8080/def/",
			input: "abc",
			url:   "https://example.com:8080/def/abc",
			path:  "/def/abc",
		},
	}

	for i, item := range data {
		u, err := url.Parse(item.root)
		a.NotError(err).NotNil(u)
		router := buildRouter(nil, nil, u)
		a.NotNil(router)

		a.Equal(router.URL(item.input), item.url, "url not equal @%d,v1=%s,v2=%s", i, router.URL(item.input), item.url)
		a.Equal(router.Path(item.input), item.path, "path not equal @%d,v1=%s,v2=%s", i, router.Path(item.input), item.path)
	}
}

func TestRouter_NewRouter(t *testing.T) {
	a := assert.New(t)
	srv := newServer(a)
	u, err := url.Parse("https://example.com")
	a.NotError(err).NotNil(u)
	router, ok := srv.Router().NewRouter("host", u, mux.NewHosts("example.com"))
	a.True(ok).NotNil(router)

	router.Prefix("/p1").Delete("/path", f1)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodDelete, "https://example.com:88/p1/path", nil)
	srv.Handler().ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, http.StatusOK)

	router.Prefix("/p1").Prefix("/p2").Put("/path", f1)
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodPut, "https://example.com:88/p1/p2/path", nil)
	srv.Handler().ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, http.StatusOK)
}

func TestPrefix(t *testing.T) {
	a := assert.New(t)
	server := newServer(a)
	srv := rest.NewServer(t, server.Handler(), nil)

	p := server.Router().Prefix("/p")
	a.NotNil(p)

	path := "/path"
	a.NotError(p.Handle(path, f1, http.MethodGet, http.MethodDelete))
	srv.Get("/root/p" + path).Do().Status(http.StatusOK)
	srv.Delete("/root/p" + path).Do().Status(http.StatusOK)
	srv.Post("/root/p"+path, nil).Do().Status(http.StatusMethodNotAllowed)

	p.Post(path, f1)
	srv.Post("/root/p"+path, nil).Do().Status(http.StatusOK)

	p.Patch(path, f1)
	srv.Patch("/root/p"+path, nil).Do().Status(http.StatusOK)

	p.Options(path, "abc")
	srv.NewRequest(http.MethodOptions, "/root/p"+path).
		Do().
		Status(http.StatusOK).
		Header("allow", "abc")

	p.Remove(path, http.MethodDelete)
	srv.Delete("/root/p" + path).Do().Status(http.StatusMethodNotAllowed)

	// resource

	path = "/resources/{id}"
	res := p.Resource(path)
	res.Get(f1).Delete(f1)
	srv.Get("/root/p" + path).Do().Status(http.StatusOK)
	srv.Delete("/root/p" + path).Do().Status(http.StatusOK)

	res.Remove(http.MethodDelete)
	srv.Delete("/root/p" + path).Do().Status(http.StatusMethodNotAllowed)
	res.Remove(http.MethodGet)
	srv.Delete("/root/p" + path).Do().Status(http.StatusNotFound)
}

func TestResource(t *testing.T) {
	a := assert.New(t)

	server := newServer(a)

	path := "/path"
	res := server.Router().Resource(path)
	a.NotNil(res)

	srv := rest.NewServer(t, server.Handler(), nil)

	res.Get(f1)
	srv.Get("/root" + path).Do().Status(http.StatusOK)

	res.Delete(f1)
	srv.Delete("/root" + path).Do().Status(http.StatusOK)

	res.Patch(f1)
	srv.Patch("/root"+path, nil).Do().Status(http.StatusOK)

	res.Put(f1)
	srv.Put("/root"+path, nil).Do().Status(http.StatusOK)

	res.Post(f1)
	srv.Post("/root"+path, nil).Do().Status(http.StatusOK)

	res.Remove(http.MethodPost)
	srv.Post("/root"+path, nil).Do().Status(http.StatusMethodNotAllowed)

	res.Options("def")
	srv.NewRequest(http.MethodOptions, "/root"+path).Do().Header("allow", "def")
}

func TestServer_Static(t *testing.T) {
	a := assert.New(t)
	server := newServer(a)
	server.SetErrorHandle(func(w http.ResponseWriter, status int) {
		w.WriteHeader(status)
		_, err := w.Write([]byte("error handler test"))
		a.NotError(err)
	}, http.StatusNotFound)

	server.Router().Mux().GetFunc("/m1/test", f201)

	r := server.Router()
	a.Error(r.Static("/path", "./testdata"))      // 不包含命名参数
	a.Error(r.Static("/path/{abc", "./testdata")) // 格式无效
	a.Error(r.Static("/path/abc}", "./testdata")) // 格式无效
	a.Error(r.Static("/path/{}", "./testdata"))   // 命名参数未指定名称
	a.Error(r.Static("/path/{}}", "./testdata"))  // 格式无效

	r.Static("/client/{path}", "./testdata/")
	server.SetErrorHandle(func(w http.ResponseWriter, status int) {
		w.WriteHeader(status)
		_, err := w.Write([]byte("error handler test"))
		a.NotError(err)
	}, http.StatusNotFound)

	srv := rest.NewServer(t, server.Handler(), nil)
	defer srv.Close()

	buf := new(bytes.Buffer)
	srv.Get("/m1/test").
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
}
