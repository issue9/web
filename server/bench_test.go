// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package server

import (
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/issue9/assert/v4"
	"github.com/issue9/mux/v9/header"

	"github.com/issue9/web"
	"github.com/issue9/web/internal/qheader"
	"github.com/issue9/web/server/servertest"
)

func BenchmarkHTTPServer_Serve(b *testing.B) {
	a := assert.New(b, false)
	srv := newTestServer(a, &Options{HTTPServer: &http.Server{Addr: ":8080"}})
	router := srv.Routers().New("srv", nil, web.WithURLDomain("http://localhost:8080/"))
	a.NotNil(router)

	router.Get("/path", func(c *web.Context) web.Responser {
		return web.Response(http.StatusOK, "/path", "h1", "h1")
	})

	defer servertest.Run(a, srv)()
	defer srv.Close(0)
	time.Sleep(500 * time.Millisecond)

	b.Run("charset", func(b *testing.B) {
		a := assert.New(b, false)
		for range b.N {
			r := servertest.Get(a, "http://localhost:8080/path").
				Header(header.ContentType, qheader.BuildContentType(header.JSON, "gbk")).
				Header(header.Accept, header.JSON).
				Header(header.AcceptCharset, "gbk;q=1,gb18080;q=0.1").
				Request()
			resp, err := http.DefaultClient.Do(r)
			a.NotError(err).NotNil(resp).Equal(resp.Header.Get("h1"), "h1")
			body, err := io.ReadAll(resp.Body)
			a.NotError(err).Equal(string(body), `"/path"`)
		}
	})

	b.Run("charset encoding", func(b *testing.B) {
		a := assert.New(b, false)
		for range b.N {
			r := servertest.Get(a, "http://localhost:8080/path").
				Header(header.ContentType, qheader.BuildContentType(header.JSON, "gbk")).
				Header(header.Accept, header.JSON).
				Header(header.AcceptCharset, "gbk;q=1,gb18080;q=0.1").
				Header(header.AcceptEncoding, "gzip").
				Request()
			resp, err := http.DefaultClient.Do(r)
			a.NotError(err).NotNil(resp).Equal(resp.Header.Get("h1"), "h1")
			body, err := io.ReadAll(resp.Body)
			a.NotError(err).NotEqual(body, `"/path"`)
		}
	})

	b.Run("none", func(b *testing.B) {
		a := assert.New(b, false)
		for range b.N {
			r := servertest.Get(a, "http://localhost:8080/path").
				Header(header.ContentType, qheader.BuildContentType(header.JSON, header.UTF8)).
				Header(header.Accept, header.JSON).
				Request()
			resp, err := http.DefaultClient.Do(r)
			a.NotError(err).NotNil(resp).Equal(resp.Header.Get("h1"), "h1")
			body, err := io.ReadAll(resp.Body)
			a.NotError(err).Equal(string(body), `"/path"`)
		}
	})
}
