// SPDX-License-Identifier: MIT

package context

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
	srv := rest.NewServer(t, server.Handler(), nil)
	defer srv.Close()

	srv.Get("/debug/pprof/").Do().Status(http.StatusNotFound)
	srv.Get("/debug/vars").Do().Status(http.StatusNotFound)
	server.SetDebugger("/debug/pprof/", "/vars")
	srv.Get("/debug/pprof/").Do().Status(http.StatusOK)
	srv.Get("/vars").Do().Status(http.StatusOK)
}
