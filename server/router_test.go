// SPDX-License-Identifier: MIT

package server_test

import (
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/issue9/assert/v2"
	"github.com/issue9/mux/v6/group"

	"github.com/issue9/web/serialization/text"
	"github.com/issue9/web/server"
	"github.com/issue9/web/server/servertest"
)

func buildMiddleware(a *assert.Assertion, v string) server.MiddlewareFunc {
	return func(next server.HandlerFunc) server.HandlerFunc {
		return func(ctx *server.Context) server.Responser {
			_, err := ctx.Response.Write([]byte(v))
			a.NotError(err)
			return next(ctx)
		}
	}
}

func TestMiddleware(t *testing.T) {
	a := assert.New(t, false)
	srv := servertest.NewTester(a, nil)

	router := srv.NewRouter(buildMiddleware(a, "b1"), buildMiddleware(a, "b2-"))
	a.NotNil(router)
	router.Get("/path", servertest.BuildHandler(201))
	prefix := router.Prefix("/p1", buildMiddleware(a, "p1"), buildMiddleware(a, "p2-"))
	a.NotNil(prefix)
	prefix.Get("/path", servertest.BuildHandler(201))

	srv.GoServe()

	srv.Get("/p1/path").
		Header("accept", text.Mimetype).
		Do(nil).
		Status(http.StatusOK). // 在 WriteHeader 之前有内容输出了
		StringBody("p2-p1b2-b1201")

	srv.Get("/path").
		Header("accept", text.Mimetype).
		Do(nil).
		Status(http.StatusOK). // 在 WriteHeader 之前有内容输出了
		StringBody("b2-b1201")

	srv.Close(0)
	srv.Wait()
}

func TestRouter(t *testing.T) {
	a := assert.New(t, false)
	s := servertest.NewTester(a, nil)
	srv := s.Server()

	s.GoServe()

	ver := group.NewHeaderVersion("ver", "v", log.Default(), "2")
	a.NotNil(ver)
	router := srv.NewRouter("ver", "https://example.com", ver)
	a.NotNil(router)

	uu, err := router.URL(false, "/posts/1", nil)
	a.NotError(err).Equal("https://example.com/posts/1", uu)

	router.Prefix("/p1").Delete("/path", servertest.BuildHandler(204))
	s.Delete("/p1/path").Header("Accept", "text/plain;v=2").Do(nil).Status(http.StatusNoContent)

	rr := srv.Router("ver")
	a.Equal(rr, router)
	a.Equal(1, len(srv.Routers())).
		Equal(srv.Routers()[0].Name(), "ver")

	// 删除整个路由
	srv.RemoveRouter("ver")
	a.Equal(0, len(srv.Routers()))
	s.Delete("/p1/path").
		Header("Accept", "text/plain;v=2").
		Do(nil).
		Status(http.StatusNotFound)

	s.Close(0)
	s.Wait()
}

func TestServer_FileServer(t *testing.T) {
	a := assert.New(t, false)
	s := servertest.NewTester(a, nil)
	s.GoServe()

	// 带版本

	ver := group.NewHeaderVersion("ver", "vv", log.Default(), "2")
	a.NotNil(ver)
	r := s.Server().NewRouter("ver", "https://example.com/version", ver)
	r.Get("/ver/{path}", s.Server().FileServer(os.DirFS("./testdata"), "path", "index.html"))

	s.Get("/ver/file1.txt").
		Header("Accept", "text/plain;vv=2").
		Do(nil).
		Status(http.StatusOK).
		StringBody("file1")

	p := group.NewPathVersion("vv", "v2")
	a.NotNil(p)
	r = s.Server().NewRouter("path", "https://example.com/path", p)
	r.Get("/path/{path}", s.Server().FileServer(os.DirFS("./testdata"), "path", "index.html"))

	s.Get("/v2/path/file1.txt").
		Do(nil).
		Status(http.StatusOK).
		StringBody("file1")

	r = s.NewRouter()
	r.Get("/m1/test", servertest.BuildHandler(201))
	r.Get("/client/{path}", s.Server().FileServer(os.DirFS("./testdata"), "path", "index.html"))

	s.Get("/m1/test").
		Header("accept", text.Mimetype).
		Do(nil).
		Status(http.StatusCreated).
		StringBody("201")

	// 定义的静态文件
	s.Get("/client/file1.txt").
		Do(nil).
		Status(http.StatusOK).
		Header("Content-Type", "text/plain; charset=utf-8").
		StringBody("file1")

	s.Get("/client/not-exists").
		Do(nil).
		Status(http.StatusNotFound)

	// 删除
	r.Remove("/client/{path}")
	s.Get("/client/file1.txt").
		Do(nil).
		Status(http.StatusNotFound)

	s.Close(0)
	s.Wait()
}
