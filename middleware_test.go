// SPDX-License-Identifier: MIT

package web

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/assert/rest"
)

func buildFilter(txt string) Filter {
	return Filter(func(next HandlerFunc) HandlerFunc {
		return HandlerFunc(func(ctx *Context) {
			fs, found := ctx.Vars["filters"]
			if !found {
				ctx.Vars["filters"] = []string{txt}
			} else {
				filters := fs.([]string)
				filters = append(filters, txt)
				ctx.Vars["filters"] = filters
			}

			next(ctx)
		})
	})
}

func TestServer_SetRecovery(t *testing.T) {
	a := assert.New(t)

	server := newServer(a)
	server.Router().Get("/panic", func(ctx *Context) { panic("panic") })
	srv := rest.NewServer(t, server.middlewares, nil)
	defer srv.Close()

	// 默认值，返回 500
	srv.Get("/root/panic").Do().Status(http.StatusInternalServerError)

	// 自定义 Recovery
	server.SetRecovery(func(w http.ResponseWriter, msg interface{}) {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte(fmt.Sprint(msg)))
	})
	srv.Get("/root/panic").Do().Status(http.StatusBadGateway).StringBody("panic")
}

func TestServer_SetDebugger(t *testing.T) {
	a := assert.New(t)
	server := newServer(a)
	srv := rest.NewServer(t, server.middlewares, nil)
	defer srv.Close()

	srv.Get("/d/pprof/").Do().Status(http.StatusNotFound)
	srv.Get("/d/vars").Do().Status(http.StatusNotFound)
	server.SetDebugger("/d/pprof/", "/vars")
	srv.Get("/root/d/pprof/").Do().Status(http.StatusOK) // 相对于 server.Root
	srv.Get("/root/vars").Do().Status(http.StatusOK)
}

func TestServer_Compress(t *testing.T) {
	a := assert.New(t)
	server := newServer(a)
	srv := rest.NewServer(t, server.middlewares, nil)
	defer srv.Close()

	server.Router().Static("/client/{path}", "./testdata/")
	srv.Get("/root/client/file1.txt").
		Header("Accept-Encoding", "gzip,deflate;q=0.8").
		Do().
		Status(http.StatusOK).
		Header("Content-Type", "text/plain; charset=utf-8").
		Header("Content-Encoding", "gzip").
		Header("Vary", "Content-Encoding")

	// 删除 gzip
	server.SetCompressAlgorithm("gzip", nil)
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
