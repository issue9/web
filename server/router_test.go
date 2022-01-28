// SPDX-License-Identifier: MIT

package server_test

import (
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/issue9/assert/v2"
	"github.com/issue9/mux/v6/muxutil"

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

func TestServer_Routers(t *testing.T) {
	a := assert.New(t, false)
	s := servertest.NewTester(a, nil)
	srv := s.Server()
	rs := srv.Routers()

	s.GoServe()

	ver := muxutil.NewHeaderVersion("ver", "v", log.Default(), "2")
	a.NotNil(ver)
	router := rs.New("ver", ver, &server.RouterOptions{
		URLDomain: "https://example.com",
	})
	a.NotNil(router)

	uu, err := router.URL(false, "/posts/1", nil)
	a.NotError(err).Equal("https://example.com/posts/1", uu)

	router.Prefix("/p1").Delete("/path", servertest.BuildHandler(204))
	s.Delete("/p1/path").Header("Accept", "text/plain;v=2").Do(nil).Status(http.StatusNoContent)

	rr := rs.Router("ver")
	a.Equal(rr, router)
	a.Equal(1, len(rs.Routers())).
		Equal(rs.Routers()[0].Name(), "ver")

	// 删除整个路由
	rs.Remove("ver")
	a.Equal(0, len(rs.Routers()))
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
	rs := s.Server().Routers()

	s.GoServe()

	// 带版本

	ver := muxutil.NewHeaderVersion("ver", "vv", log.Default(), "2")
	a.NotNil(ver)
	r := rs.New("ver", ver, &server.RouterOptions{
		URLDomain: "https://example.com/version",
	})
	r.Get("/ver/{path}", s.Server().FileServer(os.DirFS("./testdata"), "path", "index.html"))

	s.Get("/ver/file1.txt").
		Header("Accept", "text/plain;vv=2").
		Do(nil).
		Status(http.StatusOK).
		StringBody("file1")

	p := muxutil.NewPathVersion("vv", "v2")
	a.NotNil(p)
	r = rs.New("path", p, &server.RouterOptions{
		URLDomain: "https://example.com/path",
	})
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
