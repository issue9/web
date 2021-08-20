// SPDX-License-Identifier: MIT

package content

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
	"golang.org/x/text/message/catalog"

	"github.com/issue9/web/content/text"
	"github.com/issue9/web/content/text/testobject"
	"github.com/issue9/web/internal/charsetdata"
	"github.com/issue9/web/serialization"
)

func newLocale(a *assert.Assertion) *serialization.Locale {
	l := serialization.NewLocale(catalog.NewBuilder(), serialization.NewFiles(10))
	a.NotNil(l)
	return l
}

func TestContent_NewContext(t *testing.T) {
	a := assert.New(t)
	w := httptest.NewRecorder()
	lw := &bytes.Buffer{}
	l := log.New(lw, "", 0)

	c := New(DefaultBuilder, newLocale(a), language.SimplifiedChinese)
	a.NotError(c.Mimetypes().Add(text.Marshal, text.Unmarshal, text.Mimetype))

	b := c.Locale().Builder()
	a.NotError(b.SetString(language.Chinese, "test", "中文"))
	a.NotError(b.SetString(language.SimplifiedChinese, "test", "简体"))
	a.NotError(b.SetString(language.TraditionalChinese, "test", "繁体"))
	a.NotError(b.SetString(language.English, "test", "english"))

	// 错误的 accept
	r := httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Set("Accept", "not")
	ctx, status := c.NewContext(l, w, r)
	a.Equal(status, http.StatusNotAcceptable).Nil(ctx)
	a.Contains(lw.String(), "解码函数")

	// 错误的 accept-charset
	lw.Reset()
	r = httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Set("Accept", text.Mimetype)
	r.Header.Set("Accept-Charset", "unknown")
	ctx, status = c.NewContext(l, w, r)
	a.Contains(lw.String(), "字符集")
	a.Equal(status, http.StatusNotAcceptable).Nil(ctx)

	// 错误的 content-type,无输入内容
	lw.Reset()
	r = httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Set("Content-Type", ";charset=utf-8")
	ctx, status = c.NewContext(l, w, r)
	a.Equal(status, http.StatusUnsupportedMediaType).Nil(ctx)
	a.NotEmpty(lw.String())

	// 错误的 content-type,有输入内容
	lw.Reset()
	r = httptest.NewRequest(http.MethodPost, "/path", bytes.NewBufferString("[]"))
	r.Header.Set("Content-Type", ";charset=utf-8")
	ctx, status = c.NewContext(l, w, r)
	a.Equal(status, http.StatusUnsupportedMediaType).Nil(ctx)
	a.NotEmpty(lw.String())

	// 错误的 content-type，且有输入内容
	lw.Reset()
	r = httptest.NewRequest(http.MethodGet, "/path", bytes.NewBufferString("123"))
	r.Header.Set("content-type", buildContentType(text.Mimetype, "utf-"))
	ctx, status = c.NewContext(l, w, r)
	a.Equal(status, http.StatusUnsupportedMediaType).Nil(ctx)
	a.NotEmpty(lw.String())

	// 部分错误的 Accept-Language
	lw.Reset()
	r = httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Set("Accept-Language", "zh-hans;q=0.9,zh-Hant;q=xxx")
	r.Header.Set("content-type", buildContentType(text.Mimetype, DefaultCharset))
	ctx, status = c.NewContext(l, w, r)
	a.Empty(status).NotNil(ctx)
	a.Equal(ctx.OutputTag, language.MustParse("zh-hans"))
	a.Empty(lw.String())

	// 正常，指定 Accept-Language
	lw.Reset()
	r = httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Set("Accept-Language", "zh-hans;q=0.9,zh-Hant;q=0.7")
	r.Header.Set("content-type", buildContentType(text.Mimetype, DefaultCharset))
	ctx, status = c.NewContext(l, w, r)
	a.Empty(status).NotNil(ctx)
	a.Empty(lw.String())
	a.True(charsetIsNop(ctx.InputCharset)).
		Equal(ctx.OutputMimetypeName, text.Mimetype).
		Equal(ctx.OutputTag, language.SimplifiedChinese).
		NotNil(ctx.LocalePrinter)

	// 正常，未指定 Accept-Language 和 Accept-Charset 等不是必须的报头
	lw.Reset()
	r = httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Set("content-type", buildContentType(text.Mimetype, DefaultCharset))
	ctx, status = c.NewContext(l, w, r)
	a.Empty(lw.String())
	a.NotNil(ctx).Empty(status).
		True(charsetIsNop(ctx.InputCharset)).
		Equal(ctx.OutputMimetypeName, text.Mimetype)

	// 正常，未指定 Accept-Language 和 Accept-Charset 等不是必须的报头，且有输入内容
	lw.Reset()
	r = httptest.NewRequest(http.MethodGet, "/path", bytes.NewBufferString("123"))
	r.Header.Set("content-type", buildContentType(text.Mimetype, DefaultCharset))
	ctx, status = c.NewContext(l, w, r)
	a.Empty(lw.String())
	a.NotNil(ctx).Empty(status).
		True(charsetIsNop(ctx.InputCharset)).
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
		Request:       httptest.NewRequest(http.MethodGet, "/path", bytes.NewBuffer(charsetdata.GBKData1)),
		Response:      httptest.NewRecorder(),
	}
	data, err = ctx.Body()
	a.NotError(err).Equal(string(data), charsetdata.GBKString1)
	a.Equal(ctx.body, data)

	// 采用不同的编码
	c := New(DefaultBuilder, newLocale(a), language.SimplifiedChinese)
	a.NotError(c.Mimetypes().Add(text.Marshal, text.Unmarshal, text.Mimetype))
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/path", bytes.NewBuffer(charsetdata.GBKData1))
	r.Header.Set("Accept", "*/*")
	r.Header.Set("Content-Type", buildContentType(text.Mimetype, " gb18030"))
	ctx, status := c.NewContext(nil, w, r)
	a.Empty(status).NotNil(ctx)
	data, err = ctx.Body()
	a.NotError(err).Equal(string(data), charsetdata.GBKString1)
	a.Equal(ctx.body, data)
}

func TestContext_Unmarshal(t *testing.T) {
	a := assert.New(t)

	ctx := &Context{
		Request:       httptest.NewRequest(http.MethodGet, "/path", bytes.NewBufferString("test,123")),
		Response:      httptest.NewRecorder(),
		InputMimetype: text.Unmarshal,
	}

	obj := &testobject.TextObject{}
	a.NotError(ctx.Unmarshal(obj))
	a.Equal(obj.Name, "test").Equal(obj.Age, 123)

	// 无法转换
	o := &struct{}{}
	a.Error(ctx.Unmarshal(o))
}

func TestContext_Marshal(t *testing.T) {
	a := assert.New(t)
	c := New(DefaultBuilder, newLocale(a), language.SimplifiedChinese)
	a.NotError(c.Mimetypes().Add(text.Marshal, text.Unmarshal, text.Mimetype))

	// 自定义报头
	r := httptest.NewRequest(http.MethodPost, "/path", nil)
	w := httptest.NewRecorder()
	r.Header.Set("Content-Type", text.Mimetype)
	r.Header.Set("Accept", text.Mimetype)
	ctx, status := c.NewContext(nil, w, r)
	a.Empty(status).NotNil(ctx)
	obj := &testobject.TextObject{Name: "test", Age: 123}
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
	obj = &testobject.TextObject{Name: "test", Age: 1234}
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
	a.NotError(ctx.Marshal(http.StatusCreated, charsetdata.GBKString2, nil))
	a.Equal(w.Code, http.StatusCreated)
	a.Equal(w.Body.Bytes(), charsetdata.GBKData2)
}

func TestAcceptCharset(t *testing.T) {
	a := assert.New(t)

	name, enc := acceptCharset(DefaultCharset)
	a.Equal(name, DefaultCharset).
		True(charsetIsNop(enc))

	name, enc = acceptCharset("")
	a.Equal(name, DefaultCharset).
		True(charsetIsNop(enc))

	// * 表示采用默认的编码
	name, enc = acceptCharset("*")
	a.Equal(name, DefaultCharset).
		True(charsetIsNop(enc))

	name, enc = acceptCharset("gbk")
	a.Equal(name, "gbk").
		Equal(enc, simplifiedchinese.GBK)

	// 传递一个非正规名称
	name, enc = acceptCharset("chinese")
	a.Equal(name, "gbk").
		Equal(enc, simplifiedchinese.GBK)

	// q 错解析错误
	name, enc = acceptCharset("utf-8;q=x.9,gbk;q=0.8")
	a.Equal(name, "gbk").
		Equal(enc, simplifiedchinese.GBK)

	// 不支持的编码
	name, enc = acceptCharset("not-supported")
	a.Empty(name).
		Nil(enc)
}

func TestContent_acceptLanguage(t *testing.T) {
	a := assert.New(t)

	c := New(DefaultBuilder, newLocale(a), language.Afrikaans)
	b := c.Locale().Builder()
	a.NotError(b.SetString(language.Und, "lang", "und"))
	a.NotError(b.SetString(language.SimplifiedChinese, "lang", "hans"))
	a.NotError(b.SetString(language.TraditionalChinese, "lang", "hant"))
	a.NotError(b.SetString(language.AmericanEnglish, "lang", "en_US"))

	tag := c.acceptLanguage("")
	a.Equal(tag, language.Afrikaans, "v1:%s, v2:%s", tag.String(), language.Und.String())

	tag = c.acceptLanguage("zh") // 匹配 zh-hans
	a.Equal(tag, language.SimplifiedChinese, "v1:%s, v2:%s", tag.String(), language.SimplifiedChinese.String())

	tag = c.acceptLanguage("zh-Hant")
	a.Equal(tag, language.TraditionalChinese, "v1:%s, v2:%s", tag.String(), language.TraditionalChinese.String())

	tag = c.acceptLanguage("zh-Hans")
	a.Equal(tag, language.SimplifiedChinese, "v1:%s, v2:%s", tag.String(), language.SimplifiedChinese.String())

	tag = c.acceptLanguage("zh-Hans;q=0.1,zh-Hant;q=0.3,en")
	a.Equal(tag, language.AmericanEnglish, "v1:%s, v2:%s", tag.String(), language.AmericanEnglish.String())
}

func TestContent_contentType(t *testing.T) {
	a := assert.New(t)

	mt := New(DefaultBuilder, newLocale(a), language.SimplifiedChinese)
	a.NotNil(mt)

	f, e, err := mt.conentType(";;;")
	a.Error(err).Nil(f).Nil(e)

	// 不存在的 mimetype
	f, e, err = mt.conentType(buildContentType(DefaultMimetype, DefaultCharset))
	a.Error(err).Nil(f).Nil(e)

	mt.Mimetypes().Add(nil, json.Unmarshal, DefaultMimetype)
	f, e, err = mt.conentType(buildContentType(DefaultMimetype, DefaultCharset))
	a.NotError(err).NotNil(f).NotNil(e)

	// 无效的字符集名称
	f, e, err = mt.conentType(buildContentType(DefaultMimetype, "invalid-charset"))
	a.Error(err).Nil(f).Nil(e)
}
