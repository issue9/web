// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package web

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"compress/lzw"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/andybalholm/brotli"
	"github.com/issue9/assert/v5"
	"github.com/issue9/mux/v9/header"
	"github.com/issue9/mux/v9/routertest"
	"github.com/issue9/mux/v9/types"

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

	routertest.NewTester(s.internalServer.call, notFound, buildNodeHandle(http.StatusMethodNotAllowed), buildNodeHandle(http.StatusOK)).Bench(b, h)
}

func BenchmarkNewContext(b *testing.B) {
	a := assert.New(b, false)
	s := newTestServer(a)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Set(header.ContentType, qheader.BuildContentType(header.JSON, "gbk"))
	r.Header.Set(header.Accept, header.JSON)
	r.Header.Set(header.AcceptCharset, "gbk")
	for b.Loop() {
		ctx := s.NewContext(w, r, types.NewContext())
		ctx.freeContext()
	}
}

func BenchmarkContext_Render(b *testing.B) {
	a := assert.New(b, false)
	s := newTestServer(a)

	b.Run("none", func(b *testing.B) {
		for b.Loop() {
			r := httptest.NewRequest(http.MethodGet, "/path", nil)
			r.Header.Set(header.Accept, header.JSON)
			w := httptest.NewRecorder()

			ctx := s.NewContext(w, r, types.NewContext())
			ctx.apply(Response(http.StatusCreated, objectInst))
			ctx.freeContext()

			a.Equal(w.Body.Bytes(), objectJSONString)
		}
	})

	b.Run("utf8", func(b *testing.B) {
		for b.Loop() {
			r := httptest.NewRequest(http.MethodGet, "/path", nil)
			r.Header.Set(header.Accept, header.JSON)
			r.Header.Set(header.AcceptCharset, header.UTF8)
			w := httptest.NewRecorder()
			ctx := s.NewContext(w, r, types.NewContext())
			ctx.apply(Response(http.StatusCreated, objectInst))
			ctx.freeContext()

			a.Equal(w.Body.Bytes(), objectJSONString)
		}
	})

	b.Run("gbk", func(b *testing.B) {
		for b.Loop() {
			r := httptest.NewRequest(http.MethodGet, "/path", nil)
			r.Header.Set(header.Accept, header.JSON)
			r.Header.Set(header.AcceptCharset, "gbk")
			w := httptest.NewRecorder()

			ctx := s.NewContext(w, r, types.NewContext())
			ctx.apply(Response(http.StatusCreated, objectInst))
			ctx.freeContext()

			a.Equal(w.Body.Bytes(), objectGBKBytes)
		}
	})

	b.Run("charset; encoding", func(b *testing.B) {
		for b.Loop() {
			r := httptest.NewRequest(http.MethodGet, "/path", nil)
			r.Header.Set(header.Accept, header.JSON)
			r.Header.Set(header.AcceptCharset, "gbk")
			r.Header.Set(header.AcceptEncoding, "deflate")
			w := httptest.NewRecorder()

			ctx := s.NewContext(w, r, types.NewContext())
			ctx.apply(Response(http.StatusCreated, objectInst))
			ctx.freeContext()

			data, err := io.ReadAll(flate.NewReader(w.Body))
			a.NotError(err).NotNil(data).Equal(data, objectGBKBytes)
		}
	})
}

func BenchmarkContext_Unmarshal(b *testing.B) {
	a := assert.New(b, false)
	srv := newTestServer(a)

	for b.Loop() {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/path", bytes.NewBufferString(objectJSONString))
		r.Header.Set(header.ContentType, qheader.BuildContentType(header.JSON, header.UTF8))
		r.Header.Set(header.Accept, header.JSON)
		ctx := srv.NewContext(w, r, types.NewContext())

		obj := &object{}
		a.NotError(ctx.Unmarshal(obj)).
			Equal(obj, objectInst)
		ctx.freeContext()
	}
}

// 一次普通的 POST 请求过程
func BenchmarkPost(b *testing.B) {
	a := assert.New(b, false)
	srv := newTestServer(a)

	for b.Loop() {
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

	for b.Loop() {
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

	for b.Loop() {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/path", nil)
		r.Header.Set(header.ContentType, qheader.BuildContentType(header.JSON, header.UTF8))
		r.Header.Set(header.Accept, header.JSON)
		ctx := s.NewContext(w, r, types.NewContext())
		ctx.apply(Response(http.StatusTeapot, o, header.Location, "https://example.com"))
	}
}

func BenchmarkNewProblem(b *testing.B) {
	for b.Loop() {
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
	for b.Loop() {
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
	defer ctx.freeContext()

	for b.Loop() {
		p := ctx.NewFilterContext(false)
		filterContextPool.Put(p)
	}
}

func BenchmarkCodec_NewEncoder(b *testing.B) {
	b.Run("gzip", func(b *testing.B) {
		benchCompressor_NewEncoder(b, NewGzip(3))
	})

	b.Run("zstd", func(b *testing.B) {
		benchCompressor_NewEncoder(b, NewZstd())
	})

	b.Run("deflate", func(b *testing.B) {
		benchCompressor_NewEncoder(b, NewDeflate(3, nil))
	})

	b.Run("lzw", func(b *testing.B) {
		benchCompressor_NewEncoder(b, NewLZW(lzw.LSB, 5))
	})

	b.Run("br", func(b *testing.B) {
		benchCompressor_NewEncoder(b, NewBrotli(brotli.WriterOptions{}))
	})
}

func BenchmarkCodec_NewDecoder(b *testing.B) {
	a := assert.New(b, false)

	b.Run("gzip", func(b *testing.B) {
		c := NewGzip(3)
		for b.Loop() {
			wc, err := c.NewDecoder(bytes.NewBuffer(gzipInitData))
			a.NotError(err).
				NotNil(wc).
				NotError(wc.Close())
		}
	})

	b.Run("zstd", func(b *testing.B) {
		c := NewZstd()
		for b.Loop() {
			wc, err := c.NewDecoder(bytes.NewBuffer(zstdInitData))
			a.NotError(err).
				NotNil(wc).
				NotError(wc.Close())
		}
	})

	b.Run("deflate", func(b *testing.B) {
		c := NewDeflate(3, nil)
		for b.Loop() {
			wc, err := c.NewDecoder(bytes.NewBuffer(deflateInitData))
			a.NotError(err).
				NotNil(wc).
				NotError(wc.Close())
		}
	})

	b.Run("lzw", func(b *testing.B) {
		c := NewLZW(lzw.LSB, 5)
		for b.Loop() {
			wc, err := c.NewDecoder(bytes.NewBuffer(lzwInitData))
			a.NotError(err).
				NotNil(wc).
				NotError(wc.Close())
		}
	})

	b.Run("br", func(b *testing.B) {
		c := NewBrotli(brotli.WriterOptions{})
		for b.Loop() {
			wc, err := c.NewDecoder(bytes.NewBuffer(brotliInitData))
			a.NotError(err).
				NotNil(wc).
				NotError(wc.Close())
		}
	})
}

func benchCompressor_NewEncoder(b *testing.B, c Compressor) {
	a := assert.New(b, false)
	w := &bytes.Buffer{}
	for b.Loop() {
		w.Reset()

		wc, err := c.NewEncoder(w)
		a.NotError(err).
			NotNil(wc).
			NotError(wc.Close())
	}
}

func BenchmarkCodec_accept(b *testing.B) {
	a := assert.New(b, false)
	mt := newCodec(a)

	for b.Loop() {
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
		for b.Loop() {
			marshal, encoding, err := mt.contentType(qheader.BuildContentType(header.XML, header.UTF8))
			a.NotError(err).NotNil(marshal).Nil(encoding)
		}
	})

	b.Run("charset=gbk", func(b *testing.B) {
		a := assert.New(b, false)
		b.ResetTimer()
		for b.Loop() {
			marshal, encoding, err := mt.contentType(qheader.BuildContentType(header.XML, "gbk"))
			a.NotError(err).NotNil(marshal).NotNil(encoding)
		}
	})
}

func BenchmarkCodec_contentEncoding(b *testing.B) {
	b.Run("1", func(b *testing.B) {
		a := assert.New(b, false)

		c := NewCodec()
		a.NotNil(c)
		c.AddCompressor(NewZstd(), "application/*")

		for b.Loop() {
			r := bytes.NewBuffer([]byte{})
			_, err := c.contentEncoding("zstd", r)
			a.NotError(err)
		}
	})

	b.Run("5", func(b *testing.B) {
		a := assert.New(b, false)

		c := NewCodec()
		a.NotNil(c)
		c.AddCompressor(NewGzip(gzip.DefaultCompression), "application/*").
			AddCompressor(NewBrotli(brotli.WriterOptions{}), "text/*").
			AddCompressor(NewDeflate(flate.BestCompression, nil), "image/*").
			AddCompressor(NewZstd(), "application/*").
			AddCompressor(NewLZW(lzw.LSB, 8), header.Plain)

		for b.Loop() {
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
		c.AddCompressor(NewZstd(), "application/*")

		for b.Loop() {
			_, na := c.acceptEncoding(header.JSON, "zstd")
			a.False(na)
		}
	})

	b.Run("5", func(b *testing.B) {
		a := assert.New(b, false)

		c := NewCodec()
		a.NotNil(c)
		c.AddCompressor(NewGzip(gzip.DefaultCompression), "application/*").
			AddCompressor(NewBrotli(brotli.WriterOptions{}), "text/*").
			AddCompressor(NewDeflate(flate.BestCompression, nil), "image/*").
			AddCompressor(NewZstd(), "application/*").
			AddCompressor(NewLZW(lzw.LSB, 8), header.Plain)

		for b.Loop() {
			_, na := c.acceptEncoding(header.Plain, "compress")
			a.False(na)
		}
	})
}

func BenchmarkCodec_getMatchCompresses(b *testing.B) {
	a := assert.New(b, false)

	c := NewCodec()
	a.NotNil(c)
	c.AddCompressor(NewGzip(gzip.DefaultCompression), "application/*").
		AddCompressor(NewBrotli(brotli.WriterOptions{}), "text/*").
		AddCompressor(NewDeflate(flate.BestCompression, nil), "image/*").
		AddCompressor(NewZstd(), "application/*").
		AddCompressor(NewLZW(lzw.LSB, 8), header.Plain)

	for b.Loop() {
		c.getMatchCompresses(header.Plain)
	}
}
