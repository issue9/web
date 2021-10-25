// SPDX-License-Identifier: MIT

package server

import (
	"bytes"
	"mime"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/web/internal/charsetdata"
	"github.com/issue9/web/serialization/text"
	"github.com/issue9/web/serialization/text/testobject"
)

func BenchmarkServer_NewContext(b *testing.B) {
	a := assert.New(b)
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
	a := assert.New(b)
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
	a := assert.New(b)
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
	a := assert.New(b)
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
	a := assert.New(b)
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
	a := assert.New(b)
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
	a := assert.New(b)
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
	a := assert.New(b)
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
	a := assert.New(b)
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
	a := assert.New(b)

	for i := 0; i < b.N; i++ {
		a.True(len(buildContentType(DefaultMimetype, DefaultCharset)) > 0)
	}
}
