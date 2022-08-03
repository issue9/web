// SPDX-License-Identifier: MIT

package server

import (
	"bytes"
	"compress/flate"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/issue9/assert/v2"
	"github.com/issue9/assert/v2/rest"
	"github.com/issue9/mux/v7"
	"github.com/issue9/mux/v7/routertest"

	"github.com/issue9/web/serializer/text"
	"github.com/issue9/web/serializer/text/testobject"
)

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
		return Object(http.StatusOK, "/path", "h1", "h1")
	})
	go func() {
		srv.Serve()
	}()
	time.Sleep(500 * time.Millisecond)

	for i := 0; i < b.N; i++ {
		r, err := http.NewRequest(http.MethodGet, "http://localhost:8080/path", nil)
		a.NotError(err).NotNil(r)
		r.Header.Set("Content-type", buildContentType(text.Mimetype, "gbk"))
		r.Header.Set("accept", text.Mimetype)
		r.Header.Set("accept-charset", "gbk;q=1,gb18080;q=0.1")
		resp, err := http.DefaultClient.Do(r)
		a.NotError(err).NotNil(resp)
		a.Equal(resp.Header.Get("h1"), "h1")
		body, err := io.ReadAll(resp.Body)
		a.NotError(err).Equal(string(body), "/path")
	}

	srv.Close(0)
}

func BenchmarkServer_newContext(b *testing.B) {
	a := assert.New(b, false)
	srv := newServer(a, nil)

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodGet, "/path", nil)
		a.NotError(err).NotNil(r)
		r.Header.Set("Content-type", buildContentType(text.Mimetype, "gbk"))
		r.Header.Set("Accept", text.Mimetype)
		r.Header.Set("Accept-Charset", "gbk;q=1,gb18080;q=0.1")

		ctx := srv.newContext(w, r, nil)
		a.NotNil(ctx)
	}
}

func BenchmarkContext_render(b *testing.B) {
	a := assert.New(b, false)
	srv := newServer(a, nil)

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodGet, "/path", nil)
		a.NotError(err).NotNil(r)
		r.Header.Set("Accept", text.Mimetype)
		ctx := srv.newContext(w, r, nil)

		o := &testobject.TextObject{Age: 22, Name: "中文2"}
		Object(http.StatusCreated, o).Apply(ctx)
		a.Equal(w.Body.Bytes(), gbkString2)
	}
}

func BenchmarkContext_renderWithUTF8(b *testing.B) {
	a := assert.New(b, false)
	srv := newServer(a, nil)

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodGet, "/path", nil)
		a.NotError(err).NotNil(r)
		r.Header.Set("Accept", text.Mimetype)
		r.Header.Set("Accept-Charset", "utf-8")
		ctx := srv.newContext(w, r, nil)

		o := &testobject.TextObject{Age: 22, Name: "中文2"}
		Object(http.StatusCreated, o).Apply(ctx)
		a.Equal(w.Body.Bytes(), gbkString2)
	}
}

func BenchmarkContext_renderWithCharset(b *testing.B) {
	a := assert.New(b, false)
	srv := newServer(a, nil)

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodGet, "/path", nil)
		a.NotError(err).NotNil(r)
		r.Header.Set("Accept", text.Mimetype)
		r.Header.Set("Accept-Charset", "gbk;q=1,gb18080;q=0.1")
		ctx := srv.newContext(w, r, nil)

		o := &testobject.TextObject{Age: 22, Name: "中文2"}
		Object(http.StatusCreated, o).Apply(ctx)
		a.Equal(w.Body.Bytes(), gbkBytes2)
	}
}

func BenchmarkContext_renderWithCharsetEncoding(b *testing.B) {
	a := assert.New(b, false)
	srv := newServer(a, nil)

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodGet, "/path", nil)
		a.NotError(err).NotNil(r)
		r.Header.Set("Accept", text.Mimetype)
		r.Header.Set("Accept-Charset", "gbk;q=1,gb18080;q=0.1")
		r.Header.Set("Accept-Encoding", "gzip;q=0.9,deflate")

		ctx := srv.newContext(w, r, nil)
		o := &testobject.TextObject{Age: 22, Name: "中文2"}
		Object(http.StatusCreated, o).Apply(ctx)
		ctx.destroy()

		data, err := io.ReadAll(flate.NewReader(w.Body))
		a.NotError(err).NotNil(data)
		a.Equal(data, gbkBytes2)
	}
}

func BenchmarkContext_Body(b *testing.B) {
	a := assert.New(b, false)
	srv := newServer(a, nil)

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodPost, "/path", bytes.NewBufferString("request,15"))
		a.NotError(err).NotNil(r)
		r.Header.Set("Content-type", buildContentType(text.Mimetype, "utf-8"))
		r.Header.Set("Accept", text.Mimetype)
		ctx := srv.newContext(w, r, nil)

		body, err := ctx.Body()
		a.NotError(err).Equal(body, []byte("request,15"))
	}
}

func BenchmarkContext_BodyWithCharset(b *testing.B) {
	a := assert.New(b, false)
	srv := newServer(a, nil)

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodGet, "/path", bytes.NewBuffer(gbkBytes1))
		a.NotError(err).NotNil(r)
		r.Header.Set("Content-type", buildContentType(text.Mimetype, "gbk"))
		r.Header.Set("Accept", text.Mimetype)
		r.Header.Set("Accept-Charset", "gbk")
		ctx := srv.newContext(w, r, nil)

		body, err := ctx.Body()
		a.NotError(err).Equal(body, []byte(gbkString1))
	}
}

func BenchmarkContext_Unmarshal(b *testing.B) {
	a := assert.New(b, false)
	srv := newServer(a, nil)

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodPost, "/path", bytes.NewBufferString("request,15"))
		a.NotError(err).NotNil(r)
		r.Header.Set("Content-type", buildContentType(text.Mimetype, "utf-8"))
		r.Header.Set("Accept", text.Mimetype)
		ctx := srv.newContext(w, r, nil)

		obj := &testobject.TextObject{}
		a.NotError(ctx.Unmarshal(obj))
		a.Equal(obj.Age, 15).
			Equal(obj.Name, "request")
	}
}

func BenchmarkContext_UnmarshalWithUTF8(b *testing.B) {
	a := assert.New(b, false)
	srv := newServer(a, nil)

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodGet, "/path", bytes.NewBufferString(gbkString1))
		a.NotError(err).NotNil(r)
		r.Header.Set("Content-type", buildContentType(text.Mimetype, "utf-8"))
		r.Header.Set("Accept", text.Mimetype)
		ctx := srv.newContext(w, r, nil)

		obj := &testobject.TextObject{}
		a.NotError(ctx.Unmarshal(obj))
		a.Equal(obj.Age, 11).Equal(obj.Name, "中文1")
	}
}

func BenchmarkContext_UnmarshalWithCharset(b *testing.B) {
	a := assert.New(b, false)
	srv := newServer(a, nil)

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodGet, "/path", bytes.NewBuffer(gbkBytes1))
		a.NotError(err).NotNil(r)
		r.Header.Set("Content-type", buildContentType(text.Mimetype, "gbk"))
		r.Header.Set("Accept", text.Mimetype)
		r.Header.Set("Accept-Charset", "gbk")
		ctx := srv.newContext(w, r, nil)

		obj := &testobject.TextObject{}
		a.NotError(ctx.Unmarshal(obj))
		a.Equal(obj.Age, 11)
	}
}

// 一次普通的 POST 请求过程
func BenchmarkPost(b *testing.B) {
	a := assert.New(b, false)
	srv := newServer(a, nil)

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodPost, "/path", bytes.NewBufferString("request,15"))
		a.NotError(err).NotNil(r)
		r.Header.Set("Content-type", buildContentType(text.Mimetype, "utf-8"))
		r.Header.Set("Accept", text.Mimetype)
		ctx := srv.newContext(w, r, nil)

		o := &testobject.TextObject{}
		a.NotError(ctx.Unmarshal(o))
		a.Equal(o.Age, 15).
			Equal(o.Name, "request")

		o.Age++
		o.Name = "response"
		Object(http.StatusCreated, o).Apply(ctx)
		a.Equal(w.Body.String(), "response,16")
	}
}

func BenchmarkPostWithCharset(b *testing.B) {
	a := assert.New(b, false)
	srv := newServer(a, nil)

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodPost, "/path", bytes.NewBuffer(gbkBytes1))
		a.NotError(err).NotNil(r)
		r.Header.Set("Content-type", buildContentType(text.Mimetype, "gbk"))
		r.Header.Set("Accept", text.Mimetype)
		r.Header.Set("Accept-Charset", "gbk;q=1,gb18080;q=0.1")
		ctx := srv.newContext(w, r, nil)

		o := &testobject.TextObject{}
		a.NotError(ctx.Unmarshal(o))
		a.Equal(o.Age, 11).Equal(o.Name, "中文1")

		o.Age = 22
		o.Name = "中文2"
		Object(http.StatusCreated, o).Apply(ctx)
		a.Equal(w.Body.Bytes(), gbkBytes2)
	}
}

func BenchmarkBuildContentType(b *testing.B) {
	a := assert.New(b, false)

	for i := 0; i < b.N; i++ {
		a.True(len(buildContentType(DefaultMimetype, DefaultCharset)) > 0)
	}
}

func BenchmarkContext_Object(b *testing.B) {
	a := assert.New(b, false)
	s := newServer(a, nil)
	o := &testobject.TextObject{}

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r := rest.Post(a, "/path", nil).
			Header("Accept", text.Mimetype).
			Header("content-type", text.Mimetype).
			Request()
		ctx := s.newContext(w, r, nil)
		Object(http.StatusTeapot, o).Apply(ctx)
	}
}

func BenchmarkContext_Object_withHeader(b *testing.B) {
	a := assert.New(b, false)
	s := newServer(a, nil)
	o := &testobject.TextObject{}

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r := rest.Post(a, "/path", nil).
			Header("Accept", text.Mimetype).
			Header("content-type", text.Mimetype).
			Request()
		ctx := s.newContext(w, r, nil)
		Object(http.StatusTeapot, o, "Location", "https://example.com").Apply(ctx)
	}
}
