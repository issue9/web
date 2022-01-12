// SPDX-License-Identifier: MIT

package server

import (
	"bytes"
	"compress/flate"
	"io"
	"mime"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/issue9/assert/v2"
	"github.com/issue9/mux/v6/group"
	"github.com/issue9/mux/v6/routertest"

	"github.com/issue9/web/internal/charsetdata"
	"github.com/issue9/web/serialization/text"
	"github.com/issue9/web/serialization/text/testobject"
)

func BenchmarkRouter(b *testing.B) {
	a := assert.New(b, false)
	srv := newServer(a, &Options{Port: ":8080"})

	h := func(c *Context) Responser {
		_, err := c.Response.Write([]byte(c.Request.URL.Path))
		if err != nil {
			b.Error(err)
		}
		return nil
	}

	t := routertest.NewTester[HandlerFunc](srv.call)
	t.Bench(b, h)
}

func BenchmarkServer_Serve(b *testing.B) {
	a := assert.New(b, false)
	srv := newServer(a, &Options{Port: ":8080"})
	router := srv.NewRouter("srv", "http://localhost:8080/", group.MatcherFunc(group.Any))
	a.NotNil(router)

	router.Get("/path", func(c *Context) Responser {
		return Object(http.StatusOK, "/path", map[string]string{"h1": "h1"})
	})
	go func() {
		srv.Serve()
	}()
	time.Sleep(500 * time.Millisecond)

	for i := 0; i < b.N; i++ {
		r, err := http.NewRequest(http.MethodGet, "http://localhost:8080/path", nil)
		a.NotError(err).NotNil(r)
		r.Header.Set("Content-type", mime.FormatMediaType(text.Mimetype, map[string]string{"charset": "gbk"}))
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

func BenchmarkServer_NewContext(b *testing.B) {
	a := assert.New(b, false)
	srv := newServer(a, nil)

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodGet, "/path", nil)
		a.NotError(err).NotNil(r)
		r.Header.Set("Content-type", mime.FormatMediaType(text.Mimetype, map[string]string{"charset": "gbk"}))
		r.Header.Set("Accept", text.Mimetype)
		r.Header.Set("Accept-Charset", "gbk;q=1,gb18080;q=0.1")

		ctx := srv.NewContext(w, r)
		a.NotNil(ctx)
	}
}

func BenchmarkContext_Marshal(b *testing.B) {
	a := assert.New(b, false)
	srv := newServer(a, nil)

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodGet, "/path", nil)
		a.NotError(err).NotNil(r)
		r.Header.Set("Accept", text.Mimetype)
		ctx := srv.NewContext(w, r)

		obj := &testobject.TextObject{Age: 22, Name: "中文2"}
		a.NotError(ctx.Marshal(http.StatusCreated, obj, nil))
		a.Equal(w.Body.Bytes(), charsetdata.GBKString2)
	}
}

func BenchmarkContext_MarshalWithUTF8(b *testing.B) {
	a := assert.New(b, false)
	srv := newServer(a, nil)

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodGet, "/path", nil)
		a.NotError(err).NotNil(r)
		r.Header.Set("Accept", text.Mimetype)
		r.Header.Set("Accept-Charset", "utf-8")
		ctx := srv.NewContext(w, r)

		obj := &testobject.TextObject{Age: 22, Name: "中文2"}
		a.NotError(ctx.Marshal(http.StatusCreated, obj, nil))
		a.Equal(w.Body.Bytes(), charsetdata.GBKString2)
	}
}

func BenchmarkContext_MarshalWithCharset(b *testing.B) {
	a := assert.New(b, false)
	srv := newServer(a, nil)

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodGet, "/path", nil)
		a.NotError(err).NotNil(r)
		r.Header.Set("Accept", text.Mimetype)
		r.Header.Set("Accept-Charset", "gbk;q=1,gb18080;q=0.1")
		ctx := srv.NewContext(w, r)

		obj := &testobject.TextObject{Age: 22, Name: "中文2"}
		a.NotError(ctx.Marshal(http.StatusCreated, obj, nil))
		a.Equal(w.Body.Bytes(), charsetdata.GBKData2)
	}
}

func BenchmarkContext_MarshalWithCharsetEncoding(b *testing.B) {
	a := assert.New(b, false)
	srv := newServer(a, nil)

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodGet, "/path", nil)
		a.NotError(err).NotNil(r)
		r.Header.Set("Accept", text.Mimetype)
		r.Header.Set("Accept-Charset", "gbk;q=1,gb18080;q=0.1")
		r.Header.Set("Accept-Encoding", "gzip;q=0.9,deflate")

		ctx := srv.NewContext(w, r)
		obj := &testobject.TextObject{Age: 22, Name: "中文2"}
		a.NotError(ctx.Marshal(http.StatusCreated, obj, nil))
		ctx.destory()

		data, err := io.ReadAll(flate.NewReader(w.Body))
		a.NotError(err).NotNil(data)
		a.Equal(data, charsetdata.GBKData2)
	}
}

func BenchmarkContext_Unmarshal(b *testing.B) {
	a := assert.New(b, false)
	srv := newServer(a, nil)

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodPost, "/path", bytes.NewBufferString("request,15"))
		a.NotError(err).NotNil(r)
		r.Header.Set("Content-type", mime.FormatMediaType(text.Mimetype, map[string]string{"charset": "utf-8"}))
		r.Header.Set("Accept", text.Mimetype)
		ctx := srv.NewContext(w, r)

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
		r, err := http.NewRequest(http.MethodGet, "/path", bytes.NewBufferString(charsetdata.GBKString1))
		a.NotError(err).NotNil(r)
		r.Header.Set("Content-type", mime.FormatMediaType(text.Mimetype, map[string]string{"charset": "utf-8"}))
		r.Header.Set("Accept", text.Mimetype)
		ctx := srv.NewContext(w, r)

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
		r, err := http.NewRequest(http.MethodGet, "/path", bytes.NewBuffer(charsetdata.GBKData1))
		a.NotError(err).NotNil(r)
		r.Header.Set("Content-type", mime.FormatMediaType(text.Mimetype, map[string]string{"charset": "gbk"}))
		r.Header.Set("Accept", text.Mimetype)
		r.Header.Set("Accept-Charset", "gbk")
		ctx := srv.NewContext(w, r)

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
		r.Header.Set("Content-type", mime.FormatMediaType(text.Mimetype, map[string]string{"charset": "utf-8"}))
		r.Header.Set("Accept", text.Mimetype)
		ctx := srv.NewContext(w, r)

		obj := &testobject.TextObject{}
		a.NotError(ctx.Unmarshal(obj))
		a.Equal(obj.Age, 15).
			Equal(obj.Name, "request")

		obj.Age++
		obj.Name = "response"
		err = ctx.Marshal(http.StatusCreated, obj, nil)
		a.NotError(err).Equal(w.Body.String(), "response,16")
	}
}

func BenchmarkPostWithCharset(b *testing.B) {
	a := assert.New(b, false)
	srv := newServer(a, nil)

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodPost, "/path", bytes.NewBuffer(charsetdata.GBKData1))
		a.NotError(err).NotNil(r)
		r.Header.Set("Content-type", mime.FormatMediaType(text.Mimetype, map[string]string{"charset": "gbk"}))
		r.Header.Set("Accept", text.Mimetype)
		r.Header.Set("Accept-Charset", "gbk;q=1,gb18080;q=0.1")
		ctx := srv.NewContext(w, r)

		obj := &testobject.TextObject{}
		a.NotError(ctx.Unmarshal(obj))
		a.Equal(obj.Age, 11).Equal(obj.Name, "中文1")

		obj.Age = 22
		obj.Name = "中文2"
		a.NotError(ctx.Marshal(http.StatusCreated, obj, nil))
		a.Equal(w.Body.Bytes(), charsetdata.GBKData2)
	}
}

func BenchmarkBuildContentType(b *testing.B) {
	a := assert.New(b, false)

	for i := 0; i < b.N; i++ {
		a.True(len(buildContentType(DefaultMimetype, DefaultCharset)) > 0)
	}
}
