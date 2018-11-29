// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package context

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/issue9/web/mimetype"
	"github.com/issue9/web/mimetype/mimetypetest"
)

func init() {
	message.SetString(language.Chinese, "test", "中文")
	message.SetString(language.SimplifiedChinese, "test", "简体")
	message.SetString(language.TraditionalChinese, "test", "繁体")
	message.SetString(language.English, "test", "english")
}

func newContext(w http.ResponseWriter,
	r *http.Request,
	outputMimeType mimetype.MarshalFunc,
	outputCharset encoding.Encoding,
	inputMimeType mimetype.UnmarshalFunc,
	InputCharset encoding.Encoding) *Context {
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
	mt := mimetype.New()
	mt.AddMarshal(mimetype.DefaultMimetype, mimetypetest.TextMarshal)
	mt.AddUnmarshal(mimetype.DefaultMimetype, mimetypetest.TextUnmarshal)
	logwriter := new(bytes.Buffer)
	errlog := log.New(logwriter, "", 0)

	// 错误的 accept
	logwriter.Reset()
	r := httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Set("Accept", "not")
	a.Panic(func() {
		New(w, r, mt, nil)
	})
	a.Equal(logwriter.Len(), 0)

	// 错误的 accept-charset
	logwriter.Reset()
	r = httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Set("Accept", mimetype.DefaultMimetype)
	r.Header.Set("Accept-Charset", "unknown")
	a.Panic(func() {
		New(w, r, mt, errlog)
	})
	a.True(logwriter.Len() > 0)

	// 错误的 content-type,无输入内容
	logwriter.Reset()
	r = httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Set("Content-Type", ";charset=utf-8")
	a.Panic(func() {
		New(w, r, mt, errlog)
	})

	// 错误的 content-type,有输入内容
	logwriter.Reset()
	r = httptest.NewRequest(http.MethodPost, "/path", bytes.NewBufferString("[]"))
	r.Header.Set("Content-Type", ";charset=utf-8")
	a.Panic(func() {
		New(w, r, mt, errlog)
	})
	a.True(logwriter.Len() > 0)

	// 错误的 content-type，且有输入内容
	logwriter.Reset()
	r = httptest.NewRequest(http.MethodGet, "/path", bytes.NewBufferString("123"))
	r.Header.Set("Accept", mimetype.DefaultMimetype)
	r.Header.Set("content-type", buildContentType(mimetypetest.MimeType, "utf-"))
	a.Panic(func() {
		New(w, r, mt, errlog)
	})

	// 错误的 Accept-Language
	logwriter.Reset()
	r = httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Set("Accept", mimetype.DefaultMimetype)
	r.Header.Set("Accept-Language", "zh-hans;q=0.9,zh-Hant;q=xxx")
	a.Panic(func() {
		New(w, r, mt, errlog)
	})

	// 正常，指定 Accept-Language
	logwriter.Reset()
	r = httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Set("Accept", mimetype.DefaultMimetype)
	r.Header.Set("Accept-Language", "zh-hans;q=0.9,zh-Hant;q=0.7")
	var ctx *Context
	a.NotPanic(func() {
		ctx = New(w, r, mt, errlog)
	})
	a.NotNil(ctx).
		Equal(logwriter.Len(), 0).
		Equal(ctx.InputCharset, nil).
		Equal(ctx.OutputMimeTypeName, mimetype.DefaultMimetype).
		Equal(ctx.OutputTag, language.SimplifiedChinese).
		NotNil(ctx.LocalePrinter)

	// 正常，未指定 Accept-Language 和 Accept-Charset 等不是必须的报头
	logwriter.Reset()
	r = httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Set("Accept", mimetype.DefaultMimetype)
	a.NotPanic(func() {
		ctx = New(w, r, mt, errlog)
	})
	a.NotNil(ctx).
		Equal(logwriter.Len(), 0).
		Equal(ctx.InputCharset, nil).
		Equal(ctx.OutputMimeTypeName, mimetype.DefaultMimetype)

	// 正常，未指定 Accept-Language 和 Accept-Charset 等不是必须的报头，且有输入内容
	a.NotError(mt.AddUnmarshal(mimetypetest.MimeType, mimetypetest.TextUnmarshal))
	logwriter.Reset()
	r = httptest.NewRequest(http.MethodGet, "/path", bytes.NewBufferString("123"))
	r.Header.Set("Accept", mimetype.DefaultMimetype)
	r.Header.Set("content-type", buildContentType(mimetypetest.MimeType, "utf-8"))
	a.NotPanic(func() {
		ctx = New(w, r, mt, errlog)
	})
	a.NotNil(ctx).
		Equal(logwriter.Len(), 0).
		True(charsetIsNop(ctx.InputCharset)).
		Equal(ctx.OutputMimeTypeName, mimetype.DefaultMimetype)
}

func TestContext_Body(t *testing.T) {
	a := assert.New(t)
	r := httptest.NewRequest(http.MethodGet, "/path", bytes.NewBufferString("123"))
	r.Header.Set("Accept", "*/*")
	w := httptest.NewRecorder()
	ctx := newContext(w, r, mimetypetest.TextMarshal, nil, mimetypetest.TextUnmarshal, nil)

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
	ctx = newContext(w, r, mimetypetest.TextMarshal, encoding.Nop, mimetypetest.TextUnmarshal, encoding.Nop)
	data, err = ctx.Body()
	a.NotError(err).Equal(data, []byte("123"))
	a.Equal(ctx.body, data)

	// 采用不同的编码
	w.Body.Reset()
	r = httptest.NewRequest(http.MethodGet, "/path", bytes.NewBuffer(gbkdata1))
	r.Header.Set("Accept", "*/*")
	ctx = newContext(w, r, mimetypetest.TextMarshal, encoding.Nop, mimetypetest.TextUnmarshal, simplifiedchinese.GB18030)
	data, err = ctx.Body()
	a.NotError(err).Equal(string(data), gbkstr1)
	a.Equal(ctx.body, data)
}

func TestContext_Read(t *testing.T) {
	a := assert.New(t)
	r := httptest.NewRequest(http.MethodPost, "/path", bytes.NewBufferString("test,123"))
	w := httptest.NewRecorder()
	ctx := newContext(w, r, mimetypetest.TextMarshal, nil, mimetypetest.TextUnmarshal, nil)

	obj := &mimetypetest.TextObject{}
	a.True(ctx.Read(obj))
	a.Equal(obj.Name, "test").Equal(obj.Age, 123)

	o := &struct{}{}
	a.False(ctx.Read(o))
}

func TestContext_Marshal(t *testing.T) {
	a := assert.New(t)

	r := httptest.NewRequest(http.MethodPost, "/path", nil)
	w := httptest.NewRecorder()
	ctx := newContext(w, r, mimetypetest.TextMarshal, nil, mimetypetest.TextUnmarshal, nil)
	obj := &mimetypetest.TextObject{Name: "test", Age: 123}
	a.NotError(ctx.Marshal(http.StatusCreated, obj, map[string]string{"contEnt-type": "json", "content-lanGuage": "zh-hans"}))
	a.Equal(w.Code, http.StatusCreated)
	a.Equal(w.Body.String(), "test,123")
	a.Equal(w.Header().Get("content-type"), "json")
	a.Equal(w.Header().Get("content-language"), "zh-hans")

	r = httptest.NewRequest(http.MethodPost, "/path", nil)
	w = httptest.NewRecorder()
	ctx = newContext(w, r, mimetypetest.TextMarshal, encoding.Nop, mimetypetest.TextUnmarshal, encoding.Nop)
	obj = &mimetypetest.TextObject{Name: "test", Age: 1234}
	a.NotError(ctx.Marshal(http.StatusCreated, obj, nil))
	a.Equal(w.Code, http.StatusCreated)
	a.Equal(w.Body.String(), "test,1234")
	a.Equal(w.Header().Get("content-language"), "") // 未指定

	// 输出 nil
	r = httptest.NewRequest(http.MethodPost, "/path", nil)
	w = httptest.NewRecorder()
	ctx = newContext(w, r, mimetypetest.TextMarshal, encoding.Nop, mimetypetest.TextUnmarshal, encoding.Nop)
	ctx.OutputTag = language.MustParse("zh-Hans")
	a.NotError(ctx.Marshal(http.StatusCreated, nil, nil))
	a.Equal(w.Code, http.StatusCreated)
	a.Equal(w.Body.String(), "")
	a.Equal(w.Header().Get("content-language"), "zh-Hans") // 指定了输出语言

	// 输出 Nil，text. 未实现对 nil 值的解析，所以采用了 json.Marshal
	r = httptest.NewRequest(http.MethodPost, "/path", nil)
	w = httptest.NewRecorder()
	ctx = newContext(w, r, json.Marshal, encoding.Nop, mimetypetest.TextUnmarshal, encoding.Nop)
	a.NotError(ctx.Marshal(http.StatusCreated, mimetype.Nil, nil))
	a.Equal(w.Code, http.StatusCreated)
	a.Equal(w.Body.String(), "null")

	// 输出不同编码的内容
	r = httptest.NewRequest(http.MethodPost, "/path", nil)
	w = httptest.NewRecorder()
	ctx = newContext(w, r, mimetypetest.TextMarshal, simplifiedchinese.GB18030, mimetypetest.TextUnmarshal, encoding.Nop)
	a.NotError(ctx.Marshal(http.StatusCreated, gbkstr2, nil))
	a.Equal(w.Code, http.StatusCreated)
	a.Equal(w.Body.Bytes(), gbkdata2)
}

func TestContext_Render(t *testing.T) {
	a := assert.New(t)

	r := httptest.NewRequest(http.MethodPost, "/path", nil)
	w := httptest.NewRecorder()
	ctx := newContext(w, r, mimetypetest.TextMarshal, nil, mimetypetest.TextUnmarshal, nil)
	obj := &mimetypetest.TextObject{Name: "test", Age: 123}
	ctx.Render(http.StatusCreated, obj, nil)
	a.Equal(w.Code, http.StatusCreated)
	a.Equal(w.Body.String(), "test,123")

	r = httptest.NewRequest(http.MethodPost, "/path", nil)
	w = httptest.NewRecorder()
	ctx = newContext(w, r, mimetypetest.TextMarshal, nil, mimetypetest.TextUnmarshal, nil)
	obj1 := &struct{ Name string }{Name: "name"}
	ctx.Render(http.StatusCreated, obj1, nil)
	a.Equal(w.Code, http.StatusInternalServerError)
}

func TestContext_ClientIP(t *testing.T) {
	a := assert.New(t)
	w := httptest.NewRecorder()

	r := httptest.NewRequest(http.MethodPost, "/path", nil)
	ctx := newContext(w, r, mimetypetest.TextMarshal, nil, mimetypetest.TextUnmarshal, nil)
	a.Equal(ctx.ClientIP(), r.RemoteAddr)

	// httptest.NewRequest 会直接将  remote-addr 赋值为 192.0.2.1 无法测试
	r, err := http.NewRequest(http.MethodPost, "/path", nil)
	a.NotError(err).NotNil(r)
	r.Header.Set("x-real-ip", "192.168.1.1:8080")
	ctx = newContext(w, r, mimetypetest.TextMarshal, nil, mimetypetest.TextUnmarshal, nil)
	a.Equal(ctx.ClientIP(), "192.168.1.1:8080")

	r = httptest.NewRequest(http.MethodPost, "/path", nil)
	r.Header.Set("x-forwarded-for", "192.168.2.1:8080,192.168.2.2:111")
	ctx = newContext(w, r, mimetypetest.TextMarshal, nil, mimetypetest.TextUnmarshal, nil)
	a.Equal(ctx.ClientIP(), "192.168.2.1:8080")

	// 测试获取 IP 报头的优先级
	r = httptest.NewRequest(http.MethodPost, "/path", nil)
	r.Header.Set("x-forwarded-for", "192.168.2.1:8080,192.168.2.2:111")
	r.Header.Set("x-real-ip", "192.168.2.2")
	ctx = newContext(w, r, mimetypetest.TextMarshal, nil, mimetypetest.TextUnmarshal, nil)
	a.Equal(ctx.ClientIP(), "192.168.2.1:8080")

	// 测试获取 IP 报头的优先级
	r = httptest.NewRequest(http.MethodPost, "/path", nil)
	r.Header.Set("Remote-Addr", "192.168.2.0")
	r.Header.Set("x-forwarded-for", "192.168.2.1:8080,192.168.2.2:111")
	r.Header.Set("x-real-ip", "192.168.2.2")
	ctx = newContext(w, r, mimetypetest.TextMarshal, nil, mimetypetest.TextUnmarshal, nil)
	a.Equal(ctx.ClientIP(), "192.168.2.1:8080")
}
