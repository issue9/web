// SPDX-License-Identifier: MIT

package server

import (
	"net/http"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/assert/rest"
	"github.com/issue9/mux/v5/group"
)

func buildFilter(txt string) Filter {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context) {
			fs, found := ctx.Vars["filters"]
			if !found {
				ctx.Vars["filters"] = []string{txt}
			} else {
				filters := fs.([]string)
				filters = append(filters, txt)
				ctx.Vars["filters"] = filters
			}

			next(ctx)
		}
	}
}

func TestServer_Compress(t *testing.T) {
	a := assert.New(t)
	server := newServer(a)
	srv := rest.NewServer(t, server.groups, nil)
	defer srv.Close()
	router, err := server.NewRouter("default", "http://localhost:8081/root", group.MatcherFunc(group.Any))
	a.NotError(err).NotNil(router)

	a.NotError(router.Static("/client/{path}", "./testdata/", "index.html"))
	srv.Get("/root/client/file1.txt").
		Header("Accept-Encoding", "gzip,deflate;q=0.8").
		Do().
		Status(http.StatusOK).
		Header("Content-Type", "text/plain; charset=utf-8").
		Header("Content-Encoding", "gzip").
		Header("Vary", "Content-Encoding")

	// 删除 gzip
	a.NotError(server.SetCompressAlgorithm("gzip", nil))
	srv.Get("/root/client/file1.txt").
		Header("Accept-Encoding", "gzip,deflate;q=0.8").
		Do().
		Status(http.StatusOK).
		Header("Content-Type", "text/plain; charset=utf-8").
		Header("Content-Encoding", "deflate").
		Header("Vary", "Content-Encoding")

	// 禁用所有的
	server.DeleteCompressTypes("*")
	srv.Get("/root/client/file1.txt").
		Header("Accept-Encoding", "gzip,deflate;q=0.8").
		Do().
		Status(http.StatusOK).
		Header("Content-Type", "text/plain; charset=utf-8").
		Header("Content-Encoding", "")
}

func testServer_AddFilters(t *testing.T) {
	a := assert.New(t)

	server := newServer(a)
	router, err := server.NewRouter("example", "http://localhost:8081/root", group.MatcherFunc(group.Any))
	a.NotError(err).NotNil(router)
	//server.AddFilters(buildFilter("s1"), buildFilter("s2"))
	p1 := router.Prefix("/p1")

	router.Get("/test", func(ctx *Context) {
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

	srv := rest.NewServer(t, server.groups, nil)

	srv.Get("/root/test").
		Do().
		Status(201)

	srv.Get("/root/p1/test/202").
		Do().
		Status(202)

	// 运行中添加中间件
	//server.AddFilters(buildFilter("s3"), buildFilter("s4"))

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
