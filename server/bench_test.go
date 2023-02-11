// SPDX-License-Identifier: MIT

package server

import (
	"compress/flate"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/issue9/assert/v3"
	"github.com/issue9/assert/v3/rest"
	"github.com/issue9/mux/v7"
	"github.com/issue9/mux/v7/routertest"

	"github.com/issue9/web/internal/header"
	"github.com/issue9/web/internal/testdata"
)

func obj(status int, body any, kv ...string) Responser {
	return ResponserFunc(func(ctx *Context) {
		for i := 0; i < len(kv); i += 2 {
			ctx.Header().Add(kv[i], kv[i+1])
		}
		ctx.Marshal(status, body, false)
	})
}

func BenchmarkRouter(b *testing.B) {
	a := assert.New(b, false)
	srv := newServer(a, &Options{HTTPServer: &http.Server{Addr: ":8080"}})

	h := func(c *Context) Responser {
		_, err := c.Write([]byte(c.Request().URL.Path))
		if err != nil {
			b.Error(err)
		}
		return nil
	}

	routertest.NewTester(srv.call, notFound, buildNodeHandle(http.StatusMethodNotAllowed), buildNodeHandle(http.StatusOK)).Bench(b, h)
}

func BenchmarkServer_Serve(b *testing.B) {
	a := assert.New(b, false)
	srv := newServer(a, &Options{HTTPServer: &http.Server{Addr: ":8080"}})
	router := srv.Routers().New("srv", nil, mux.URLDomain("http://localhost:8080/"))
	a.NotNil(router)

	router.Get("/path", func(c *Context) Responser {
		return obj(http.StatusOK, "/path", "h1", "h1")
	})
	go func() {
		srv.Serve()
	}()
	time.Sleep(500 * time.Millisecond)
	defer srv.Close(0)

	b.Run("charset", func(b *testing.B) {
		a := assert.New(b, false)
		for i := 0; i < b.N; i++ {
			r := rest.Get(a, "http://localhost:8080/path").
				Header("Content-type", header.BuildContentType("application/json", "gbk")).
				Header("accept", "application/json").
				Header("accept-charset", "gbk;q=1,gb18080;q=0.1").
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
			r := rest.Get(a, "http://localhost:8080/path").
				Header("Content-type", header.BuildContentType("application/json", "gbk")).
				Header("accept", "application/json").
				Header("accept-charset", "gbk;q=1,gb18080;q=0.1").
				Header("accept-encoding", "gzip").
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
			r := rest.Get(a, "http://localhost:8080/path").
				Header("Content-type", header.BuildContentType("application/json", header.UTF8Name)).
				Header("accept", "application/json").
				Request()
			resp, err := http.DefaultClient.Do(r)
			a.NotError(err).NotNil(resp)
			a.Equal(resp.Header.Get("h1"), "h1")
			body, err := io.ReadAll(resp.Body)
			a.NotError(err).Equal(string(body), `"/path"`)
		}
	})
}

func BenchmarkServer_newContext(b *testing.B) {
	a := assert.New(b, false)
	srv := newServer(a, nil)

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r := rest.Get(a, "/path").
			Header("Content-type", header.BuildContentType("application/json", "gbk")).
			Header("Accept", "application/json").
			Header("Accept-Charset", "gbk;q=1,gb18080;q=0.1").
			Request()
		ctx := srv.newContext(w, r, nil)
		ctx.destroy()
	}
}

func BenchmarkContext_render(b *testing.B) {
	a := assert.New(b, false)
	srv := newServer(a, nil)

	b.Run("none", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			r := rest.Get(a, "/path").Header("Accept", "application/json").Request()
			ctx := srv.newContext(w, r, nil)

			obj(http.StatusCreated, testdata.ObjectInst).Apply(ctx)
			a.Equal(w.Body.Bytes(), testdata.ObjectJSONString)
		}
	})

	b.Run("utf8", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			r := rest.Get(a, "/path").
				Header("Accept", "application/json").
				Header("Accept-Charset", "utf-8").
				Request()
			ctx := srv.newContext(w, r, nil)

			obj(http.StatusCreated, testdata.ObjectInst).Apply(ctx)
			a.Equal(w.Body.Bytes(), testdata.ObjectJSONString)
		}
	})

	b.Run("gbk", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			r := rest.Get(a, "/path").
				Header("Accept", "application/json").
				Header("Accept-Charset", "gbk;q=1,gb18080;q=0.1").
				Request()
			ctx := srv.newContext(w, r, nil)

			obj(http.StatusCreated, testdata.ObjectInst).Apply(ctx)
			a.Equal(w.Body.Bytes(), testdata.ObjectGBKBytes)
		}
	})

	b.Run("charset; encoding", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			r := rest.Get(a, "/path").
				Header("Accept", "application/json").
				Header("Accept-Charset", "gbk;q=1,gb18080;q=0.1").
				Header("Accept-Encoding", "gzip;q=0.9,deflate").
				Request()

			ctx := srv.newContext(w, r, nil)
			obj(http.StatusCreated, testdata.ObjectInst).Apply(ctx)
			ctx.destroy()

			data, err := io.ReadAll(flate.NewReader(w.Body))
			a.NotError(err).NotNil(data)
			a.Equal(data, testdata.ObjectGBKBytes)
		}
	})
}

func BenchmarkContext_Body(b *testing.B) {
	a := assert.New(b, false)
	srv := newServer(a, nil)

	b.Run("none", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			r := rest.Post(a, "/path", []byte(testdata.ObjectJSONString)).
				Header("Content-type", header.BuildContentType("application/json", "utf-8")).
				Header("Accept", "application/json").
				Request()
			ctx := srv.newContext(w, r, nil)

			body, err := ctx.RequestBody()
			a.NotError(err).Equal(body, []byte(testdata.ObjectJSONString))
		}
	})

	b.Run("charset=gbk", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			r := rest.Post(a, "/path", testdata.ObjectGBKBytes).
				Header("Content-type", header.BuildContentType("application/json", "gbk")).
				Header("Accept", "application/json").
				Request()
			ctx := srv.newContext(w, r, nil)

			body, err := ctx.RequestBody()
			a.NotError(err).Equal(body, []byte(testdata.ObjectJSONString))
		}
	})
}

func BenchmarkContext_Unmarshal(b *testing.B) {
	a := assert.New(b, false)
	srv := newServer(a, nil)

	b.Run("none", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			r := rest.Post(a, "/path", []byte(testdata.ObjectJSONString)).
				Header("Content-type", header.BuildContentType("application/json", "utf-8")).
				Header("Accept", "application/json").
				Request()
			ctx := srv.newContext(w, r, nil)

			obj := &testdata.Object{}
			a.NotError(ctx.Unmarshal(obj)).
				Equal(obj, testdata.ObjectInst)
		}
	})

	b.Run("utf8", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			r := rest.Post(a, "/path", []byte(testdata.ObjectJSONString)).
				Header("Content-type", header.BuildContentType("application/json", "utf-8")).
				Header("Accept", "application/json").
				Request()
			ctx := srv.newContext(w, r, nil)

			obj := &testdata.Object{}
			a.NotError(ctx.Unmarshal(obj)).
				Equal(obj, testdata.ObjectInst)
		}
	})
}

// 一次普通的 POST 请求过程
func BenchmarkPost(b *testing.B) {
	a := assert.New(b, false)
	srv := newServer(a, nil)

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r := rest.Post(a, "/path", []byte(testdata.ObjectJSONString)).
			Header("Content-type", header.BuildContentType("application/json", "utf-8")).
			Header("Accept", "application/json").
			Request()
		ctx := srv.newContext(w, r, nil)

		o := &testdata.Object{}
		a.NotError(ctx.Unmarshal(o)).
			Equal(o, testdata.ObjectInst)

		o.Age++
		o.Name = "response"
		obj(http.StatusCreated, o).Apply(ctx)
		a.Equal(w.Body.String(), `{"name":"response","Age":457}`)
	}
}

func BenchmarkPostWithCharset(b *testing.B) {
	a := assert.New(b, false)
	srv := newServer(a, nil)

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r := rest.Post(a, "/path", testdata.ObjectGBKBytes).
			Header("Content-type", header.BuildContentType("application/json", "gbk")).
			Header("Accept", "application/json").
			Request()
		ctx := srv.newContext(w, r, nil)

		o := &testdata.Object{}
		a.NotError(ctx.Unmarshal(o)).
			Equal(o, testdata.ObjectInst)
	}
}

func BenchmarkContext_Object(b *testing.B) {
	a := assert.New(b, false)
	s := newServer(a, nil)
	o := &testdata.Object{}

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r := rest.Post(a, "/path", nil).
			Header("Accept", "application/json").
			Header("content-type", "application/json").
			Request()
		ctx := s.newContext(w, r, nil)
		obj(http.StatusTeapot, o).Apply(ctx)
	}
}

func BenchmarkContext_Object_withHeader(b *testing.B) {
	a := assert.New(b, false)
	s := newServer(a, nil)
	o := &testdata.Object{}

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r := rest.Post(a, "/path", nil).
			Header("Accept", "application/json").
			Header("content-type", "application/json").
			Request()
		ctx := s.newContext(w, r, nil)
		obj(http.StatusTeapot, o, "Location", "https://example.com").Apply(ctx)
	}
}
