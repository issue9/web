// SPDX-License-Identifier: MIT

package web

import (
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
