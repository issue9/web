// SPDX-License-Identifier: MIT

package web

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/assert/rest"
	"github.com/issue9/mux/v3"
)

var (
	_ Prefix = &routerPrefix{}
	_ Prefix = &modulePrefix{}

	_ Resource = &routerResource{}
	_ Resource = &moduleResource{}
)

var f1 = func(ctx *Context) { ctx.Render(http.StatusOK, nil, nil) }

func TestRouter(t *testing.T) {
	a := assert.New(t)
	server := newServer(a)
	srv := rest.NewServer(t, server.middlewares, nil)
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

func TestRouter_URL(t *testing.T) {
	a := assert.New(t)

	data := []*struct {
		root      string            // 项目根路径
		input     string            // 输入的内容
		params    map[string]string // 输入路径中带的参数
		url, path string            // 输出内容，分别为 onlyPath 不同值的表现。
	}{
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
		router.Get(item.input, f1)

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
	a.Equal("https://example.com/posts/1", url)
	path, err := router.Path("/posts/1", nil)
	a.Equal("/posts/1", path)

	router.Prefix("/p1").Delete("/path", f1)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodDelete, "https://example.com:88/p1/path", nil)
	srv.middlewares.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, http.StatusOK)

	router.Prefix("/p1").Prefix("/p2").Put("/path", f1)
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodPut, "https://example.com:88/p1/p2/path", nil)
	srv.middlewares.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, http.StatusOK)
}

func TestRouterPrefix(t *testing.T) {
	a := assert.New(t)
	server := newServer(a)
	srv := rest.NewServer(t, server.middlewares, nil)

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

func TestRouterResource(t *testing.T) {
	a := assert.New(t)

	server := newServer(a)

	path := "/path"
	res := server.Router().Resource(path)
	a.NotNil(res)

	srv := rest.NewServer(t, server.middlewares, nil)

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
	r.Static("/admin/{path}", "./testdata")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://example.com/blog/admin/file1.txt", nil)
	server.middlewares.ServeHTTP(w, req)
	a.Equal(w.Result().StatusCode, http.StatusOK)
}

func TestServerFilters(t *testing.T) {
	a := assert.New(t)

	server := newServer(a)
	server.AddFilters(buildFilter("s1"), buildFilter("s2"))
	router := server.Router()
	p1 := router.Prefix("/p1", buildFilter("p11"), buildFilter("p12"))
	p2 := p1.Prefix("/p2", buildFilter("p21"), buildFilter("p22"))
	r1 := router.Resource("/r1", buildFilter("r11"), buildFilter("r12"))
	r2 := p1.Resource("/r2", buildFilter("r21"), buildFilter("r22"))

	server.Router().Get("/test", func(ctx *Context) {
		a.Equal(ctx.Vars["filters"], []string{"s1", "s2"})
		ctx.Render(201, nil, nil)
	})

	p1.Get("/test/202", func(ctx *Context) {
		a.Equal(ctx.Vars["filters"], []string{"s1", "s2", "p11", "p12"}) // 必须要是 server 的先于 prefix 的
		ctx.Render(202, nil, nil)
	})

	p2.Get("/test/202", func(ctx *Context) {
		a.Equal(ctx.Vars["filters"], []string{"s1", "s2", "p11", "p12", "p21", "p22"})
		ctx.Render(202, nil, nil)
	})

	// 以下为动态添加中间件之后的对比方式

	p1.Get("/test/203", func(ctx *Context) {
		a.Equal(ctx.Vars["filters"], []string{"s1", "s2", "s3", "s4", "p11", "p12"})
		ctx.Render(203, nil, nil)
	})

	p2.Get("/test/203", func(ctx *Context) {
		a.Equal(ctx.Vars["filters"], []string{"s1", "s2", "s3", "s4", "p11", "p12", "p21", "p22"})
		ctx.Render(203, nil, nil)
	})

	r1.Get(func(ctx *Context) {
		a.Equal(ctx.Vars["filters"], []string{"s1", "s2", "s3", "s4", "r11", "r12"})
		ctx.Render(204, nil, nil)
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

	srv.Get("/root/p1/p2/test/202").
		Do().
		Status(202)

	// 运行中添加中间件
	server.AddFilters(buildFilter("s3"), buildFilter("s4"))

	srv.Get("/root/p1/test/203").
		Do().
		Status(203)

	srv.Get("/root/p1/p2/test/203").
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

	server := newServer(a)
	m, err := server.NewModule("m1", "m1 desc")
	a.NotError(err).NotNil(m)
	p := m.Prefix("/p")
	a.NotNil(p)
	path := "/path"
	res := p.Resource(path)
	res.Delete(f1)
	res.Get(f1)
	res.Post(f1)
	res.Patch(f1)
	res.Put(f1)
	res.Options("abcdef")

	a.NotError(server.initModules(log.New(ioutil.Discard, "", 0)))

	srv := rest.NewServer(t, server.middlewares, nil)
	srv.Delete("/root/p" + path).Do().Status(http.StatusOK)
	srv.Get("/root/p" + path).Do().Status(http.StatusOK)
	srv.Post("/root/p"+path, nil).Do().Status(http.StatusOK)
	srv.Patch("/root/p"+path, nil).Do().Status(http.StatusOK)
	srv.Put("/root/p"+path, nil).Do().Status(http.StatusOK)
	srv.NewRequest(http.MethodOptions, "/root/p"+path).
		Do().
		Status(http.StatusOK).
		Header("Allow", "abcdef")
}

func TestModulePrefix(t *testing.T) {
	a := assert.New(t)

	server := newServer(a)
	m, err := server.NewModule("m1", "m1 desc")
	a.NotError(err).NotNil(m)
	p := m.Prefix("/p")
	a.NotNil(p)
	path := "/path"
	p.Delete(path, f1)
	p.Get(path, f1)
	p.Post(path, f1)
	p.Patch(path, f1)
	p.Put(path, f1)
	p.Options(path, "abcdef")

	a.NotError(server.initModules(log.New(ioutil.Discard, "", 0)))

	srv := rest.NewServer(t, server.middlewares, nil)
	srv.Delete("/root/p" + path).Do().Status(http.StatusOK)
	srv.Get("/root/p" + path).Do().Status(http.StatusOK)
	srv.Post("/root/p"+path, nil).Do().Status(http.StatusOK)
	srv.Patch("/root/p"+path, nil).Do().Status(http.StatusOK)
	srv.Put("/root/p"+path, nil).Do().Status(http.StatusOK)
	srv.NewRequest(http.MethodOptions, "/root/p"+path).
		Do().
		Status(http.StatusOK).
		Header("Allow", "abcdef")
}

func TestModule_Handle(t *testing.T) {
	a := assert.New(t)

	server := newServer(a)
	m, err := server.NewModule("m1", "m1 desc")
	a.NotError(err).NotNil(m)

	path := "/path"
	a.NotError(m.Handle(path, f1, http.MethodGet, http.MethodDelete))

	a.NotError(server.initModules(log.New(ioutil.Discard, "", 0)))
	srv := rest.NewServer(t, server.middlewares, nil)

	srv.Get("/root" + path).Do().Status(http.StatusOK)
	srv.Delete("/root" + path).Do().Status(http.StatusOK)
	srv.Post("/root"+path, nil).Do().Status(http.StatusMethodNotAllowed)

	// 不指定请求方法，表示所有请求方法

	server = newServer(a)
	m, err = server.NewModule("m1", "m1 desc")
	a.NotError(err).NotNil(m)
	path = "/path1"
	a.NotError(m.Handle(path, f1))

	a.NotError(server.initModules(log.New(ioutil.Discard, "", 0)))
	srv = rest.NewServer(t, server.middlewares, nil)

	srv.Delete("/root" + path).Do().Status(http.StatusOK)
	srv.Patch("/root"+path, nil).Do().Status(http.StatusOK)

	// 各个请求方法

	server = newServer(a)
	m, err = server.NewModule("m1", "m1 desc")
	a.NotError(err).NotNil(m)
	path = "/path2"
	m.Delete(path, f1)
	m.Get(path, f1)
	m.Post(path, f1)
	m.Patch(path, f1)
	m.Put(path, f1)

	a.NotError(server.initModules(log.New(ioutil.Discard, "", 0)))
	srv = rest.NewServer(t, server.middlewares, nil)

	srv.Delete("/root" + path).Do().Status(http.StatusOK)
	srv.Get("/root" + path).Do().Status(http.StatusOK)
	srv.Post("/root"+path, nil).Do().Status(http.StatusOK)
	srv.Patch("/root"+path, nil).Do().Status(http.StatusOK)
	srv.Put("/root"+path, nil).Do().Status(http.StatusOK)
	srv.NewRequest(http.MethodOptions, "/root"+path).
		Do().
		Status(http.StatusOK).
		Header("Allow", "DELETE, GET, HEAD, OPTIONS, PATCH, POST, PUT")
}

func TestModulePrefix_Filters(t *testing.T) {
	a := assert.New(t)

	server := newServer(a)
	m1, err := server.NewModule("m1", "m1 desc")
	a.NotError(err).NotNil(m1)
	m1.AddFilters(buildFilter("m1"), buildFilter("m2"))
	p1 := m1.Prefix("/p1", buildFilter("p1"), buildFilter("p2"))

	m1.Get("/test", func(ctx *Context) {
		a.Equal(ctx.Vars["filters"], []string{"m1", "m2", "s1", "s2"})
		ctx.Render(http.StatusCreated, nil, nil) // 不能输出 200 的状态码
	})

	p1.Get("/test", func(ctx *Context) {
		a.Equal(ctx.Vars["filters"], []string{"m1", "m2", "s1", "s2", "p1", "p2"}) // 必须要是 server 的先于 prefix 的
		ctx.Render(http.StatusAccepted, nil, nil)                                  // 不能输出 200 的状态码
	})

	// 在所有的路由项注册之后才添加中间件
	m1.AddFilters(buildFilter("s1"), buildFilter("s2"))

	a.NotError(server.initModules(log.New(ioutil.Discard, "", 0)))

	srv := rest.NewServer(t, server.middlewares, nil)

	srv.Get("/root/test").
		Do().
		Status(http.StatusCreated) // 验证状态码是否正确

	srv.Get("/root/p1/test").
		Do().
		Status(http.StatusAccepted) // 验证状态码是否正确
}

func TestModule_Options(t *testing.T) {
	a := assert.New(t)

	server := newServer(a)
	m1, err := server.NewModule("m1", "m1 desc")
	a.NotError(err).NotNil(m1)
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

	a.NotError(server.initModules(log.New(ioutil.Discard, "", 0)))
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

	server = newServer(a)
	m1, err = server.NewModule("m1", "m1 desc")
	a.NotError(err).NotNil(m1)
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
	a.NotError(server.initModules(log.New(ioutil.Discard, "", 0)))

	srv = rest.NewServer(t, server.middlewares, nil)
	srv.NewRequest(http.MethodOptions, "/root/test").
		Do().
		Header("Server", "m1").
		Status(http.StatusAccepted)
}
