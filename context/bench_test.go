// SPDX-License-Identifier: MIT

package context

import (
	"bytes"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/issue9/assert"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"

	"github.com/issue9/web/context/mimetype"
	"github.com/issue9/web/context/mimetype/mimetypetest"
)

var (
	gbkstr1            = "中文1,11"
	gbkstr2            = "中文2,22"
	gbkdata1, gbkdata2 []byte
)

func init() {
	reader := transform.NewReader(strings.NewReader(gbkstr1), simplifiedchinese.GBK.NewEncoder())
	gbkdata, err := ioutil.ReadAll(reader)
	if err != nil {
		panic(err)
	}
	gbkdata1 = gbkdata

	reader = transform.NewReader(strings.NewReader(gbkstr2), simplifiedchinese.GBK.NewEncoder())
	gbkdata, err = ioutil.ReadAll(reader)
	if err != nil {
		panic(err)
	}
	gbkdata2 = gbkdata
}

func BenchmarkParseContentType(b *testing.B) {
	a := assert.New(b)

	for i := 0; i < b.N; i++ {
		_, _, err := parseContentType("application/json;charset=utf-8")
		a.NotError(err)
	}
}

func BenchmarkBuildContentType(b *testing.B) {
	a := assert.New(b)

	for i := 0; i < b.N; i++ {
		a.True(len(buildContentType(mimetype.DefaultMimetype, utfName)) > 0)
	}
}

func BenchmarkBuilder_New(b *testing.B) {
	a := assert.New(b)
	bb := newBuilder(a)

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/path", nil)
		r.Header.Set("Content-type", buildContentType(mimetypetest.Mimetype, "gbk"))
		r.Header.Set("Accept", mimetypetest.Mimetype)
		r.Header.Set("Accept-Charset", "gbk;q=1,gb18080;q=0.1")

		ctx := bb.New(w, r)
		a.NotNil(ctx)
	}
}

func BenchmarkContext_Marshal(b *testing.B) {
	a := assert.New(b)
	bb := newBuilder(a)

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/path", nil)
		r.Header.Set("Accept", mimetypetest.Mimetype)
		ctx := bb.New(w, r)

		obj := &mimetypetest.TextObject{Age: 22, Name: "中文2"}
		a.NotError(ctx.Marshal(http.StatusCreated, obj, nil))
		a.Equal(w.Body.Bytes(), gbkstr2)
	}
}

func BenchmarkContext_MarshalWithUTF8(b *testing.B) {
	a := assert.New(b)
	bb := newBuilder(a)

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/path", nil)
		r.Header.Set("Accept", mimetypetest.Mimetype)
		r.Header.Set("Accept-Charset", "utf-8")
		ctx := bb.New(w, r)

		obj := &mimetypetest.TextObject{Age: 22, Name: "中文2"}
		a.NotError(ctx.Marshal(http.StatusCreated, obj, nil))
		a.Equal(w.Body.Bytes(), gbkstr2)
	}
}

func BenchmarkContext_MarshalWithCharset(b *testing.B) {
	a := assert.New(b)
	bb := newBuilder(a)

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/path", nil)
		r.Header.Set("Accept", mimetypetest.Mimetype)
		r.Header.Set("Accept-Charset", "gbk;q=1,gb18080;q=0.1")
		ctx := bb.New(w, r)

		obj := &mimetypetest.TextObject{Age: 22, Name: "中文2"}
		a.NotError(ctx.Marshal(http.StatusCreated, obj, nil))
		a.Equal(w.Body.Bytes(), gbkdata2)
	}
}

func BenchmarkContext_Unmarshal(b *testing.B) {
	a := assert.New(b)
	bb := newBuilder(a)

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/path", bytes.NewBufferString("request,15"))
		r.Header.Set("Content-type", buildContentType(mimetypetest.Mimetype, "utf-8"))
		r.Header.Set("Accept", mimetypetest.Mimetype)
		ctx := bb.New(w, r)

		obj := &mimetypetest.TextObject{}
		a.NotError(ctx.Unmarshal(obj))
		a.Equal(obj.Age, 15).
			Equal(obj.Name, "request")
	}
}

func BenchmarkContext_UnmarshalWithUTF8(b *testing.B) {
	a := assert.New(b)
	bb := newBuilder(a)

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/path", bytes.NewBufferString(gbkstr1))
		r.Header.Set("Content-type", buildContentType(mimetypetest.Mimetype, "utf-8"))
		r.Header.Set("Accept", mimetypetest.Mimetype)
		ctx := bb.New(w, r)

		obj := &mimetypetest.TextObject{}
		a.NotError(ctx.Unmarshal(obj))
		a.Equal(obj.Age, 11).Equal(obj.Name, "中文1")
	}
}

func BenchmarkContext_UnmarshalWithCharset(b *testing.B) {
	a := assert.New(b)
	bb := newBuilder(a)

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/path", bytes.NewBuffer(gbkdata1))
		r.Header.Set("Content-type", buildContentType(mimetypetest.Mimetype, "gbk"))
		r.Header.Set("Accept", mimetypetest.Mimetype)
		r.Header.Set("Accept-Charset", "gbk")
		ctx := bb.New(w, r)

		obj := &mimetypetest.TextObject{}
		a.NotError(ctx.Unmarshal(obj))
		a.Equal(obj.Age, 11)
	}
}

// 一次普通的 POST 请求过程
func BenchmarkPost(b *testing.B) {
	a := assert.New(b)
	bb := newBuilder(a)

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/path", bytes.NewBufferString("request,15"))
		r.Header.Set("Content-type", buildContentType(mimetypetest.Mimetype, "utf-8"))
		r.Header.Set("Accept", mimetypetest.Mimetype)
		ctx := bb.New(w, r)

		obj := &mimetypetest.TextObject{}
		a.NotError(ctx.Unmarshal(obj))
		a.Equal(obj.Age, 15).
			Equal(obj.Name, "request")

		obj.Age++
		obj.Name = "response"
		ctx.Render(http.StatusCreated, obj, nil)
		a.Equal(w.Body.String(), "response,16")
	}
}

func BenchmarkPostWithCharset(b *testing.B) {
	a := assert.New(b)
	bb := newBuilder(a)

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/path", bytes.NewBuffer(gbkdata1))
		r.Header.Set("Content-type", buildContentType(mimetypetest.Mimetype, "gbk"))
		r.Header.Set("Accept", mimetypetest.Mimetype)
		r.Header.Set("Accept-Charset", "gbk;q=1,gb18080;q=0.1")
		ctx := bb.New(w, r)

		obj := &mimetypetest.TextObject{}
		a.NotError(ctx.Unmarshal(obj))
		a.Equal(obj.Age, 11).Equal(obj.Name, "中文1")

		obj.Age = 22
		obj.Name = "中文2"
		a.NotError(ctx.Marshal(http.StatusCreated, obj, nil))
		a.Equal(w.Body.Bytes(), gbkdata2)
	}
}

func BenchmarkBuilder_Marshal(b *testing.B) {
	a := assert.New(b)
	builder := newBuilder(a)
	a.NotNil(builder)

	a.NotError(builder.AddMarshal("font/wottf", xml.Marshal))

	for i := 0; i < b.N; i++ {
		name, marshal, err := builder.Marshal("font/wottf;q=0.9")
		a.NotError(err).
			NotEmpty(name).
			NotNil(marshal)
	}
}

func BenchmarkBuilder_Unmarshal(b *testing.B) {
	a := assert.New(b)
	builder := newBuilder(a)
	a.NotNil(builder)

	a.NotError(builder.AddUnmarshal("font/wottf", xml.Unmarshal))

	for i := 0; i < b.N; i++ {
		marshal, err := builder.Unmarshal("font/wottf")
		a.NotError(err).
			NotNil(marshal)
	}
}
