// SPDX-License-Identifier: MIT

package server

import (
	"bytes"
	"mime"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/issue9/assert/v2"
	"github.com/issue9/assert/v2/rest"
	"github.com/issue9/mux/v5/group"

	"github.com/issue9/web/internal/charsetdata"
	"github.com/issue9/web/serialization/text"
	"github.com/issue9/web/serialization/text/testobject"
)

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
		rest.NewRequest(a, nil, http.MethodGet, "http://localhost:8080/path").
			Header("Content-type", mime.FormatMediaType(text.Mimetype, map[string]string{"charset": "gbk"})).
			Header("accept", text.Mimetype).
			Header("accept-charset", "gbk;q=1,gb18080;q=0.1").
			Do().
			Header("h1", "h1").
			StringBody("/path")
	}

	srv.Close(0)
}

func BenchmarkServer_NewContext(b *testing.B) {
	a := assert.New(b, false)
	srv := newServer(a, nil)

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/path", nil)
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
		r := httptest.NewRequest(http.MethodGet, "/path", nil)
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
		r := httptest.NewRequest(http.MethodGet, "/path", nil)
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
		r := httptest.NewRequest(http.MethodGet, "/path", nil)
		r.Header.Set("Accept", text.Mimetype)
		r.Header.Set("Accept-Charset", "gbk;q=1,gb18080;q=0.1")
		ctx := srv.NewContext(w, r)

		obj := &testobject.TextObject{Age: 22, Name: "中文2"}
		a.NotError(ctx.Marshal(http.StatusCreated, obj, nil))
		a.Equal(w.Body.Bytes(), charsetdata.GBKData2)
	}
}

func BenchmarkContext_Unmarshal(b *testing.B) {
	a := assert.New(b, false)
	srv := newServer(a, nil)

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/path", bytes.NewBufferString("request,15"))
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
		r := httptest.NewRequest(http.MethodGet, "/path", bytes.NewBufferString(charsetdata.GBKString1))
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
		r := httptest.NewRequest(http.MethodGet, "/path", bytes.NewBuffer(charsetdata.GBKData1))
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
		r := httptest.NewRequest(http.MethodPost, "/path", bytes.NewBufferString("request,15"))
		r.Header.Set("Content-type", mime.FormatMediaType(text.Mimetype, map[string]string{"charset": "utf-8"}))
		r.Header.Set("Accept", text.Mimetype)
		ctx := srv.NewContext(w, r)

		obj := &testobject.TextObject{}
		a.NotError(ctx.Unmarshal(obj))
		a.Equal(obj.Age, 15).
			Equal(obj.Name, "request")

		obj.Age++
		obj.Name = "response"
		err := ctx.Marshal(http.StatusCreated, obj, nil)
		a.NotError(err).Equal(w.Body.String(), "response,16")
	}
}

func BenchmarkPostWithCharset(b *testing.B) {
	a := assert.New(b, false)
	srv := newServer(a, nil)

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/path", bytes.NewBuffer(charsetdata.GBKData1))
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
