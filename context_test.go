// SPDX-License-Identifier: MIT

package web

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/issue9/assert/v3"
	"github.com/issue9/assert/v3/rest"
	"github.com/issue9/mux/v7"
	"golang.org/x/text/language"

	"github.com/issue9/web/internal/header"
	"github.com/issue9/web/logs"
	"github.com/issue9/web/serializer"
	"github.com/issue9/web/servertest"
)

var _ http.ResponseWriter = &Context{}

func marshalTest(_ *Context, v any) ([]byte, error) {
	switch vv := v.(type) {
	case error:
		return nil, vv
	default:
		return nil, serializer.ErrUnsupported()
	}
}

func unmarshalTest(bs []byte, v any) error {
	return serializer.ErrUnsupported()
}

func marshalJSON(ctx *Context, obj any) ([]byte, error) {
	return json.Marshal(obj)
}

func marshalXML(ctx *Context, obj any) ([]byte, error) {
	return xml.Marshal(obj)
}

func TestContext_vars(t *testing.T) {
	a := assert.New(t, false)
	r := rest.Get(a, "/path").Header("Accept", "*/*").Request()
	w := httptest.NewRecorder()
	ctx := newTestServer(a, nil).newContext(w, r, nil)
	a.NotNil(ctx)

	type (
		t1 struct{ int }
		t2 int64
		t3 t2
	)
	var (
		k1    = t1{1}
		k2 t2 = 1
		k3 t3 = 1
	)

	ctx.SetVar(k1, 1)
	ctx.SetVar(k2, 2)
	ctx.SetVar(k3, 3)

	v1, found := ctx.GetVar(k1)
	a.True(found).Equal(v1, 1)

	v2, found := ctx.GetVar(k2)
	a.True(found).Equal(v2, 2)

	v3, found := ctx.GetVar(k3)
	a.True(found).Equal(v3, 3)
}

func TestServer_Context(t *testing.T) {
	a := assert.New(t, false)
	lw := &bytes.Buffer{}
	o := &Options{
		Locale:     &Locale{Language: language.SimplifiedChinese},
		Logs:       &logs.Options{Handler: logs.NewTextHandler("2006-01-02", lw), Levels: logs.AllLevels()},
		HTTPServer: &http.Server{Addr: ":8080"},
	}
	srv := newTestServer(a, o)
	router := srv.NewRouter("def", nil)
	router.Get("/path", func(ctx *Context) Responser { ctx.Render(http.StatusOK, nil); return nil })

	defer servertest.Run(a, srv)()
	defer srv.Close(0)

	// 错误的 accept
	servertest.Get(a, "http://localhost:8080/path").Header("Accept", "not").
		Do(nil).
		Status(http.StatusNotAcceptable)
	a.Contains(lw.String(), Phrase("not found serialization for %s", "not").LocaleString(srv.LocalePrinter()))

	// 错误的 accept-charset
	lw.Reset()
	servertest.Get(a, "http://localhost:8080/path").Header("Accept", "not").
		Header("Accept", "application/json").
		Header("Accept-Charset", "unknown").
		Do(nil).
		Status(http.StatusNotAcceptable)
	a.Contains(lw.String(), Phrase("not found charset for %s", "unknown").LocaleString(srv.LocalePrinter()))

	// 错误的 content-type,无输入内容
	lw.Reset()
	servertest.Get(a, "http://localhost:8080/path").Header("Content-Type", ";charset=utf-8").Do(nil).
		Status(http.StatusUnsupportedMediaType)
	a.NotEmpty(lw.String())

	// 错误的 content-type,有输入内容
	lw.Reset()
	servertest.Post(a, "http://localhost:8080/path", []byte("[]")).Header("Content-Type", ";charset=utf-8").Do(nil).
		Status(http.StatusUnsupportedMediaType)
	a.NotEmpty(lw.String())

	// 错误的 content-type，且有输入内容
	lw.Reset()
	servertest.Post(a, "http://localhost:8080/path", []byte("123")).
		Header("Content-Type", header.BuildContentType("application/json", "utf-")).
		Do(nil).
		Status(http.StatusUnsupportedMediaType)
	a.NotEmpty(lw.String())

	// 部分错误的 Accept-Language
	var b1 time.Time
	lw.Reset()
	router.Get("/p2", func(ctx *Context) Responser {
		a.NotNil(ctx).NotEmpty(ctx.ID())
		a.Equal(ctx.LanguageTag(), language.MustParse("zh-hans"))
		a.Empty(lw.String())
		a.NotZero(ctx.Begin())
		b1 = ctx.Begin()
		return nil
	})
	servertest.Get(a, "http://localhost:8080/p2").
		Header("Accept-Language", "zh-hans;q=0.9,zh-Hant;q=xxx").
		Header("content-type", header.BuildContentType("application/json", header.UTF8Name)).
		Do(nil).
		Success()
	time.Sleep(500 * time.Microsecond) // 保证后续的 Context.Begin 与此值有时间差。

	// 正常，指定 Accept-Language，采用默认的 accept
	lw.Reset()
	router.Get("/p3", func(ctx *Context) Responser {
		a.NotNil(ctx).NotEmpty(ctx.ID())
		a.True(header.CharsetIsNop(ctx.inputCharset)).
			Equal(ctx.Mimetype(false), "application/json").
			Equal(ctx.outputCharsetName, "utf-8").
			Equal(ctx.inputMimetype, UnmarshalFunc(json.Unmarshal)).
			Equal(ctx.LanguageTag(), language.SimplifiedChinese).
			NotNil(ctx.LocalePrinter())
		return nil
	})
	servertest.Get(a, "http://localhost:8080/p3").
		Header("Accept-Language", "cmn-hans;q=0.9,zh-Hant;q=0.7").
		Header("content-type", header.BuildContentType("application/json", header.UTF8Name)).
		Do(nil).
		Success()
	a.Empty(lw.String())

	// 正常，未指定 Accept-Language 和 Accept-Charset 等不是必须的报头
	lw.Reset()
	router.Get("/p4", func(ctx *Context) Responser {
		a.NotNil(ctx).
			True(header.CharsetIsNop(ctx.inputCharset)).
			Equal(ctx.Mimetype(false), "application/json").
			Equal(ctx.outputCharsetName, header.UTF8Name)
		b2 := ctx.Begin()
		a.True(b1.Before(b2))
		return nil
	})
	servertest.Get(a, "http://localhost:8080/p4").
		Header("content-type", header.BuildContentType("application/json", header.UTF8Name)).
		Header("accept", "application/json;q=0.2,text/plain;q=0.9").
		Do(nil)
	a.Empty(lw.String())

	// 正常，未指定 Accept-Language 和 Accept-Charset 等不是必须的报头，且有输入内容
	lw.Reset()
	router.Post("/p5", func(ctx *Context) Responser {
		a.NotNil(ctx).
			True(header.CharsetIsNop(ctx.inputCharset)).
			Equal(ctx.Mimetype(false), "application/json").
			Equal(ctx.outputCharsetName, header.UTF8Name)
		ctx.WriteHeader(http.StatusCreated)
		return nil
	})
	servertest.Post(a, "http://localhost:8080/p5", []byte("123")).
		Header("content-type", header.BuildContentType("application/json", header.UTF8Name)).
		Header("accept", "application/json;q=0.2,text/*;q=0.9").
		Do(nil).
		Status(http.StatusCreated)
	a.Empty(lw.String())
}

func TestContext_SetMimetype(t *testing.T) {
	a := assert.New(t, false)
	srv := newTestServer(a, nil)

	w := httptest.NewRecorder()
	r := rest.Post(a, "/path", []byte("123")).
		Header("accept", "application/json").
		Request()
	ctx := srv.NewContext(w, r)
	a.NotNil(ctx)

	a.PanicString(func() {
		ctx.SetMimetype("not-exists")
	}, "指定的编码 not-exists 不存在")
	a.Equal(ctx.Mimetype(false), "application/json") // 不改变原有的值

	ctx.SetMimetype("application/xml")
	a.Equal(ctx.Mimetype(false), "application/xml")

	ctx.Render(200, 200) // 输出内容
	a.PanicString(func() {
		ctx.SetMimetype("application/json")
	}, "已有内容输出，不可再更改！")
}

func TestContext_SetCharset(t *testing.T) {
	a := assert.New(t, false)
	srv := newTestServer(a, nil)

	w := httptest.NewRecorder()
	r := rest.Post(a, "/path", []byte("123")).
		Header("accept", "application/json").
		Request()
	ctx := srv.newContext(w, r, nil)
	a.NotNil(ctx)

	a.PanicString(func() {
		ctx.SetCharset("not-exists")
	}, "指定的字符集 not-exists 不存在")

	ctx.SetCharset("gb2312")
	a.Equal(ctx.Charset(), "gbk")

	ctx.Render(200, 200) // 输出内容
	a.PanicString(func() {
		ctx.SetCharset("gb18030")
	}, "已有内容输出，不可再更改！")
}

func TestContext_SetEncoding(t *testing.T) {
	a := assert.New(t, false)
	srv := newTestServer(a, nil)

	w := httptest.NewRecorder()
	r := rest.Post(a, "/path", []byte("123")).
		Header("accept", "application/json").
		Request()
	ctx := srv.newContext(w, r, nil)
	a.NotNil(ctx)

	a.PanicString(func() {
		ctx.SetEncoding("*;q=0")
	}, "指定的压缩编码 *;q=0 不存在")

	ctx.SetEncoding("gzip")
	a.Equal(ctx.Encoding(), "gzip")

	ctx.Write([]byte("200"))
	a.PanicString(func() {
		ctx.SetEncoding("br")
	}, "已有内容输出，不可再更改！")
}

func TestContext_SetLanguage(t *testing.T) {
	a := assert.New(t, false)
	srv := newTestServer(a, nil)

	w := httptest.NewRecorder()
	r := rest.Post(a, "/path", []byte("123")).
		Header("accept", "application/json").
		Request()
	ctx := srv.newContext(w, r, nil)
	a.NotNil(ctx)

	a.Equal(ctx.LanguageTag(), ctx.Server().Language())

	cmnHant := language.MustParse("cmn-hant")
	ctx.SetLanguage(cmnHant)
	a.Equal(ctx.LanguageTag(), cmnHant)
}

func TestContext_IsXHR(t *testing.T) {
	a := assert.New(t, false)

	srv := newTestServer(a, nil)
	router := srv.NewRouter("router", nil, mux.URLDomain("https://example.com"))
	a.NotNil(router)
	router.Get("/not-xhr", func(ctx *Context) Responser {
		a.False(ctx.IsXHR())
		return nil
	})
	router.Get("/xhr", func(ctx *Context) Responser {
		a.True(ctx.IsXHR())
		return nil
	})

	r := rest.Get(a, "/not-xhr").Request()
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)

	r = rest.Get(a, "/xhr").Request()
	r.Header.Set("X-Requested-With", "XMLHttpRequest")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
}

func TestServer_acceptLanguage(t *testing.T) {
	a := assert.New(t, false)

	srv := newTestServer(a, &Options{Locale: &Locale{Language: language.Afrikaans}})
	b := srv.CatalogBuilder()
	a.NotError(b.SetString(language.Und, "lang", "und"))
	a.NotError(b.SetString(language.SimplifiedChinese, "lang", "hans"))
	a.NotError(b.SetString(language.TraditionalChinese, "lang", "hant"))
	a.NotError(b.SetString(language.AmericanEnglish, "lang", "en_US"))

	tag := srv.acceptLanguage("")
	a.Equal(tag, language.Afrikaans, "v1:%s, v2:%s", tag.String(), language.Und.String())

	tag = srv.acceptLanguage("zh") // 匹配 zh-hans
	a.Equal(tag, language.SimplifiedChinese, "v1:%s, v2:%s", tag.String(), language.SimplifiedChinese.String())

	tag = srv.acceptLanguage("zh-Hant")
	a.Equal(tag, language.TraditionalChinese, "v1:%s, v2:%s", tag.String(), language.TraditionalChinese.String())

	tag = srv.acceptLanguage("zh-Hans")
	a.Equal(tag, language.SimplifiedChinese, "v1:%s, v2:%s", tag.String(), language.SimplifiedChinese.String())

	tag = srv.acceptLanguage("english") // english 非正确的 tag，但是常用。
	a.Equal(tag, language.AmericanEnglish, "v1:%s, v2:%s", tag.String(), language.AmericanEnglish.String())

	tag = srv.acceptLanguage("zh-Hans;q=0.1,zh-Hant;q=0.3,en")
	a.Equal(tag, language.AmericanEnglish, "v1:%s, v2:%s", tag.String(), language.AmericanEnglish.String())
}

func TestContext_ClientIP(t *testing.T) {
	a := assert.New(t, false)
	w := httptest.NewRecorder()

	r := rest.Post(a, "/path", nil).Request()
	ctx := newTestServer(a, nil).newContext(w, r, nil)
	a.Equal(ctx.ClientIP(), r.RemoteAddr)

	r = rest.Post(a, "/path", nil).Header("x-real-ip", "192.168.1.1:8080").Request()
	ctx = newTestServer(a, nil).newContext(w, r, nil)
	a.Equal(ctx.ClientIP(), "192.168.1.1:8080")

	r = rest.Post(a, "/path", nil).
		Header("x-forwarded-for", "192.168.2.1:8080,192.168.2.2:111").
		Request()
	ctx = newTestServer(a, nil).newContext(w, r, nil)
	a.Equal(ctx.ClientIP(), "192.168.2.1:8080")

	// 测试获取 IP 报头的优先级
	r = rest.Post(a, "/path", nil).
		Header("x-forwarded-for", "192.168.2.1:8080,192.168.2.2:111").
		Header("x-real-ip", "192.168.2.2").
		Request()
	ctx = newTestServer(a, nil).newContext(w, r, nil)
	a.Equal(ctx.ClientIP(), "192.168.2.1:8080")

	// 测试获取 IP 报头的优先级
	r = rest.Post(a, "/path", nil).
		Header("Remote-Addr", "192.168.2.0").
		Header("x-forwarded-for", "192.168.2.1:8080,192.168.2.2:111").
		Header("x-real-ip", "192.168.2.2").
		Request()
	ctx = newTestServer(a, nil).newContext(w, r, nil)
	a.Equal(ctx.ClientIP(), "192.168.2.1:8080")
}
