// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package web

import (
	"bytes"
	"compress/flate"
	"encoding/json/v2"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/goccy/go-yaml"
	"github.com/issue9/assert/v6"
	"github.com/issue9/assert/v6/rest"
	"github.com/issue9/mux/v10"
	"github.com/issue9/mux/v10/header"
	"github.com/issue9/mux/v10/routertest"
	"github.com/issue9/mux/v10/types"
	"golang.org/x/text/language"

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
		Header(header.AcceptCharset, "gbk").
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

	b.Run("none", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			r := rest.Get(a, "/path").Header(header.AcceptEncoding, header.Identity).Header(header.Accept, header.JSON).Request()
			w := httptest.NewRecorder()

			ctx := s.NewContext(w, r, types.NewRoute())
			ctx.apply(Response(http.StatusCreated, objectInst))
			ctx.freeContext()

			a.Equal(w.Body.Bytes(), objectJSONString)
		}
	})

	b.Run("utf8", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			r := rest.Get(a, "/path").
				Header(header.Accept, header.JSON).
				Header(header.AcceptCharset, header.UTF8).
				Header(header.AcceptEncoding, header.Identity).
				Request()
			w := httptest.NewRecorder()
			ctx := s.NewContext(w, r, types.NewRoute())
			ctx.apply(Response(http.StatusCreated, objectInst))
			ctx.freeContext()

			a.Equal(w.Body.Bytes(), objectJSONString)
		}
	})

	b.Run("gbk", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			r := rest.Get(a, "/path").
				Header(header.Accept, header.JSON).
				Header(header.AcceptCharset, "gbk").
				Header(header.AcceptEncoding, header.Identity).
				Request()
			w := httptest.NewRecorder()

			ctx := s.NewContext(w, r, types.NewRoute())
			ctx.apply(Response(http.StatusCreated, objectInst))
			ctx.freeContext()

			a.Equal(w.Body.Bytes(), objectGBKBytes)
		}
	})

	b.Run("charset; encoding", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			r := rest.Get(a, "/path").
				Header(header.Accept, header.JSON).
				Header(header.AcceptCharset, "gbk").
				Header(header.AcceptEncoding, "deflate").
				Request()
			w := httptest.NewRecorder()

			ctx := s.NewContext(w, r, types.NewRoute())
			ctx.apply(Response(http.StatusCreated, objectInst))
			ctx.freeContext()

			data, err := io.ReadAll(flate.NewReader(w.Body))
			a.NotError(err).NotNil(data).Equal(data, objectGBKBytes)
		}
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

func BenchmarkNewProblem(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		p := newProblem()
		p.WithExtensions(&object{Name: "n1", Age: 11}).WithParam("p1", "v1")
		problemPool.Put(p)
	}
}

func BenchmarkProblem_JSON(b *testing.B) {
	benchProblem_marshal_unmarshal(b, func(a any) ([]byte, error) { return json.Marshal(a) }, func(b []byte, a any) error { return json.Unmarshal(b, a) })
}

func BenchmarkProblem_CBOR(b *testing.B) {
	benchProblem_marshal_unmarshal(b, cbor.Marshal, cbor.Unmarshal)
}

func BenchmarkProblem_YAML(b *testing.B) {
	benchProblem_marshal_unmarshal(b, yaml.Marshal, yaml.Unmarshal)
}

func benchProblem_marshal_unmarshal(b *testing.B, marshal func(any) ([]byte, error), unmarshal func([]byte, any) error) {
	a := assert.New(b, false)

	p := newProblem()
	p.Type = "https://iana.org/assignments/http-problem-types#date"
	p.Title = "http-problem-types#date"
	p.Detail = "https://iana.org/assignments/http-problem-types#date"
	p.Status = 400
	p.WithInstance("500")
	p.WithExtensions("extension")
	p.WithParam("p1", "p1-1")
	p.WithParam("p1", "p1-2")
	p.WithParam("p2", "p2")

	var data []byte
	b.Run("marshal", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			data, _ = marshal(p)
		}
	})

	out := newProblem()
	b.Run("unmarshal", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			unmarshal(data, out)
		}
	})

	a.Equal(p, out)
}

func BenchmarkCodec_accept(b *testing.B) {
	a := assert.New(b, false)
	mt := newCodec(a)
	b.ReportAllocs()

	var item MediaType
	for b.Loop() {
		item, _ = mt.accept("application/json;q=0.9")
	}
	a.NotNil(item)
}

func BenchmarkCodec_contentType(b *testing.B) {
	a := assert.New(b, false)
	mt := newCodec(a)

	b.Run("charset=utf-8", func(b *testing.B) {
		a := assert.New(b, false)
		b.ResetTimer()
		b.ReportAllocs()
		for b.Loop() {
			marshal, encoding, err := mt.contentType(qheader.BuildContentType(header.XML, header.UTF8))
			a.NotError(err).NotNil(marshal).Nil(encoding)
		}
	})

	b.Run("charset=gbk", func(b *testing.B) {
		a := assert.New(b, false)
		b.ResetTimer()
		b.ReportAllocs()
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
		c.AddCompressor(newZstd(), "application/*")
		b.ReportAllocs()

		r := bytes.NewBuffer([]byte{})
		var err error
		for b.Loop() {
			r.Reset()
			_, err = c.contentEncoding("zstd", r)
		}
		a.NotError(err)
	})

	b.Run("5", func(b *testing.B) {
		a := assert.New(b, false)
		b.ReportAllocs()

		c := NewCodec()
		a.NotNil(c)
		c.AddCompressor(newGzip(), "application/*").
			AddCompressor(newBrotli(), "text/*").
			AddCompressor(newDeflate(), "image/*").
			AddCompressor(newZstd(), "application/*").
			AddCompressor(newLZW(), header.Plain)

		r := bytes.NewBuffer([]byte{})
		var err error
		for b.Loop() {
			r.Reset()
			_, err = c.contentEncoding("zstd", r)
		}
		a.NotError(err)
	})
}

func BenchmarkCodec_acceptEncoding(b *testing.B) {
	b.Run("1", func(b *testing.B) {
		a := assert.New(b, false)
		b.ReportAllocs()

		c := NewCodec()
		a.NotNil(c)
		c.AddCompressor(newZstd(), "application/*")

		var err error
		for b.Loop() {
			_, err = c.acceptEncoding(header.JSON, "zstd")
		}
		a.NotError(err)
	})

	b.Run("5", func(b *testing.B) {
		a := assert.New(b, false)
		b.ReportAllocs()

		c := NewCodec()
		a.NotNil(c)
		c.AddCompressor(newGzip(), "application/*").
			AddCompressor(newBrotli(), "text/*").
			AddCompressor(newDeflate(), "image/*").
			AddCompressor(newZstd(), "application/*").
			AddCompressor(newLZW(), header.Plain)

		var err error
		for b.Loop() {
			_, err = c.acceptEncoding(header.Plain, "compress")
		}
		a.NotError(err)
	})
}

func BenchmarkCodec_getMatchCompresses(b *testing.B) {
	a := assert.New(b, false)
	b.ReportAllocs()

	c := NewCodec()
	a.NotNil(c)
	c.AddCompressor(newGzip(), "application/*").
		AddCompressor(newBrotli(), "text/*").
		AddCompressor(newDeflate(), "image/*").
		AddCompressor(newZstd(), "application/*").
		AddCompressor(newLZW(), header.Plain)

	for b.Loop() {
		c.getMatchCompresses(header.Plain)
	}
}

func BenchmarkLocale_NewPrinter(b *testing.B) {
	l := newLocale(language.SimplifiedChinese, nil)

	b.Run("equal Locale.id", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			l.NewPrinter(language.SimplifiedChinese)
		}
	})

	b.Run("not equal Locale.id", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			l.NewPrinter(language.TraditionalChinese)
		}
	})

	langs := []language.Tag{
		language.Chinese,
		language.SimplifiedChinese,
		language.TraditionalChinese,
		language.MustParse("zh-CN"),
		language.MustParse("cmn-Hans"),
	}
	b.Run("rand id", func(b *testing.B) {
		b.ReportAllocs()
		size := len(langs)
		for i := range b.N {
			l.NewPrinter(langs[i%size])
		}
	})
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
	time.Sleep(500 * time.Millisecond)

	b.Run("charset", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			r := srv.Get("/path").
				Header(header.ContentType, qheader.BuildContentType(header.JSON, "gbk")).
				Header(header.Accept, header.JSON).
				Header(header.AcceptCharset, "gbk;q=1,gb18080;q=0.1").
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

	b.Run("charset encoding", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			r := srv.Get("/path").
				Header(header.ContentType, qheader.BuildContentType(header.JSON, "gbk")).
				Header(header.Accept, header.JSON).
				Header(header.AcceptCharset, "gbk;q=1,gb18080;q=0.1").
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
