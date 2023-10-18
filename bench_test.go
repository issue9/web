// SPDX-License-Identifier: MIT

package web

import (
	"compress/flate"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/assert/v3/rest"

	"github.com/issue9/web/internal/header"
	"github.com/issue9/web/internal/testdata"
	"github.com/issue9/web/servertest"
)

func BenchmarkNewContext(b *testing.B) {
	a := assert.New(b, false)
	srv := newTestServer(a)

	w := httptest.NewRecorder()
	r := servertest.Get(a, "/path").
		Header("Content-type", header.BuildContentType("application/json", "gbk")).
		Header("Accept", "application/json").
		Header("Accept-Charset", "gbk;q=1,gb18080;q=0.1").
		Request()
	for i := 0; i < b.N; i++ {
		ctx := srv.NewContext(w, r)
		ctx.Free()
	}
}

func BenchmarkContext_Render(b *testing.B) {
	a := assert.New(b, false)
	srv := newTestServer(a)

	b.Run("none", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			r := rest.Get(a, "/path").Header("Accept", "application/json").Request()
			ctx := srv.NewContext(w, r)

			Response(http.StatusCreated, testdata.ObjectInst).Apply(ctx)
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
			ctx := srv.NewContext(w, r)

			Response(http.StatusCreated, testdata.ObjectInst).Apply(ctx)
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
			ctx := srv.NewContext(w, r)

			Response(http.StatusCreated, testdata.ObjectInst).Apply(ctx)
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

			ctx := srv.NewContext(w, r)
			Response(http.StatusCreated, testdata.ObjectInst).Apply(ctx)
			ctx.Free()

			data, err := io.ReadAll(flate.NewReader(w.Body))
			a.NotError(err).NotNil(data)
			a.Equal(data, testdata.ObjectGBKBytes)
		}
	})
}

func BenchmarkContext_RequestBody(b *testing.B) {
	a := assert.New(b, false)
	srv := newTestServer(a)

	b.Run("empty", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			r := rest.Post(a, "/path", nil).
				Header("Content-type", header.BuildContentType("application/json", "utf-8")).
				Header("Accept", "application/json").
				Request()
			ctx := srv.NewContext(w, r)

			body, err := ctx.RequestBody()
			a.NotError(err).Empty(body)
		}
	})

	b.Run("charset=utf-8", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			r := rest.Post(a, "/path", []byte(testdata.ObjectJSONString)).
				Header("Content-type", header.BuildContentType("application/json", "utf-8")).
				Header("Accept", "application/json").
				Request()
			ctx := srv.NewContext(w, r)

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
			ctx := srv.NewContext(w, r)

			body, err := ctx.RequestBody()
			a.NotError(err).Equal(body, []byte(testdata.ObjectJSONString))
		}
	})
}

func BenchmarkContext_Unmarshal(b *testing.B) {
	a := assert.New(b, false)
	srv := newTestServer(a)

	b.Run("none", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			r := rest.Post(a, "/path", []byte(testdata.ObjectJSONString)).
				Header("Content-type", header.BuildContentType("application/json", "utf-8")).
				Header("Accept", "application/json").
				Request()
			ctx := srv.NewContext(w, r)

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
			ctx := srv.NewContext(w, r)

			obj := &testdata.Object{}
			a.NotError(ctx.Unmarshal(obj)).
				Equal(obj, testdata.ObjectInst)
		}
	})
}

// 一次普通的 POST 请求过程
func BenchmarkPost(b *testing.B) {
	a := assert.New(b, false)
	srv := newTestServer(a)

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r := rest.Post(a, "/path", []byte(testdata.ObjectJSONString)).
			Header("Content-type", header.BuildContentType("application/json", "utf-8")).
			Header("Accept", "application/json").
			Request()
		ctx := srv.NewContext(w, r)

		o := &testdata.Object{}
		a.NotError(ctx.Unmarshal(o)).
			Equal(o, testdata.ObjectInst)

		o.Age++
		o.Name = "response"
		Response(http.StatusCreated, o).Apply(ctx)
		a.Equal(w.Body.String(), `{"name":"response","Age":457}`)
	}
}

func BenchmarkPostWithCharset(b *testing.B) {
	a := assert.New(b, false)
	srv := newTestServer(a)

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r := rest.Post(a, "/path", testdata.ObjectGBKBytes).
			Header("Content-type", header.BuildContentType("application/json", "gbk")).
			Header("Accept", "application/json").
			Request()
		ctx := srv.NewContext(w, r)

		o := &testdata.Object{}
		a.NotError(ctx.Unmarshal(o)).
			Equal(o, testdata.ObjectInst)
	}
}

func BenchmarkContext_Object(b *testing.B) {
	a := assert.New(b, false)
	s := newTestServer(a)
	o := &testdata.Object{}

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r := rest.Post(a, "/path", nil).
			Header("Accept", "application/json").
			Header("content-type", "application/json").
			Request()
		ctx := s.NewContext(w, r)
		Response(http.StatusTeapot, o).Apply(ctx)
	}
}

func BenchmarkContext_Object_withHeader(b *testing.B) {
	a := assert.New(b, false)
	s := newTestServer(a)
	o := &testdata.Object{}

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r := rest.Post(a, "/path", nil).
			Header("Accept", "application/json").
			Header("content-type", "application/json").
			Request()
		ctx := s.NewContext(w, r)
		Response(http.StatusTeapot, o, "Location", "https://example.com").Apply(ctx)
	}
}

func BenchmarkNewRFC7807(b *testing.B) {
	for i := 0; i < b.N; i++ {
		p := newRFC7807().Init("id", "title", "detail", 400)
		p.WithExtensions(&object{Name: "n1", Age: 11})
		p.WithParam("p1", "v1")
		rfc7807Pool.Put(p)
	}
}

func BenchmarkRFC7807_unmarshal_json(b *testing.B) {
	a := assert.New(b, false)
	s := newTestServer(a)

	w := httptest.NewRecorder()
	r := rest.Post(a, "/path", nil).
		Header("Accept", "application/json").
		Header("content-type", "application/json").
		Request()
	ctx := s.NewContext(w, r)

	p := newRFC7807().Init("id", "title", "detail", 400)
	p.WithExtensions(&object{Name: "n1", Age: 11})
	p.WithParam("p1", "v1")
	for i := 0; i < b.N; i++ {
		p.Apply(ctx)
	}
}
