// SPDX-License-Identifier: MIT

package server

import (
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/issue9/assert/v3"
	"github.com/issue9/mux/v7"

	"github.com/issue9/web"
	"github.com/issue9/web/internal/header"
	"github.com/issue9/web/server/servertest"
)

func BenchmarkHTTPServer_Serve(b *testing.B) {
	a := assert.New(b, false)
	srv := newTestServer(a, &Options{HTTPServer: &http.Server{Addr: ":8080"}})
	router := srv.Routers().New("srv", nil, mux.URLDomain("http://localhost:8080/"))
	a.NotNil(router)

	router.Get("/path", func(c *web.Context) web.Responser {
		return web.Response(http.StatusOK, "/path", "h1", "h1")
	})

	defer servertest.Run(a, srv)()
	defer srv.Close(0)
	time.Sleep(500 * time.Millisecond)

	b.Run("charset", func(b *testing.B) {
		a := assert.New(b, false)
		for i := 0; i < b.N; i++ {
			r := servertest.Get(a, "http://localhost:8080/path").
				Header(header.ContentType, header.BuildContentType("application/json", "gbk")).
				Header(header.Accept, "application/json").
				Header(header.AcceptCharset, "gbk;q=1,gb18080;q=0.1").
				Request()
			resp, err := http.DefaultClient.Do(r)
			a.NotError(err).NotNil(resp)
			a.Equal(resp.Header.Get("h1"), "h1")
			body, err := io.ReadAll(resp.Body)
			a.NotError(err).Equal(string(body), `"/path"`)
		}
	})

	b.Run("charset encoding", func(b *testing.B) {
		a := assert.New(b, false)
		for i := 0; i < b.N; i++ {
			r := servertest.Get(a, "http://localhost:8080/path").
				Header(header.ContentType, header.BuildContentType("application/json", "gbk")).
				Header(header.Accept, "application/json").
				Header(header.AcceptCharset, "gbk;q=1,gb18080;q=0.1").
				Header(header.AcceptEncoding, "gzip").
				Request()
			resp, err := http.DefaultClient.Do(r)
			a.NotError(err).NotNil(resp)
			a.Equal(resp.Header.Get("h1"), "h1")
			body, err := io.ReadAll(resp.Body)
			a.NotError(err).NotEqual(body, `"/path"`)
		}
	})

	b.Run("none", func(b *testing.B) {
		a := assert.New(b, false)
		for i := 0; i < b.N; i++ {
			r := servertest.Get(a, "http://localhost:8080/path").
				Header(header.ContentType, header.BuildContentType("application/json", header.UTF8Name)).
				Header(header.Accept, "application/json").
				Request()
			resp, err := http.DefaultClient.Do(r)
			a.NotError(err).NotNil(resp)
			a.Equal(resp.Header.Get("h1"), "h1")
			body, err := io.ReadAll(resp.Body)
			a.NotError(err).Equal(string(body), `"/path"`)
		}
	})
}
