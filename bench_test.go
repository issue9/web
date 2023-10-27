// SPDX-License-Identifier: MIT

package web

import (
	"bytes"
	"compress/flate"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert/v3"

	"github.com/issue9/web/internal/header"
	"github.com/issue9/web/internal/testdata"
)

func BenchmarkNewContext(b *testing.B) {
	a := assert.New(b, false)
	s := newTestServer(a)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Set(header.ContentType, header.BuildContentType("application/json", "gbk"))
	r.Header.Set(header.Accept, "application/json")
	r.Header.Set(header.AcceptCharset, "gbk")
	for i := 0; i < b.N; i++ {
		ctx := NewContext(s, w, r, nil, header.RequestIDKey)
		ctx.Free()
	}
}

func BenchmarkContext_Render(b *testing.B) {
	a := assert.New(b, false)
	srv := newTestServer(a)

	b.Run("none", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			r := httptest.NewRequest(http.MethodGet, "/path", nil)
			r.Header.Set(header.Accept, "application/json")
			w := httptest.NewRecorder()

			ctx := srv.NewContext(w, r)
			ctx.apply(Response(http.StatusCreated, testdata.ObjectInst))
			ctx.Free()

			a.Equal(w.Body.Bytes(), testdata.ObjectJSONString)
		}
	})

	b.Run("utf8", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			r := httptest.NewRequest(http.MethodGet, "/path", nil)
			r.Header.Set(header.Accept, "application/json")
			r.Header.Set(header.AcceptCharset, header.UTF8Name)
			w := httptest.NewRecorder()
			ctx := srv.NewContext(w, r)
			ctx.apply(Response(http.StatusCreated, testdata.ObjectInst))
			ctx.Free()

			a.Equal(w.Body.Bytes(), testdata.ObjectJSONString)
		}
	})

	b.Run("gbk", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			r := httptest.NewRequest(http.MethodGet, "/path", nil)
			r.Header.Set(header.Accept, "application/json")
			r.Header.Set(header.AcceptCharset, "gbk")
			w := httptest.NewRecorder()

			ctx := srv.NewContext(w, r)
			ctx.apply(Response(http.StatusCreated, testdata.ObjectInst))
			ctx.Free()

			a.Equal(w.Body.Bytes(), testdata.ObjectGBKBytes)
		}
	})

	b.Run("charset; encoding", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			r := httptest.NewRequest(http.MethodGet, "/path", nil)
			r.Header.Set(header.Accept, "application/json")
			r.Header.Set(header.AcceptCharset, "gbk")
			r.Header.Set(header.AcceptEncoding, "deflate")
			w := httptest.NewRecorder()

			ctx := srv.NewContext(w, r)
			ctx.apply(Response(http.StatusCreated, testdata.ObjectInst))
			ctx.Free()

			data, err := io.ReadAll(flate.NewReader(w.Body))
			a.NotError(err).NotNil(data)
			a.Equal(data, testdata.ObjectGBKBytes)
		}
	})
}

func BenchmarkContext_Unmarshal(b *testing.B) {
	a := assert.New(b, false)
	srv := newTestServer(a)

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/path", bytes.NewBufferString(testdata.ObjectJSONString))
		r.Header.Set(header.ContentType, header.BuildContentType("application/json", header.UTF8Name))
		r.Header.Set(header.Accept, "application/json")
		ctx := srv.NewContext(w, r)

		obj := &testdata.Object{}
		a.NotError(ctx.Unmarshal(obj)).
			Equal(obj, testdata.ObjectInst)
		ctx.Free()
	}
}

// 一次普通的 POST 请求过程
func BenchmarkPost(b *testing.B) {
	a := assert.New(b, false)
	srv := newTestServer(a)

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/path", bytes.NewBufferString(testdata.ObjectJSONString))
		r.Header.Set(header.ContentType, header.BuildContentType("application/json", header.UTF8Name))
		r.Header.Set(header.Accept, "application/json")
		ctx := srv.NewContext(w, r)

		o := &testdata.Object{}
		a.NotError(ctx.Unmarshal(o)).
			Equal(o, testdata.ObjectInst)

		o.Age++
		o.Name = "response"
		ctx.apply(Response(http.StatusCreated, o))
		a.Equal(w.Body.String(), `{"name":"response","Age":457}`)
	}
}

func BenchmarkContext_Object(b *testing.B) {
	a := assert.New(b, false)
	s := newTestServer(a)
	o := &testdata.Object{}

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/path", nil)
		r.Header.Set(header.ContentType, header.BuildContentType("application/json", header.UTF8Name))
		r.Header.Set(header.Accept, "application/json")
		ctx := s.NewContext(w, r)
		ctx.apply(Response(http.StatusTeapot, o))
	}
}

func BenchmarkContext_Object_withHeader(b *testing.B) {
	a := assert.New(b, false)
	s := newTestServer(a)
	o := &testdata.Object{}

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/path", nil)
		r.Header.Set(header.ContentType, header.BuildContentType("application/json", header.UTF8Name))
		r.Header.Set(header.Accept, "application/json")
		ctx := s.NewContext(w, r)
		ctx.apply(Response(http.StatusTeapot, o, "Location", "https://example.com"))
	}
}

func BenchmarkNewRFC7807(b *testing.B) {
	for i := 0; i < b.N; i++ {
		p := newRFC7807()
		p.Init("id", "title", "detail", 400)
		p.WithExtensions(&object{Name: "n1", Age: 11}).WithParam("p1", "v1")
		rfc7807Pool.Put(p)
	}
}

func BenchmarkRFC7807_unmarshal_json(b *testing.B) {
	a := assert.New(b, false)
	s := newTestServer(a)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/path", nil)
	r.Header.Set(header.ContentType, header.BuildContentType("application/json", header.UTF8Name))
	r.Header.Set(header.Accept, "application/json")
	ctx := s.NewContext(w, r)

	p := newRFC7807()
	p.Init("id", "title", "detail", 400)
	p.WithExtensions(&object{Name: "n1", Age: 11}).WithParam("p1", "v1")
	for i := 0; i < b.N; i++ {
		p.Apply(ctx)
	}
}

func BenchmarkNewFilterContext(b *testing.B) {
	a := assert.New(b, false)
	s := newTestServer(a)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/path", nil)
	r.Header.Set(header.ContentType, header.BuildContentType("application/json", header.UTF8Name))
	r.Header.Set(header.Accept, "application/json")
	ctx := s.NewContext(w, r)
	defer ctx.Free()

	for i := 0; i < b.N; i++ {
		p := ctx.newFilterContext(false)
		filterContextPool.Put(p)
	}
}
