// SPDX-License-Identifier: MIT

package module

import (
	"io/ioutil"
	"log"
	"net/http"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/assert/rest"

	"github.com/issue9/web/context"
)

var (
	f1 = func(ctx *context.Context) {
		ctx.Render(http.StatusOK, nil, nil)
	}
)

func TestResource(t *testing.T) {
	a := assert.New(t)

	server := newServer(a)
	srv := rest.NewServer(t, server.ctxServer.Handler(), nil)
	m := server.NewModule("m1", "m1 desc")
	a.NotNil(m)
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

	a.NotError(server.Init("", log.New(ioutil.Discard, "", 0)))

	srv.Delete("/p" + path).Do().Status(http.StatusOK)
	srv.Get("/p" + path).Do().Status(http.StatusOK)
	srv.Post("/p"+path, nil).Do().Status(http.StatusOK)
	srv.Patch("/p"+path, nil).Do().Status(http.StatusOK)
	srv.Put("/p"+path, nil).Do().Status(http.StatusOK)
	srv.NewRequest(http.MethodOptions, "/p"+path).
		Do().
		Status(http.StatusOK).
		Header("Allow", "abcdef")
}

func TestPrefix(t *testing.T) {
	a := assert.New(t)

	server := newServer(a)
	srv := rest.NewServer(t, server.ctxServer.Handler(), nil)
	m := server.NewModule("m1", "m1 desc")
	a.NotNil(m)
	p := m.Prefix("/p")
	a.NotNil(p)
	path := "/path"
	p.Delete(path, f1)
	p.Get(path, f1)
	p.Post(path, f1)
	p.Patch(path, f1)
	p.Put(path, f1)
	p.Options(path, "abcdef")

	a.NotError(server.Init("", log.New(ioutil.Discard, "", 0)))

	srv.Delete("/p" + path).Do().Status(http.StatusOK)
	srv.Get("/p" + path).Do().Status(http.StatusOK)
	srv.Post("/p"+path, nil).Do().Status(http.StatusOK)
	srv.Patch("/p"+path, nil).Do().Status(http.StatusOK)
	srv.Put("/p"+path, nil).Do().Status(http.StatusOK)
	srv.NewRequest(http.MethodOptions, "/p"+path).
		Do().
		Status(http.StatusOK).
		Header("Allow", "abcdef")
}

func TestModule_Handle(t *testing.T) {
	a := assert.New(t)

	server := newServer(a)
	srv := rest.NewServer(t, server.ctxServer.Handler(), nil)
	m := server.NewModule("m1", "m1 desc")
	a.NotNil(m)

	path := "/path"
	a.NotError(m.Handle(path, f1, http.MethodGet, http.MethodDelete))

	a.NotError(server.Init("", log.New(ioutil.Discard, "", 0)))

	srv.Get(path).Do().Status(http.StatusOK)
	srv.Delete(path).Do().Status(http.StatusOK)
	srv.Post(path, nil).Do().Status(http.StatusMethodNotAllowed)

	// 不指定请求方法，表示所有请求方法

	server = newServer(a)
	srv = rest.NewServer(t, server.ctxServer.Handler(), nil)
	m = server.NewModule("m1", "m1 desc")
	a.NotNil(m)
	path = "/path1"
	a.NotError(m.Handle(path, f1))

	a.NotError(server.Init("", log.New(ioutil.Discard, "", 0)))

	srv.Delete(path).Do().Status(http.StatusOK)
	srv.Patch(path, nil).Do().Status(http.StatusOK)

	// 各个请求方法

	server = newServer(a)
	srv = rest.NewServer(t, server.ctxServer.Handler(), nil)
	m = server.NewModule("m1", "m1 desc")
	a.NotNil(m)
	path = "/path2"
	m.Delete(path, f1)
	m.Get(path, f1)
	m.Post(path, f1)
	m.Patch(path, f1)
	m.Put(path, f1)

	a.NotError(server.Init("", log.New(ioutil.Discard, "", 0)))

	srv.Delete(path).Do().Status(http.StatusOK)
	srv.Get(path).Do().Status(http.StatusOK)
	srv.Post(path, nil).Do().Status(http.StatusOK)
	srv.Patch(path, nil).Do().Status(http.StatusOK)
	srv.Put(path, nil).Do().Status(http.StatusOK)
	srv.NewRequest(http.MethodOptions, path).
		Do().
		Status(http.StatusOK).
		Header("Allow", "DELETE, GET, HEAD, OPTIONS, PATCH, POST, PUT")
}

func buildFilter(num int) context.Filter {
	return context.Filter(func(next context.HandlerFunc) context.HandlerFunc {
		return context.HandlerFunc(func(ctx *context.Context) {
			fs, found := ctx.Vars["filters"]
			if !found {
				ctx.Vars["filters"] = []int{num}
			} else {
				filters := fs.([]int)
				filters = append(filters, num)
				ctx.Vars["filters"] = filters
			}

			next(ctx)
		})
	})
}

func TestPrefix_Filters(t *testing.T) {
	a := assert.New(t)

	server := newServer(a)
	m1 := server.NewModule("m1", "m1 desc")
	m1.AddFilters(buildFilter(1), buildFilter(2))
	p1 := m1.Prefix("/p1", buildFilter(3), buildFilter(4))
	m1.AddFilters(buildFilter(8), buildFilter(9))

	m1.Get("/test", func(ctx *context.Context) {
		a.Equal(ctx.Vars["filters"], []int{1, 2, 8, 9})
		ctx.Render(http.StatusCreated, nil, nil) // 不能输出 200 的状态码
	})

	p1.Get("/test", func(ctx *context.Context) {
		a.Equal(ctx.Vars["filters"], []int{1, 2, 8, 9, 3, 4}) // 必须要是 server 的先于 prefix 的
		ctx.Render(http.StatusAccepted, nil, nil)             // 不能输出 200 的状态码
	})

	a.NotError(server.Init("", log.New(ioutil.Discard, "", 0)))

	srv := rest.NewServer(t, server.ctxServer.Handler(), nil)

	srv.Get("/test").
		Do().
		Status(http.StatusCreated) // 验证状态码是否正确

	srv.Get("/p1/test").
		Do().
		Status(http.StatusAccepted) // 验证状态码是否正确
}

func TestModule_Options(t *testing.T) {
	a := assert.New(t)

	server := newServer(a)
	m1 := server.NewModule("m1", "m1 desc")
	m1.AddFilters(func(next context.HandlerFunc) context.HandlerFunc {
		return context.HandlerFunc(func(ctx *context.Context) {
			ctx.Response.Header().Set("Server", "m1")
			next(ctx)
		})
	})

	m1.Get("/test", func(ctx *context.Context) {
		ctx.Render(http.StatusCreated, nil, nil) // 不能输出 200 的状态码
	})
	m1.Options("/test", "GET, OPTIONS, PUT")
	a.NotError(server.Init("", log.New(ioutil.Discard, "", 0)))

	srv := rest.NewServer(t, server.ctxServer.Handler(), nil)

	srv.Get("/test").
		Do().
		Header("Server", "m1").
		Status(http.StatusCreated) // 验证状态码是否正确

	// OPTIONS 不添加中间件
	srv.NewRequest(http.MethodOptions, "/test").
		Do().
		Header("Server", "").
		Status(http.StatusOK)

	// 通 Handle 修改的 OPTIONS，正常接受中间件

	server = newServer(a)
	m1 = server.NewModule("m1", "m1 desc")
	m1.AddFilters(func(next context.HandlerFunc) context.HandlerFunc {
		return context.HandlerFunc(func(ctx *context.Context) {
			ctx.Response.Header().Set("Server", "m1")
			next(ctx)
		})
	})

	m1.Get("/test", func(ctx *context.Context) {
		ctx.Render(http.StatusCreated, nil, nil)
	})
	m1.Handle("/test", func(ctx *context.Context) {
		ctx.Render(http.StatusAccepted, nil, nil)
	}, http.MethodOptions)
	a.NotError(server.Init("", log.New(ioutil.Discard, "", 0)))

	srv = rest.NewServer(t, server.ctxServer.Handler(), nil)
	srv.NewRequest(http.MethodOptions, "/test").
		Do().
		Header("Server", "m1").
		Status(http.StatusAccepted)
}
