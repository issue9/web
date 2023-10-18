// SPDX-License-Identifier: MIT

package server

import (
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/issue9/assert/v3"
	"github.com/issue9/mux/v7"
	"github.com/issue9/mux/v7/group"

	"github.com/issue9/web"
	"github.com/issue9/web/servertest"
)

func buildMiddleware(a *assert.Assertion, v string) web.Middleware {
	return web.MiddlewareFunc(func(next web.HandlerFunc) web.HandlerFunc {
		return func(ctx *web.Context) web.Responser {
			h := ctx.Header()
			val := h.Get("h")
			h.Set("h", v+val)

			resp := next(ctx)
			a.NotNil(resp)
			return resp
		}
	})
}

func TestServer_Routers(t *testing.T) {
	a := assert.New(t, false)
	srv := newTestServer(a, nil)

	defer servertest.Run(a, srv)()
	defer srv.Close(500 * time.Millisecond)

	ver := group.NewHeaderVersion("ver", "v", log.Default(), "2")
	a.NotNil(ver)
	r1 := srv.NewRouter("ver", ver, mux.URLDomain("https://example.com"))
	a.NotNil(r1)

	uu, err := r1.URL(false, "/posts/1", nil)
	a.NotError(err).Equal("https://example.com/posts/1", uu)

	r1.Prefix("/p1").Delete("/path", buildHandler(http.StatusCreated))
	servertest.Delete(a, "http://localhost:8080/p1/path").
		Header("Accept", "application/json;v=2").
		Do(nil).
		Status(http.StatusCreated)
	servertest.NewRequest(a, http.MethodOptions, "http://localhost:8080/p1/path").
		Header("Accept", "application/json;v=2").
		Do(nil).Status(http.StatusOK)

	r2 := srv.GetRouter("ver")
	a.Equal(r2, r1)
	a.Equal(1, len(srv.Routers())).
		Equal(srv.Routers()[0].Name(), "ver")

	// 删除整个路由
	srv.RemoveRouter("ver")
	a.Equal(0, len(srv.Routers()))
	servertest.Delete(a, "http://localhost:8080/p1/path").
		Header("Accept", "application/json;v=2").
		Do(nil).
		Status(http.StatusNotFound)
}



func TestMiddleware(t *testing.T) {
	a := assert.New(t, false)
	srv := newTestServer(a, nil)
	count := 0

	router := srv.NewRouter("def", nil)
	router.Use(buildMiddleware(a, "b1"), buildMiddleware(a, "b2-"), web.MiddlewareFunc(func(next web.HandlerFunc) web.HandlerFunc {
		return func(ctx *web.Context) web.Responser {
			ctx.OnExit(func(*web.Context, int) {
				count++
			})
			return next(ctx)
		}
	}))
	a.NotNil(router)
	router.Get("/path", buildHandler(201))
	prefix := router.Prefix("/p1", buildMiddleware(a, "p1"), buildMiddleware(a, "p2-"))
	a.NotNil(prefix)
	prefix.Get("/path", buildHandler(201))

	defer servertest.Run(a, srv)()
	defer srv.Close(500 * time.Millisecond)

	servertest.Get(a, "http://localhost:8080/p1/path").
		Header("accept", "application/json").
		Do(nil).
		Status(http.StatusCreated).
		Header("h", "p1p2-b1b2-").
		StringBody("201")
	a.Equal(count, 1)

	servertest.Get(a, "http://localhost:8080/path").
		Header("accept", "application/json").
		Do(nil).
		Status(http.StatusCreated).
		Header("h", "b1b2-").
		StringBody("201")
	a.Equal(count, 2)
}
