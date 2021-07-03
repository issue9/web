// SPDX-License-Identifier: MIT

package content

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/issue9/assert"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/language"
	"golang.org/x/text/transform"

	"github.com/issue9/web/content/text"
)

var (
	gbkString1         = "中文1,11"
	gbkString2         = "中文2,22"
	gbkData1, gbkData2 []byte
)

func init() {
	reader := transform.NewReader(strings.NewReader(gbkString1), simplifiedchinese.GBK.NewEncoder())
	gbkData, err := ioutil.ReadAll(reader)
	if err != nil {
		panic(err)
	}
	gbkData1 = gbkData

	reader = transform.NewReader(strings.NewReader(gbkString2), simplifiedchinese.GBK.NewEncoder())
	gbkData, err = ioutil.ReadAll(reader)
	if err != nil {
		panic(err)
	}
	gbkData2 = gbkData
}

func TestContent_NewContext(t *testing.T) {
	a := assert.New(t)
	w := httptest.NewRecorder()

	c := New(DefaultBuilder)
	a.NotError(c.AddMimetype(text.Mimetype, text.Marshal, text.Unmarshal))

	b := c.CatalogBuilder()
	a.NotError(b.SetString(language.Chinese, "test", "中文"))
	a.NotError(b.SetString(language.SimplifiedChinese, "test", "简体"))
	a.NotError(b.SetString(language.TraditionalChinese, "test", "繁体"))
	a.NotError(b.SetString(language.English, "test", "english"))

	// 错误的 accept
	r := httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Set("Accept", "not")
	ctx, status := c.NewContext(w, r)
	a.Equal(status, http.StatusNotAcceptable).Nil(ctx)

	// 错误的 accept-charset
	r = httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Set("Accept", DefaultMimetype)
	r.Header.Set("Accept-Charset", "unknown")
	ctx, status = c.NewContext(w, r)
	a.Equal(status, http.StatusNotAcceptable).Nil(ctx)

	// 错误的 content-type,无输入内容
	r = httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Set("Content-Type", ";charset=utf-8")
	ctx, status = c.NewContext(w, r)
	a.Equal(status, http.StatusUnsupportedMediaType).Nil(ctx)

	// 错误的 content-type,有输入内容
	r = httptest.NewRequest(http.MethodPost, "/path", bytes.NewBufferString("[]"))
	r.Header.Set("Content-Type", ";charset=utf-8")
	ctx, status = c.NewContext(w, r)
	a.Equal(status, http.StatusUnsupportedMediaType).Nil(ctx)

	// 错误的 content-type，且有输入内容
	r = httptest.NewRequest(http.MethodGet, "/path", bytes.NewBufferString("123"))
	r.Header.Set("content-type", BuildContentType(text.Mimetype, "utf-"))
	ctx, status = c.NewContext(w, r)
	a.Equal(status, http.StatusUnsupportedMediaType).Nil(ctx)

	// 错误的 Accept-Language
	r = httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Set("Accept-Language", "zh-hans;q=0.9,zh-Hant;q=xxx")
	r.Header.Set("content-type", BuildContentType(text.Mimetype, DefaultCharset))
	ctx, status = c.NewContext(w, r)
	a.Equal(status, 0).NotNil(ctx)
	a.Equal(ctx.OutputTag, language.MustParse("zh-hans"))

	// 正常，指定 Accept-Language
	r = httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Set("Accept-Language", "zh-hans;q=0.9,zh-Hant;q=0.7")
	r.Header.Set("content-type", BuildContentType(text.Mimetype, DefaultCharset))
	ctx, status = c.NewContext(w, r)
	a.Equal(status, 0).NotNil(ctx)
	a.True(CharsetIsNop(ctx.InputCharset)).
		Equal(ctx.OutputMimetypeName, text.Mimetype).
		Equal(ctx.OutputTag, language.SimplifiedChinese).
		NotNil(ctx.LocalePrinter)

	// 正常，未指定 Accept-Language 和 Accept-Charset 等不是必须的报头
	r = httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Set("content-type", BuildContentType(text.Mimetype, DefaultCharset))
	ctx, status = c.NewContext(w, r)
	a.NotNil(ctx).Empty(status).
		True(CharsetIsNop(ctx.InputCharset)).
		Equal(ctx.OutputMimetypeName, text.Mimetype)

	// 正常，未指定 Accept-Language 和 Accept-Charset 等不是必须的报头，且有输入内容
	r = httptest.NewRequest(http.MethodGet, "/path", bytes.NewBufferString("123"))
	r.Header.Set("content-type", BuildContentType(text.Mimetype, DefaultCharset))
	ctx, status = c.NewContext(w, r)
	a.NotNil(ctx).Empty(status).
		True(CharsetIsNop(ctx.InputCharset)).
		Equal(ctx.OutputMimetypeName, text.Mimetype)
}

func TestContext_Body(t *testing.T) {
	a := assert.New(t)
	r := httptest.NewRequest(http.MethodGet, "/path", bytes.NewBufferString("123"))
	w := httptest.NewRecorder()

	// 未缓存
	ctx := &Context{
		Request:  r,
		Response: w,
	}
	a.Nil(ctx.body)
	data, err := ctx.Body()
	a.NotError(err).Equal(data, []byte("123"))
	a.Equal(ctx.body, data)

	// 读取缓存内容
	data, err = ctx.Body()
	a.NotError(err).Equal(data, []byte("123"))
	a.Equal(ctx.body, data)

	// 采用 Nop 即 utf-8 编码
	ctx = &Context{
		OutputCharset: encoding.Nop,
		InputCharset:  encoding.Nop,
		Request:       httptest.NewRequest(http.MethodGet, "/path", bytes.NewBufferString("123")),
		Response:      httptest.NewRecorder(),
	}
	data, err = ctx.Body()
	a.NotError(err).Equal(data, []byte("123"))
	a.Equal(ctx.body, data)

	// 采用不同的编码
	ctx = &Context{
		OutputCharset: encoding.Nop,
		InputCharset:  simplifiedchinese.GB18030,
		Request:       httptest.NewRequest(http.MethodGet, "/path", bytes.NewBuffer(gbkData1)),
		Response:      httptest.NewRecorder(),
	}
	data, err = ctx.Body()
	a.NotError(err).Equal(string(data), gbkString1)
	a.Equal(ctx.body, data)

	// 采用不同的编码
	c := New(DefaultBuilder)
	a.NotError(c.AddMimetype(text.Mimetype, text.Marshal, text.Unmarshal))
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/path", bytes.NewBuffer(gbkData1))
	r.Header.Set("Accept", "*/*")
	r.Header.Set("Content-Type", BuildContentType(text.Mimetype, " gb18030"))
	ctx, status := c.NewContext(w, r)
	a.Empty(status).NotNil(ctx)
	data, err = ctx.Body()
	a.NotError(err).Equal(string(data), gbkString1)
	a.Equal(ctx.body, data)
}

func TestContext_Unmarshal(t *testing.T) {
	a := assert.New(t)

	ctx := &Context{
		Request:       httptest.NewRequest(http.MethodGet, "/path", bytes.NewBufferString("test,123")),
		Response:      httptest.NewRecorder(),
		InputMimetype: text.Unmarshal,
	}

	obj := &text.TestObject{}
	a.NotError(ctx.Unmarshal(obj))
	a.Equal(obj.Name, "test").Equal(obj.Age, 123)

	// 无法转换
	o := &struct{}{}
	a.Error(ctx.Unmarshal(o))
}

func TestContext_Marshal(t *testing.T) {
	a := assert.New(t)
	c := New(DefaultBuilder)
	a.NotError(c.AddMimetype(text.Mimetype, text.Marshal, text.Unmarshal))

	// 自定义报头
	r := httptest.NewRequest(http.MethodPost, "/path", nil)
	w := httptest.NewRecorder()
	r.Header.Set("Content-Type", text.Mimetype)
	r.Header.Set("Accept", text.Mimetype)
	ctx, status := c.NewContext(w, r)
	a.Empty(status).NotNil(ctx)
	obj := &text.TestObject{Name: "test", Age: 123}
	a.NotError(ctx.Marshal(http.StatusCreated, obj, map[string]string{"contEnt-type": "json", "content-lanGuage": "zh-hans"}))
	a.Equal(w.Code, http.StatusCreated)
	a.Equal(w.Body.String(), "test,123")
	a.Equal(w.Header().Get("content-type"), "json")
	a.Equal(w.Header().Get("content-language"), "zh-hans")

	w = httptest.NewRecorder()
	ctx = &Context{
		Request:        httptest.NewRequest(http.MethodGet, "/path", nil),
		Response:       w,
		OutputMimetype: text.Marshal,
	}
	obj = &text.TestObject{Name: "test", Age: 1234}
	a.NotError(ctx.Marshal(http.StatusCreated, obj, nil))
	a.Equal(w.Code, http.StatusCreated)
	a.Equal(w.Body.String(), "test,1234")
	a.Equal(w.Header().Get("content-language"), "") // 未指定

	// 输出 nil
	w = httptest.NewRecorder()
	ctx = &Context{
		Request:        httptest.NewRequest(http.MethodGet, "/path", nil),
		Response:       w,
		OutputMimetype: text.Marshal,
		OutputTag:      language.MustParse("zh-Hans"),
	}
	a.NotError(ctx.Marshal(http.StatusCreated, nil, nil))
	a.Equal(w.Code, http.StatusCreated)
	a.Equal(w.Body.String(), "")
	a.Equal(w.Header().Get("content-language"), "zh-Hans") // 指定了输出语言

	// 输出不同编码的内容
	w = httptest.NewRecorder()
	ctx = &Context{
		Request:           httptest.NewRequest(http.MethodGet, "/path", nil),
		Response:          w,
		OutputMimetype:    text.Marshal,
		OutputTag:         language.MustParse("zh-Hans"),
		OutputCharset:     simplifiedchinese.GB18030,
		OutputCharsetName: "gbk",
	}
	a.NotError(ctx.Marshal(http.StatusCreated, gbkString2, nil))
	a.Equal(w.Code, http.StatusCreated)
	a.Equal(w.Body.Bytes(), gbkData2)
}
