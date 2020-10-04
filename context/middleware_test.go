// SPDX-License-Identifier: MIT

package context

import (
	"net/http"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/assert/rest"
)

func buildFilter(num int) Filter {
	return Filter(func(next HandlerFunc) HandlerFunc {
		return HandlerFunc(func(ctx *Context) {
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
	server.AddFilters(buildFilter(1), buildFilter(2))
	p2 := server.Prefix("/p2", buildFilter(3), buildFilter(4))

	server.Get("/test", func(ctx *Context) {
		a.Equal(ctx.Vars["filters"], []int{1, 2})
		ctx.Render(http.StatusCreated, nil, nil) // 不能输出 200 的状态码
	})

	p2.Get("/test", func(ctx *Context) {
		a.Equal(ctx.Vars["filters"], []int{1, 2, 3, 4}) // 必须要是 server 的先于 prefix 的
		ctx.Render(http.StatusAccepted, nil, nil)       // 不能输出 200 的状态码
	})

	srv := rest.NewServer(t, server.Handler(), nil)

	srv.Get("/test").
		Do().
		Status(http.StatusCreated) // 验证状态码是否正确

	srv.Get("/p2/test").
		Do().
		Status(http.StatusAccepted) // 验证状态码是否正确
}

func TestServer_AddFilters(t *testing.T) {
	a := assert.New(t)
	server := newServer(a)
	server.AddFilters(buildFilter(1), buildFilter(2))
	server.Get("/test", func(ctx *Context) {
		a.Equal(ctx.Vars["filters"], []int{1, 2}) // 查看调用顺序是否正确
		ctx.Render(http.StatusAccepted, nil, nil) // 不能输出 200 的状态码
	})

	rest.NewServer(t, server.Handler(), nil).
		Get("/test").
		Do().
		Status(http.StatusAccepted) // 验证状态码是否正确
}

func TestServer_SetDebugger(t *testing.T) {
	a := assert.New(t)
	server := newServer(a)
	srv := rest.NewServer(t, server.Handler(), nil)
	defer srv.Close()

	srv.Get("/debug/pprof/").Do().Status(http.StatusNotFound)
	srv.Get("/debug/vars").Do().Status(http.StatusNotFound)
	server.SetDebugger("/debug/pprof/", "/vars")
	srv.Get("/debug/pprof/").Do().Status(http.StatusOK)
	srv.Get("/vars").Do().Status(http.StatusOK)
}
