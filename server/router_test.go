// SPDX-License-Identifier: MIT

package server

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

var f204 = func(ctx *Context) { ctx.Render(http.StatusNoContent, nil, nil) }

func TestRouter(t *testing.T) {
	a := assert.New(t)
	server := newServer(a)
	srv := rest.NewServer(t, server.middlewares, nil)
	router := server.Router()

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

	for i, item := range data {
		u, err := url.Parse(item.root)
		a.NotError(err).NotNil(u)
		router := buildRouter(newServer(a), mux.Default(), u)
		a.NotNil(router)
		router.Get(item.input, f204)

		url, err := router.URL(item.input, item.params)
		a.NotError(err)
		a.Equal(url, item.url, "url not equal @%d,v1=%s,v2=%s", i, url, item.url)
		path, err := router.Path(item.input, item.params)
		a.NotError(err)
		a.Equal(path, item.path, "path not equal @%d,v1=%s,v2=%s", i, path, item.path)
	}

	u, err := url.Parse("https://example.com/blog")
	a.NotError(err).NotNil(u)
	router := buildRouter(newServer(a), mux.Default(), u)
	a.NotNil(router)
	url, err := router.URL("", nil)
	a.NotError(err).Equal(url, "https://example.com/blog")

	p, err := router.Path("", nil)
	a.NotError(err).Equal(p, "/blog")
}

func TestRouter_NewRouter(t *testing.T) {
	a := assert.New(t)
	srv := newServer(a)
	u, err := url.Parse("https://example.com")
	a.NotError(err).NotNil(u)
	router, ok := srv.Router().NewRouter("host", u, mux.NewHosts("example.com"))
	a.True(ok).NotNil(router)

	url, err := router.URL("/posts/1", nil)
	a.NotError(err).Equal("https://example.com/posts/1", url)
	path, err := router.Path("/posts/1", nil)
	a.NotError(err).Equal("/posts/1", path)

	router.Prefix("/p1").Delete("/path", f204)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodDelete, "https://example.com:88/p1/path", nil)
	srv.middlewares.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, http.StatusNoContent)
}

func TestRouterPrefix(t *testing.T) {
	a := assert.New(t)
	server := newServer(a)
	srv := rest.NewServer(t, server.middlewares, nil)

	p := server.Router().Prefix("/p")
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
	res.Get(f204).Delete(f204)
	srv.Get("/root/p" + path).Do().Status(http.StatusNoContent)
	srv.Delete("/root/p" + path).Do().Status(http.StatusNoContent)

	res.Remove(http.MethodDelete)
	srv.Delete("/root/p" + path).Do().Status(http.StatusMethodNotAllowed)
	res.Remove(http.MethodGet)
	srv.Delete("/root/p" + path).Do().Status(http.StatusNotFound)
}

func TestRouterResource(t *testing.T) {
	a := assert.New(t)

	server := newServer(a)

	path := "/path"
	res := server.Router().Resource(path)
	a.NotNil(res)

	srv := rest.NewServer(t, server.middlewares, nil)

	res.Get(f204)
	srv.Get("/root" + path).Do().Status(http.StatusNoContent)

	res.Delete(f204)
	srv.Delete("/root" + path).Do().Status(http.StatusNoContent)

	res.Patch(f204)
	srv.Patch("/root"+path, nil).Do().Status(http.StatusNoContent)

	res.Put(f204)
	srv.Put("/root"+path, nil).Do().Status(http.StatusNoContent)

	res.Post(f204)
	srv.Post("/root"+path, nil).Do().Status(http.StatusNoContent)

	res.Remove(http.MethodPost)
	srv.Post("/root"+path, nil).Do().Status(http.StatusMethodNotAllowed)

	res.Options("def")
	srv.NewRequest(http.MethodOptions, "/root"+path).Do().Header("allow", "def")
}

func TestRouter_Static(t *testing.T) {
	a := assert.New(t)
	server := newServer(a)
	server.SetErrorHandle(func(w http.ResponseWriter, status int) {
		w.WriteHeader(status)
		_, err := w.Write([]byte("error handler test"))
		a.NotError(err)
	}, http.StatusNotFound)

	server.Router().Mux().GetFunc("/m1/test", f201)

	r := server.Router()
	a.Error(r.Static("/path", "./testdata", "index.html"))      // 不包含命名参数
	a.Error(r.Static("/path/{abc", "./testdata", "index.html")) // 格式无效
	a.Error(r.Static("/path/abc}", "./testdata", "index.html")) // 格式无效
	a.Error(r.Static("/path/{}", "./testdata", "index.html"))   // 命名参数未指定名称
	a.Error(r.Static("/path/{}}", "./testdata", "index.html"))  // 格式无效

	r.Static("/client/{path}", "./testdata/", "index.html")
	server.SetErrorHandle(func(w http.ResponseWriter, status int) {
		w.WriteHeader(status)
		_, err := w.Write([]byte("error handler test"))
		a.NotError(err)
	}, http.StatusNotFound)

	srv := rest.NewServer(t, server.middlewares, nil)
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

	// 带域名
	server = newServer(a)
	u, err := url.Parse("https://example.com/blog")
	a.NotError(err).NotNil(u)
	r, ok := server.Router().NewRouter("example", u, mux.NewHosts("example.com"))
	a.True(ok).NotNil(r)
	r.Static("/admin/{path}", "./testdata", "index.html")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://example.com/blog/admin/file1.txt", nil)
	server.middlewares.ServeHTTP(w, req)
	a.Equal(w.Result().StatusCode, http.StatusOK)
}

func TestRouter_AddFilters(t *testing.T) {
	a := assert.New(t)

	server := newServer(a)
	router := server.Router()
	router.AddFilters(buildFilter("s1"), buildFilter("s2"))
	p1 := router.Prefix("/p1", buildFilter("p11"), buildFilter("p12"))
	r1 := router.Resource("/r1", buildFilter("r11"), buildFilter("r12"))
	r2 := p1.Resource("/r2", buildFilter("r21"), buildFilter("r22"))

	server.Router().Get("/test", func(ctx *Context) {
		a.Equal(ctx.Vars["filters"], []string{"s1", "s2"})
		ctx.Render(201, nil, nil)
	})

	p1.Get("/test/202", func(ctx *Context) {
		a.Equal(ctx.Vars["filters"], []string{"s1", "s2", "p11", "p12"}) // 必须要是 router 的先于 prefix 的
		ctx.Render(202, nil, nil)
	})

	// 以下为动态添加中间件之后的对比方式

	p1.Get("/test/203", func(ctx *Context) {
		a.Equal(ctx.Vars["filters"], []string{"s1", "s2", "s3", "s4", "p11", "p12"})
		ctx.Render(203, nil, nil)
	})

	r1.Get(func(ctx *Context) {
		a.Equal(ctx.Vars["filters"], []string{"s1", "s2", "s3", "s4", "r11", "r12"})
		ctx.Render(204, nil, nil) // 检测是否报 http: request method or response status code does not allow body
	})

	r2.Get(func(ctx *Context) {
		a.Equal(ctx.Vars["filters"], []string{"s1", "s2", "s3", "s4", "p11", "p12", "r21", "r22"})
		ctx.Render(205, nil, nil)
	})

	srv := rest.NewServer(t, server.middlewares, nil)

	srv.Get("/root/test").
		Do().
		Status(201)

	srv.Get("/root/p1/test/202").
		Do().
		Status(202)

	// 运行中添加中间件
	router.AddFilters(buildFilter("s3"), buildFilter("s4"))

	srv.Get("/root/p1/test/203").
		Do().
		Status(203)

	srv.Get("/root/r1").
		Do().
		Status(204)

	srv.Get("/root/p1/r2").
		Do().
		Status(205)
}

func TestModuleResource(t *testing.T) {
	a := assert.New(t)

	m := NewModule("m1", "m1 desc")
	a.NotNil(m)
	p := m.Prefix("/p")
	a.NotNil(p)
	path := "/path"
	res := p.Resource(path)
	res.Delete(f204)
	res.Get(f204)
	res.Post(f204)
	res.Patch(f204)
	res.Put(f204)
	res.Options("abcdef")

	server := newServer(a)
	a.NotError(server.AddModule(m))
	a.NotError(server.initModules())

	srv := rest.NewServer(t, server.middlewares, nil)
	srv.Delete("/root/p" + path).Do().Status(http.StatusNoContent)
	srv.Get("/root/p" + path).Do().Status(http.StatusNoContent)
	srv.Post("/root/p"+path, nil).Do().Status(http.StatusNoContent)
	srv.Patch("/root/p"+path, nil).Do().Status(http.StatusNoContent)
	srv.Put("/root/p"+path, nil).Do().Status(http.StatusNoContent)
	srv.NewRequest(http.MethodOptions, "/root/p"+path).
		Do().
		Status(http.StatusOK).
		Header("Allow", "abcdef")
}

func TestModulePrefix(t *testing.T) {
	a := assert.New(t)

	m := NewModule("m1", "m1 desc")
	a.NotNil(m)
	p := m.Prefix("/p")
	a.NotNil(p)
	path := "/path"
	p.Delete(path, f204)
	p.Get(path, f204)
	p.Post(path, f204)
	p.Patch(path, f204)
	p.Put(path, f204)
	p.Options(path, "abcdef")

	server := newServer(a)
	a.NotError(server.AddModule(m))
	a.NotError(server.initModules())

	srv := rest.NewServer(t, server.middlewares, nil)
	srv.Delete("/root/p" + path).Do().Status(http.StatusNoContent)
	srv.Get("/root/p" + path).Do().Status(http.StatusNoContent)
	srv.Post("/root/p"+path, nil).Do().Status(http.StatusNoContent)
	srv.Patch("/root/p"+path, nil).Do().Status(http.StatusNoContent)
	srv.Put("/root/p"+path, nil).Do().Status(http.StatusNoContent)
	srv.NewRequest(http.MethodOptions, "/root/p"+path).
		Do().
		Status(http.StatusOK).
		Header("Allow", "abcdef")
}

func TestModule_Handle(t *testing.T) {
	a := assert.New(t)

	m := NewModule("m1", "m1 desc")
	a.NotNil(m)

	path := "/path"
	a.NotError(m.Handle(path, f204, http.MethodGet, http.MethodDelete))

	server := newServer(a)
	a.NotError(server.AddModule(m))
	a.NotError(server.initModules())
	srv := rest.NewServer(t, server.middlewares, nil)

	srv.Get("/root" + path).Do().Status(http.StatusNoContent)
	srv.Delete("/root" + path).Do().Status(http.StatusNoContent)
	srv.Post("/root"+path, nil).Do().Status(http.StatusMethodNotAllowed)

	// 不指定请求方法，表示所有请求方法

	m = NewModule("m1", "m1 desc")
	a.NotNil(m)
	path = "/path1"
	a.NotError(m.Handle(path, f204))

	server = newServer(a)
	a.NotError(server.AddModule(m))
	a.NotError(server.initModules())
	srv = rest.NewServer(t, server.middlewares, nil)

	srv.Delete("/root" + path).Do().Status(http.StatusNoContent)
	srv.Patch("/root"+path, nil).Do().Status(http.StatusNoContent)

	// 各个请求方法

	m = NewModule("m1", "m1 desc")
	a.NotNil(m)
	path = "/path2"
	m.Delete(path, f204)
	m.Get(path, f204)
	m.Post(path, f204)
	m.Patch(path, f204)
	m.Put(path, f204)

	server = newServer(a)
	a.NotError(server.AddModule(m))
	a.NotError(server.initModules())
	srv = rest.NewServer(t, server.middlewares, nil)

	srv.Delete("/root" + path).Do().Status(http.StatusNoContent)
	srv.Get("/root" + path).Do().Status(http.StatusNoContent)
	srv.Post("/root"+path, nil).Do().Status(http.StatusNoContent)
	srv.Patch("/root"+path, nil).Do().Status(http.StatusNoContent)
	srv.Put("/root"+path, nil).Do().Status(http.StatusNoContent)
	srv.NewRequest(http.MethodOptions, "/root"+path).
		Do().
		Status(http.StatusOK).
		Header("Allow", "DELETE, GET, HEAD, OPTIONS, PATCH, POST, PUT")
}

func TestModule_Options(t *testing.T) {
	a := assert.New(t)

	m1 := NewModule("m1", "m1 desc")
	a.NotNil(m1)
	m1.AddFilters(func(next HandlerFunc) HandlerFunc {
		return HandlerFunc(func(ctx *Context) {
			ctx.Response.Header().Set("Server", "m1")
			next(ctx)
		})
	})

	m1.Get("/test", func(ctx *Context) {
		ctx.Render(http.StatusCreated, nil, nil) // 不能输出 200 的状态码
	})
	m1.Options("/test", "GET, OPTIONS, PUT")

	server := newServer(a)
	a.NotError(server.AddModule(m1))
	a.NotError(server.initModules())
	srv := rest.NewServer(t, server.middlewares, nil)

	srv.Get("/root/test").
		Do().
		Header("Server", "m1").
		Status(http.StatusCreated) // 验证状态码是否正确

	// OPTIONS 不添加中间件
	srv.NewRequest(http.MethodOptions, "/root/test").
		Do().
		Header("Server", "").
		Status(http.StatusOK)

	// 通 Handle 修改的 OPTIONS，正常接受中间件

	m1 = NewModule("m1", "m1 desc")
	a.NotNil(m1)
	m1.AddFilters(func(next HandlerFunc) HandlerFunc {
		return HandlerFunc(func(ctx *Context) {
			ctx.Response.Header().Set("Server", "m1")
			next(ctx)
		})
	})

	m1.Get("/test", func(ctx *Context) {
		ctx.Render(http.StatusCreated, nil, nil)
	})
	m1.Handle("/test", func(ctx *Context) {
		ctx.Render(http.StatusAccepted, nil, nil)
	}, http.MethodOptions)

	server = newServer(a)
	a.NotError(server.AddModule(m1))
	a.NotError(server.initModules())

	srv = rest.NewServer(t, server.middlewares, nil)
	srv.NewRequest(http.MethodOptions, "/root/test").
		Do().
		Header("Server", "m1").
		Status(http.StatusAccepted)
}
