// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package web

import (
	"bytes"
	"compress/flate"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert/v6"
	"github.com/issue9/assert/v6/rest"
	"github.com/issue9/mux/v10"
	"github.com/issue9/mux/v10/header"
	"github.com/issue9/mux/v10/routertest"
	"github.com/issue9/mux/v10/types"

	"github.com/issue9/web/internal/qheader"
)

func BenchmarkRouter(b *testing.B) {
	a := assert.New(b, false)

	h := func(c *Context) Responser {
		_, err := c.Write([]byte(c.Request().URL.Path))
		if err != nil {
			b.Error(err)
		}
		return nil
	}

	tt := routertest.NewTester(func(o ...mux.Option) *routertest.TestRouter[HandlerFunc] {
		s := newTestServer(a, false)
		r := s.Routers().g.New("main", nil, o...)
		return &routertest.TestRouter[HandlerFunc]{
			ServeHTTP: s.routers.g.ServeHTTP,
			Handle:    func(pattern string, h HandlerFunc, methods ...string) { r.Handle(pattern, h, nil, methods...) },
			Clean:     r.Clean,
			Remove:    r.Remove,
			URL:       r.URL,
		}
	})

	tt.Bench(b, h)
}

func BenchmarkNewContext(b *testing.B) {
	b.ReportAllocs()
	a := assert.New(b, false)
	s := newTestServer(a, false)

	w := httptest.NewRecorder()
	r := rest.Get(a, "/path").
		Header(header.ContentType, qheader.BuildContentType(header.JSON, "gbk")).
		Header(header.Accept, header.JSON).
		Header(header.AcceptEncoding, header.Identity).
		Request()
	for b.Loop() {
		ctx := s.NewContext(w, r, types.NewRoute())
		ctx.freeContext()
	}
}

func BenchmarkContext_Render(b *testing.B) {
	b.ReportAllocs()
	a := assert.New(b, false)
	s := newTestServer(a, false)

	b.Run("json;identity", func(b *testing.B) {
		b.ReportAllocs()
		var body *bytes.Buffer
		for b.Loop() {
			r := rest.Get(a, "/path").
				Header(header.AcceptEncoding, header.Identity).
				Header(header.Accept, header.JSON).
				Request()
			w := httptest.NewRecorder()

			ctx := s.NewContext(w, r, types.NewRoute())
			ctx.apply(Response(http.StatusCreated, objectInst))
			ctx.freeContext()

			body = w.Body
		}
		a.Equal(body.Bytes(), objectJSONString)
	})

	b.Run("json,deflate", func(b *testing.B) {
		b.ReportAllocs()
		var body *bytes.Buffer
		for b.Loop() {
			r := rest.Get(a, "/path").
				Header(header.Accept, header.JSON).
				Header(header.AcceptEncoding, "deflate").
				Request()
			w := httptest.NewRecorder()

			ctx := s.NewContext(w, r, types.NewRoute())
			ctx.apply(Response(http.StatusCreated, objectInst))
			ctx.freeContext()

			body = w.Body
		}
		data, err := io.ReadAll(flate.NewReader(body))
		a.NotError(err).NotNil(data).Equal(data, objectJSONString)
	})
}

func BenchmarkContext_Unmarshal(b *testing.B) {
	a := assert.New(b, false)
	srv := newTestServer(a, false)

	b.ReportAllocs()
	for b.Loop() {
		w := httptest.NewRecorder()
		r := rest.Post(a, "/path", []byte(objectJSONString)).
			Header(header.ContentType, qheader.BuildContentType(header.JSON, header.UTF8)).
			Header(header.Accept, header.JSON).
			Header(header.AcceptEncoding, header.Identity).
			Request()
		ctx := srv.NewContext(w, r, types.NewRoute())

		obj := &object{}
		a.NotError(ctx.Unmarshal(obj)).
			Equal(obj, objectInst)
		ctx.freeContext()
	}
}

// 一次普通的 POST 请求过程
func BenchmarkPost(b *testing.B) {
	a := assert.New(b, false)
	srv := newTestServer(a, false)

	b.ReportAllocs()
	for b.Loop() {
		w := httptest.NewRecorder()
		r := rest.Post(a, "/path", []byte(objectJSONString)).
			Header(header.ContentType, qheader.BuildContentType(header.JSON, header.UTF8)).
			Header(header.Accept, header.JSON).
			Header(header.AcceptEncoding, header.Identity).
			Request()
		ctx := srv.NewContext(w, r, types.NewRoute())

		o := &object{}
		a.NotError(ctx.Unmarshal(o)).
			Equal(o, objectInst)

		o.Age++
		o.Name = "response"
		ctx.apply(Response(http.StatusCreated, o))
		a.Equal(w.Body.String(), `{"name":"response","Age":457}`)
	}
}

func BenchmarkContext_Object(b *testing.B) {
	a := assert.New(b, false)
	s := newTestServer(a, false)
	o := &object{}

	b.ReportAllocs()
	for b.Loop() {
		w := httptest.NewRecorder()
		r := rest.Post(a, "/path", nil).
			Header(header.ContentType, qheader.BuildContentType(header.JSON, header.UTF8)).
			Header(header.Accept, header.JSON).
			Header(header.AcceptEncoding, header.Identity).
			Request()
		ctx := s.NewContext(w, r, types.NewRoute())
		ctx.apply(Response(http.StatusTeapot, o))
	}
}

func BenchmarkContext_Object_withHeader(b *testing.B) {
	a := assert.New(b, false)
	s := newTestServer(a, false)
	o := &object{}

	b.ReportAllocs()
	for b.Loop() {
		w := httptest.NewRecorder()
		r := rest.Post(a, "/path", nil).
			Header(header.ContentType, qheader.BuildContentType(header.JSON, header.UTF8)).
			Header(header.Accept, header.JSON).
			Header(header.AcceptEncoding, header.Identity).
			Request()
		ctx := s.NewContext(w, r, types.NewRoute())
		ctx.apply(Response(http.StatusTeapot, o, header.Location, "https://example.com"))
	}
}

func BenchmarkServer_Serve(b *testing.B) {
	a := assert.New(b, false)
	srv := newTestServer(a, false)
	router := srv.Routers().New("srv", nil)
	a.NotNil(router)

	router.Get("/path", func(c *Context) Responser {
		return Response(http.StatusOK, "/path", "h1", "h1")
	})

	defer srv.Run()()
	defer srv.Close(0)

	b.Run("encoding", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			r := srv.Get("/path").
				Header(header.ContentType, qheader.BuildContentType(header.JSON, "gbk")).
				Header(header.Accept, header.JSON).
				Header(header.AcceptEncoding, "gzip").
				Request()
			resp, err := http.DefaultClient.Do(r)
			if err != nil {
				b.Log(err)
			}
			_, err = io.ReadAll(resp.Body)
			if err != nil {
				b.Log(err)
			}
		}
	})

	b.Run("none", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			r := srv.Get("/path").
				Header(header.ContentType, qheader.BuildContentType(header.JSON, header.UTF8)).
				Header(header.Accept, header.JSON).
				Request()
			resp, err := http.DefaultClient.Do(r)
			if err != nil {
				b.Log(err)
			}
			_, err = io.ReadAll(resp.Body)
			if err != nil {
				b.Log(err)
			}
		}
	})
}
