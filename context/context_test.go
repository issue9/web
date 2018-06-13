// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package context

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
	xencoding "golang.org/x/text/encoding"
	"golang.org/x/text/encoding/simplifiedchinese"

	"github.com/issue9/web/encoding"
	"github.com/issue9/web/encoding/test"
)

func newContext(w http.ResponseWriter,
	r *http.Request,
	outputMimeType encoding.MarshalFunc,
	outputCharset xencoding.Encoding,
	inputMimeType encoding.UnmarshalFunc,
	InputCharset xencoding.Encoding) *Context {
	return &Context{
		Response:       w,
		Request:        r,
		OutputCharset:  outputCharset,
		OutputMimeType: outputMimeType,

		InputMimeType: inputMimeType,
		InputCharset:  InputCharset,
	}
}

func TestNew(t *testing.T) {
	a := assert.New(t)
	w := httptest.NewRecorder()

	// 错误的 accept
	r := httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Set("Accept", "not")
	a.Panic(func() {
		New(w, r)
	})

	// 错误的 accept-charset
	r = httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Set("Accept", encoding.DefaultMimeType)
	r.Header.Set("Accept-Charset", "unknown")
	a.Panic(func() {
		New(w, r)
	})

	// 错误的 content-type
	r = httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Set("Content-Type", ";charset=utf-8")
	a.Panic(func() {
		New(w, r)
	})

	// 正常
	r = httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Set("Accept", encoding.DefaultMimeType)
	ctx := New(w, r)
	a.NotNil(ctx)
}

func TestContext_Body(t *testing.T) {
	a := assert.New(t)
	r := httptest.NewRequest(http.MethodGet, "/path", bytes.NewBufferString("123"))
	r.Header.Set("Accept", "*/*")
	w := httptest.NewRecorder()
	ctx := newContext(w, r, encoding.TextMarshal, nil, encoding.TextUnmarshal, nil)

	// 未缓存
	a.Nil(ctx.body)
	data, err := ctx.Body()
	a.NotError(err).Equal(data, []byte("123"))
	a.Equal(ctx.body, data)

	// 读取缓存内容
	data, err = ctx.Body()
	a.NotError(err).Equal(data, []byte("123"))
	a.Equal(ctx.body, data)

	// 采用 Nop 即 utf-8 编码
	w.Body.Reset()
	r = httptest.NewRequest(http.MethodGet, "/path", bytes.NewBufferString("123"))
	r.Header.Set("Accept", "*/*")
	ctx = newContext(w, r, encoding.TextMarshal, xencoding.Nop, encoding.TextUnmarshal, xencoding.Nop)
	data, err = ctx.Body()
	a.NotError(err).Equal(data, []byte("123"))
	a.Equal(ctx.body, data)

	// 采用不同的编码
	w.Body.Reset()
	r = httptest.NewRequest(http.MethodGet, "/path", bytes.NewBuffer(gbkdata1))
	r.Header.Set("Accept", "*/*")
	ctx = newContext(w, r, encoding.TextMarshal, xencoding.Nop, encoding.TextUnmarshal, simplifiedchinese.GB18030)
	data, err = ctx.Body()
	a.NotError(err).Equal(string(data), gbkstr1)
	a.Equal(ctx.body, data)
}

func TestContext_Read(t *testing.T) {
	a := assert.New(t)
	r := httptest.NewRequest(http.MethodPost, "/path", bytes.NewBufferString("test,123"))
	w := httptest.NewRecorder()
	ctx := newContext(w, r, encoding.TextMarshal, nil, encoding.TextUnmarshal, nil)

	obj := &test.TextObject{}
	a.True(ctx.Read(obj))
	a.Equal(obj.Name, "test").Equal(obj.Age, 123)

	o := &struct{}{}
	a.False(ctx.Read(o))
}

func TestContext_Marshal(t *testing.T) {
	a := assert.New(t)

	r := httptest.NewRequest(http.MethodPost, "/path", nil)
	w := httptest.NewRecorder()
	ctx := newContext(w, r, encoding.TextMarshal, nil, encoding.TextUnmarshal, nil)
	obj := &test.TextObject{Name: "test", Age: 123}
	a.NotError(ctx.Marshal(http.StatusCreated, obj, map[string]string{"contEnt-type": "json"}))
	a.Equal(w.Code, http.StatusCreated)
	a.Equal(w.Body.String(), "test,123")
	a.Equal(w.Header().Get("content-type"), "json")

	r = httptest.NewRequest(http.MethodPost, "/path", nil)
	w = httptest.NewRecorder()
	ctx = newContext(w, r, encoding.TextMarshal, xencoding.Nop, encoding.TextUnmarshal, xencoding.Nop)
	obj = &test.TextObject{Name: "test", Age: 1234}
	a.NotError(ctx.Marshal(http.StatusCreated, obj, nil))
	a.Equal(w.Code, http.StatusCreated)
	a.Equal(w.Body.String(), "test,1234")

	// 输出 nil
	r = httptest.NewRequest(http.MethodPost, "/path", nil)
	w = httptest.NewRecorder()
	ctx = newContext(w, r, encoding.TextMarshal, xencoding.Nop, encoding.TextUnmarshal, xencoding.Nop)
	a.NotError(ctx.Marshal(http.StatusCreated, nil, nil))
	a.Equal(w.Code, http.StatusCreated)
	a.Equal(w.Body.String(), "")

	// 输出 Nil，encoding.Text 未实现对 nil 值的解析，所以采用了 json.Marshal
	r = httptest.NewRequest(http.MethodPost, "/path", nil)
	w = httptest.NewRecorder()
	ctx = newContext(w, r, json.Marshal, xencoding.Nop, encoding.TextUnmarshal, xencoding.Nop)
	a.NotError(ctx.Marshal(http.StatusCreated, encoding.Nil, nil))
	a.Equal(w.Code, http.StatusCreated)
	a.Equal(w.Body.String(), "null")

	// 输出不同编码的内容
	r = httptest.NewRequest(http.MethodPost, "/path", nil)
	w = httptest.NewRecorder()
	ctx = newContext(w, r, encoding.TextMarshal, simplifiedchinese.GB18030, encoding.TextUnmarshal, xencoding.Nop)
	a.NotError(ctx.Marshal(http.StatusCreated, gbkstr2, nil))
	a.Equal(w.Code, http.StatusCreated)
	a.Equal(w.Body.Bytes(), gbkdata2)
}

func TestContext_Render(t *testing.T) {
	a := assert.New(t)

	r := httptest.NewRequest(http.MethodPost, "/path", nil)
	w := httptest.NewRecorder()
	ctx := newContext(w, r, encoding.TextMarshal, nil, encoding.TextUnmarshal, nil)
	obj := &test.TextObject{Name: "test", Age: 123}
	ctx.Render(http.StatusCreated, obj, nil)
	a.Equal(w.Code, http.StatusCreated)
	a.Equal(w.Body.String(), "test,123")

	r = httptest.NewRequest(http.MethodPost, "/path", nil)
	w = httptest.NewRecorder()
	ctx = newContext(w, r, encoding.TextMarshal, nil, encoding.TextUnmarshal, nil)
	obj1 := &struct{ Name string }{Name: "name"}
	ctx.Render(http.StatusCreated, obj1, nil)
	a.Equal(w.Code, http.StatusInternalServerError)
}

func TestContext_ClientIP(t *testing.T) {
	a := assert.New(t)
	w := httptest.NewRecorder()

	r := httptest.NewRequest(http.MethodPost, "/path", nil)
	ctx := newContext(w, r, encoding.TextMarshal, nil, encoding.TextUnmarshal, nil)
	a.Equal(ctx.ClientIP(), r.RemoteAddr)

	// httptest.NewRequest 会直接将  remote-addr 赋值为 192.0.2.1 无法测试
	r, err := http.NewRequest(http.MethodPost, "/path", nil)
	a.NotError(err).NotNil(r)
	r.Header.Set("x-real-ip", "192.168.1.1:8080")
	ctx = newContext(w, r, encoding.TextMarshal, nil, encoding.TextUnmarshal, nil)
	a.Equal(ctx.ClientIP(), "192.168.1.1:8080")

	r = httptest.NewRequest(http.MethodPost, "/path", nil)
	r.Header.Set("x-forwarded-for", "192.168.2.1:8080,192.168.2.2:111")
	ctx = newContext(w, r, encoding.TextMarshal, nil, encoding.TextUnmarshal, nil)
	a.Equal(ctx.ClientIP(), "192.168.2.1:8080")
}
