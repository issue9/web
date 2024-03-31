// SPDX-FileCopyrightText: 2018-2024 caixw
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

	"github.com/issue9/assert/v4"
	"github.com/issue9/mux/v8/header"
	"github.com/issue9/mux/v8/routertest"
	"github.com/issue9/mux/v8/types"

	"github.com/issue9/web/internal/qheader"
)

func BenchmarkRouter(b *testing.B) {
	a := assert.New(b, false)
	s := newTestServer(a)

	h := func(c *Context) Responser {
		_, err := c.Write([]byte(c.Request().URL.Path))
		if err != nil {
			b.Error(err)
		}
		return nil
	}

	routertest.NewTester(s.InternalServer.call, notFound, buildNodeHandle(http.StatusMethodNotAllowed), buildNodeHandle(http.StatusOK)).Bench(b, h)
}

func BenchmarkNewContext(b *testing.B) {
	a := assert.New(b, false)
	s := newTestServer(a)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Set(header.ContentType, qheader.BuildContentType("application/json", "gbk"))
	r.Header.Set(header.Accept, "application/json")
	r.Header.Set(header.AcceptCharset, "gbk")
	for range b.N {
		ctx := s.NewContext(w, r, types.NewContext())
		s.freeContext(ctx)
	}
}

func BenchmarkContext_Render(b *testing.B) {
	a := assert.New(b, false)
	s := newTestServer(a)

	b.Run("none", func(b *testing.B) {
		for range b.N {
			r := httptest.NewRequest(http.MethodGet, "/path", nil)
			r.Header.Set(header.Accept, header.JSON)
			w := httptest.NewRecorder()

			ctx := s.NewContext(w, r, types.NewContext())
			ctx.apply(Response(http.StatusCreated, objectInst))
			s.freeContext(ctx)

			a.Equal(w.Body.Bytes(), objectJSONString)
		}
	})

	b.Run("utf8", func(b *testing.B) {
		for range b.N {
			r := httptest.NewRequest(http.MethodGet, "/path", nil)
			r.Header.Set(header.Accept, header.JSON)
			r.Header.Set(header.AcceptCharset, header.UTF8)
			w := httptest.NewRecorder()
			ctx := s.NewContext(w, r, types.NewContext())
			ctx.apply(Response(http.StatusCreated, objectInst))
			s.freeContext(ctx)

			a.Equal(w.Body.Bytes(), objectJSONString)
		}
	})

	b.Run("gbk", func(b *testing.B) {
		for range b.N {
			r := httptest.NewRequest(http.MethodGet, "/path", nil)
			r.Header.Set(header.Accept, header.JSON)
			r.Header.Set(header.AcceptCharset, "gbk")
			w := httptest.NewRecorder()

			ctx := s.NewContext(w, r, types.NewContext())
			ctx.apply(Response(http.StatusCreated, objectInst))
			s.freeContext(ctx)

			a.Equal(w.Body.Bytes(), objectGBKBytes)
		}
	})

	b.Run("charset; encoding", func(b *testing.B) {
		for range b.N {
			r := httptest.NewRequest(http.MethodGet, "/path", nil)
			r.Header.Set(header.Accept, header.JSON)
			r.Header.Set(header.AcceptCharset, "gbk")
			r.Header.Set(header.AcceptEncoding, "deflate")
			w := httptest.NewRecorder()

			ctx := s.NewContext(w, r, types.NewContext())
			ctx.apply(Response(http.StatusCreated, objectInst))
			s.freeContext(ctx)

			data, err := io.ReadAll(flate.NewReader(w.Body))
			a.NotError(err).NotNil(data).Equal(data, objectGBKBytes)
		}
	})
}

func BenchmarkContext_Unmarshal(b *testing.B) {
	a := assert.New(b, false)
	srv := newTestServer(a)

	for range b.N {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/path", bytes.NewBufferString(objectJSONString))
		r.Header.Set(header.ContentType, qheader.BuildContentType(header.JSON, header.UTF8))
		r.Header.Set(header.Accept, header.JSON)
		ctx := srv.NewContext(w, r, types.NewContext())

		obj := &object{}
		a.NotError(ctx.Unmarshal(obj)).
			Equal(obj, objectInst)
		srv.freeContext(ctx)
	}
}

// 一次普通的 POST 请求过程
func BenchmarkPost(b *testing.B) {
	a := assert.New(b, false)
	srv := newTestServer(a)

	for range b.N {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/path", bytes.NewBufferString(objectJSONString))
		r.Header.Set(header.ContentType, qheader.BuildContentType(header.JSON, header.UTF8))
		r.Header.Set(header.Accept, header.JSON)
		ctx := srv.NewContext(w, r, types.NewContext())

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
	s := newTestServer(a)
	o := &object{}

	for range b.N {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/path", nil)
		r.Header.Set(header.ContentType, qheader.BuildContentType(header.JSON, header.UTF8))
		r.Header.Set(header.Accept, header.JSON)
		ctx := s.NewContext(w, r, types.NewContext())
		ctx.apply(Response(http.StatusTeapot, o))
	}
}

func BenchmarkContext_Object_withHeader(b *testing.B) {
	a := assert.New(b, false)
	s := newTestServer(a)
	o := &object{}

	for range b.N {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/path", nil)
		r.Header.Set(header.ContentType, qheader.BuildContentType(header.JSON, header.UTF8))
		r.Header.Set(header.Accept, header.JSON)
		ctx := s.NewContext(w, r, types.NewContext())
		ctx.apply(Response(http.StatusTeapot, o, header.Location, "https://example.com"))
	}
}

func BenchmarkNewProblem(b *testing.B) {
	for range b.N {
		p := newProblem()
		p.Type = "id"
		p.Title = "title"
		p.Detail = "detail"
		p.Status = 400
		p.WithExtensions(&object{Name: "n1", Age: 11}).WithParam("p1", "v1")
		problemPool.Put(p)
	}
}

func BenchmarkProblem_unmarshal_json(b *testing.B) {
	a := assert.New(b, false)
	s := newTestServer(a)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/path", nil)
	r.Header.Set(header.ContentType, qheader.BuildContentType(header.JSON, header.UTF8))
	r.Header.Set(header.Accept, header.JSON)
	ctx := s.NewContext(w, r, types.NewContext())

	p := newProblem()
	p.Type = "id"
	p.Title = "title"
	p.Detail = "detail"
	p.Status = 400
	p.WithExtensions(&object{Name: "n1", Age: 11}).WithParam("p1", "v1")
	for range b.N {
		p.Apply(ctx)
	}
}

func BenchmarkNewFilterContext(b *testing.B) {
	a := assert.New(b, false)
	s := newTestServer(a)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/path", nil)
	r.Header.Set(header.ContentType, qheader.BuildContentType(header.JSON, header.UTF8))
	r.Header.Set(header.Accept, header.JSON)
	ctx := s.NewContext(w, r, types.NewContext())
	defer s.freeContext(ctx)

	for range b.N {
		p := ctx.NewFilterContext(false)
		filterContextPool.Put(p)
	}
}

func BenchmarkCodec_accept(b *testing.B) {
	a := assert.New(b, false)
	mt := newCodec(a)

	for range b.N {
		item := mt.accept("application/json;q=0.9")
		a.NotNil(item)
	}
}

func BenchmarkCodec_contentType(b *testing.B) {
	a := assert.New(b, false)
	mt := newCodec(a)

	b.Run("charset=utf-8", func(b *testing.B) {
		a := assert.New(b, false)
		b.ResetTimer()
		for range b.N {
			marshal, encoding, err := mt.contentType("application/xml;charset=utf-8")
			a.NotError(err).NotNil(marshal).Nil(encoding)
		}
	})

	b.Run("charset=gbk", func(b *testing.B) {
		a := assert.New(b, false)
		b.ResetTimer()
		for range b.N {
			marshal, encoding, err := mt.contentType("application/xml;charset=gbk")
			a.NotError(err).NotNil(marshal).NotNil(encoding)
		}
	})
}

func BenchmarkCodec_contentEncoding(b *testing.B) {
	b.Run("1", func(b *testing.B) {
		a := assert.New(b, false)

		c := NewCodec()
		a.NotNil(c)
		c.AddCompressor(&compressorTest{name: "zstd"}, "application/*")

		for range b.N {
			r := bytes.NewBuffer([]byte{})
			_, err := c.contentEncoding("zstd", r)
			a.NotError(err)
		}
	})

	b.Run("5", func(b *testing.B) {
		a := assert.New(b, false)

		c := NewCodec()
		a.NotNil(c)
		c.AddCompressor(&compressorTest{name: "gzip"}, "application/*").
			AddCompressor(&compressorTest{name: "br"}, "text/*").
			AddCompressor(&compressorTest{name: "deflate"}, "image/*").
			AddCompressor(&compressorTest{name: "zstd"}, "application/*").
			AddCompressor(&compressorTest{name: "compress"}, "text/plain")

		for range b.N {
			r := bytes.NewBuffer([]byte{})
			_, err := c.contentEncoding("zstd", r)
			a.NotError(err)
		}
	})
}

func BenchmarkCodec_acceptEncoding(b *testing.B) {
	b.Run("1", func(b *testing.B) {
		a := assert.New(b, false)

		c := NewCodec()
		a.NotNil(c)
		c.AddCompressor(&compressorTest{name: "zstd"}, "application/*")

		for range b.N {
			_, na := c.acceptEncoding("application/json", "zstd")
			a.False(na)
		}
	})

	b.Run("5", func(b *testing.B) {
		a := assert.New(b, false)

		c := NewCodec()
		a.NotNil(c)
		c.AddCompressor(&compressorTest{name: "gzip"}, "application/*").
			AddCompressor(&compressorTest{name: "br"}, "text/*").
			AddCompressor(&compressorTest{name: "deflate"}, "image/*").
			AddCompressor(&compressorTest{name: "zstd"}, "application/*").
			AddCompressor(&compressorTest{name: "compress"}, "text/plain")

		for range b.N {
			_, na := c.acceptEncoding("text/plain", "compress")
			a.False(na)
		}
	})
}

func BenchmarkCodec_getMatchCompresses(b *testing.B) {
	a := assert.New(b, false)

	c := NewCodec()
	a.NotNil(c)
	c.AddCompressor(&compressorTest{name: "gzip"}, "application/*").
		AddCompressor(&compressorTest{name: "br"}, "text/*").
		AddCompressor(&compressorTest{name: "deflate"}, "image/*").
		AddCompressor(&compressorTest{name: "zstd"}, "application/*").
		AddCompressor(&compressorTest{name: "compress"}, "text/plain")

	for range b.N {
		c.getMatchCompresses("text/plan")
	}
}
