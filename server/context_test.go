// SPDX-License-Identifier: MIT

package server

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/issue9/assert/v3"
	"github.com/issue9/assert/v3/rest"
	"github.com/issue9/localeutil"
	"github.com/issue9/mux/v7"
	"github.com/issue9/term/v3/colors"
	"golang.org/x/text/language"

	"github.com/issue9/web/internal/header"
	"github.com/issue9/web/logs"
	"github.com/issue9/web/serializer"
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

func newServer(a *assert.Assertion, o *Options) *Server {
	if o == nil {
		o = &Options{HTTPServer: &http.Server{Addr: ":8080"}, LanguageTag: language.English} // 指定不存在的语言
	}
	if o.Logs == nil { // 默认重定向到 os.Stderr
		o.Logs = logs.New(logs.NewTermWriter("[15:04:05]", colors.Red, os.Stderr), true, true)
	}

	srv, err := New("app", "0.1.0", o)
	a.NotError(err).NotNil(srv)
	a.Equal(srv.Name(), "app").Equal(srv.Version(), "0.1.0")

	// mimetype
	mimetype := srv.Mimetypes()
	mimetype.Add("application/json", marshalJSON, json.Unmarshal, "application/problem+json")
	mimetype.Add("application/xml", marshalXML, xml.Unmarshal, "")
	mimetype.Add("application/test", marshalTest, unmarshalTest, "")
	mimetype.Add("nil", nil, nil, "")
	a.Equal(mimetype.Len(), 4)

	// locale
	b := srv.CatalogBuilder()
	a.NotError(b.SetString(language.Und, "lang", "und"))
	a.NotError(b.SetString(language.SimplifiedChinese, "lang", "hans"))
	a.NotError(b.SetString(language.TraditionalChinese, "lang", "hant"))

	// encoding
	e := srv.Encodings()
	e.Add("gzip", "gzip", EncodingGZip(8))
	e.Add("deflate", "deflate", EncodingDeflate(8))
	e.Allow("*", "gzip", "deflate")

	srv.Problems().Add(&StatusProblem{ID: "41110", Status: 411, Title: localeutil.Phrase("lang"), Detail: localeutil.Phrase("41110")})

	return srv
}

func TestNew(t *testing.T) {
	a := assert.New(t, false)

	srv, err := New("app", "0.1.0", nil)
	a.NotError(err).NotNil(srv)
	a.False(srv.Uptime().IsZero())
	a.NotNil(srv.Cache())
	a.Equal(srv.Location(), time.Local)
	a.Equal(srv.httpServer.Handler, srv.routers)
	a.Equal(srv.httpServer.Addr, "")
}

func TestContext_Vars(t *testing.T) {
	a := assert.New(t, false)
	r := rest.Get(a, "/path").Header("Accept", "*/*").Request()
	w := httptest.NewRecorder()
	ctx := newServer(a, nil).newContext(w, r, nil)
	a.NotNil(ctx)

	type (
		t1 int
		t2 int64
		t3 = t2
	)
	var (
		v1 t1 = 1
		v2 t2 = 1
		v3 t3 = 1
	)

	ctx.Vars[v1] = 1
	ctx.Vars[v2] = 2
	ctx.Vars[v3] = 3

	a.Equal(ctx.Vars[v1], 1).Equal(ctx.Vars[v2], 3)
}

func TestServer_newContext(t *testing.T) {
	a := assert.New(t, false)
	lw := &bytes.Buffer{}
	l := logs.New(logs.NewTextWriter("2006-01-02", lw), false, false)
	srv := newServer(a, &Options{LanguageTag: language.SimplifiedChinese, Logs: l})
	l.Enable(logs.Debug)

	// 错误的 accept
	w := httptest.NewRecorder()
	r := rest.Get(a, "/path").Header("Accept", "not").Request()
	srv.newContext(w, r, nil)
	a.Equal(w.Code, http.StatusNotAcceptable)
	a.Contains(lw.String(), localeutil.Phrase("not found serialization for %s", "not").LocaleString(srv.LocalePrinter()))

	// 错误的 accept-charset
	lw.Reset()
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").Header("Accept", "not").
		Header("Accept", "application/json").
		Header("Accept-Charset", "unknown").
		Request()
	srv.newContext(w, r, nil)
	a.Equal(w.Code, http.StatusNotAcceptable)
	a.Contains(lw.String(), localeutil.Phrase("not found charset for %s", "unknown").LocaleString(srv.LocalePrinter()))

	// 错误的 content-type,无输入内容
	lw.Reset()
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").Header("Content-Type", ";charset=utf-8").Request()
	srv.newContext(w, r, nil)
	a.Equal(w.Code, http.StatusUnsupportedMediaType)
	a.NotEmpty(lw.String())

	// 错误的 content-type,有输入内容
	lw.Reset()
	w = httptest.NewRecorder()

	r = rest.Post(a, "/path", []byte("[]")).Header("Content-Type", ";charset=utf-8").Request()
	srv.newContext(w, r, nil)
	a.Equal(w.Code, http.StatusUnsupportedMediaType)
	a.NotEmpty(lw.String())

	// 错误的 content-type，且有输入内容
	lw.Reset()
	w = httptest.NewRecorder()
	r = rest.Post(a, "/path", []byte("123")).
		Header("Content-Type", header.BuildContentType("application/json", "utf-")).
		Request()
	srv.newContext(w, r, nil)
	a.Equal(w.Code, http.StatusUnsupportedMediaType)
	a.NotEmpty(lw.String())

	// 部分错误的 Accept-Language
	lw.Reset()
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").
		Header("Accept-Language", "zh-hans;q=0.9,zh-Hant;q=xxx").
		Header("content-type", header.BuildContentType("application/json", header.UTF8Name)).
		Request()
	ctx := srv.newContext(w, r, nil)
	a.NotNil(ctx).NotEmpty(ctx.ID())
	a.Equal(ctx.LanguageTag(), language.MustParse("zh-hans"))
	a.Empty(lw.String())

	// 正常，指定 Accept-Language，采用默认的 accept
	lw.Reset()
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").
		Header("Accept-Language", "zh-hans;q=0.9,zh-Hant;q=0.7").
		Header("content-type", header.BuildContentType("application/json", header.UTF8Name)).
		Request()
	ctx = srv.newContext(w, r, nil)
	a.NotNil(ctx).NotEmpty(ctx.ID())
	a.Empty(lw.String())
	a.True(header.CharsetIsNop(ctx.inputCharset)).
		Equal(ctx.Mimetype(false), "application/json").
		Equal(ctx.outputCharsetName, "utf-8").
		Equal(ctx.inputMimetype, UnmarshalFunc(json.Unmarshal)).
		Equal(ctx.LanguageTag(), language.SimplifiedChinese).
		NotNil(ctx.LocalePrinter())

	// 正常，未指定 Accept-Language 和 Accept-Charset 等不是必须的报头
	lw.Reset()
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").
		Header("content-type", header.BuildContentType("application/json", header.UTF8Name)).
		Header("accept", "application/json;q=0.2,text/plain;q=0.9").
		Request()
	ctx = srv.newContext(w, r, nil)
	a.Empty(lw.String())
	a.NotNil(ctx).
		True(header.CharsetIsNop(ctx.inputCharset)).
		Equal(ctx.Mimetype(false), "application/json").
		Equal(ctx.outputCharsetName, header.UTF8Name)

	// 正常，未指定 Accept-Language 和 Accept-Charset 等不是必须的报头，且有输入内容
	lw.Reset()
	w = httptest.NewRecorder()
	r = rest.Post(a, "/path", []byte("123")).
		Header("content-type", header.BuildContentType("application/json", header.UTF8Name)).
		Header("accept", "application/json;q=0.2,text/*;q=0.9").
		Request()
	ctx = srv.newContext(w, r, nil)
	a.Empty(lw.String())
	a.NotNil(ctx).
		True(header.CharsetIsNop(ctx.inputCharset)).
		Equal(ctx.Mimetype(false), "application/json").
		Equal(ctx.outputCharsetName, header.UTF8Name)
}

func TestContext_SetMimetype(t *testing.T) {
	a := assert.New(t, false)
	srv := newServer(a, nil)

	w := httptest.NewRecorder()
	r := rest.Post(a, "/path", []byte("123")).
		Header("accept", "application/json").
		Request()
	ctx := srv.newContext(w, r, nil)
	a.NotNil(ctx)

	a.PanicString(func() {
		ctx.SetMimetype("not-exists")
	}, "指定的编码 not-exists 不存在")
	a.Equal(ctx.Mimetype(false), "application/json") // 不改变原有的值

	ctx.SetMimetype("application/xml")
	a.Equal(ctx.Mimetype(false), "application/xml")

	ctx.Marshal(200, 200, false) // 输出内容
	a.PanicString(func() {
		ctx.SetMimetype("application/json")
	}, "已有内容输出，不可再更改！")
}

func TestContext_SetCharset(t *testing.T) {
	a := assert.New(t, false)
	srv := newServer(a, nil)

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

	ctx.Marshal(200, 200, false) // 输出内容
	a.PanicString(func() {
		ctx.SetCharset("gb18030")
	}, "已有内容输出，不可再更改！")
}

func TestContext_SetEncoding(t *testing.T) {
	a := assert.New(t, false)
	srv := newServer(a, nil)

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
	srv := newServer(a, nil)

	w := httptest.NewRecorder()
	r := rest.Post(a, "/path", []byte("123")).
		Header("accept", "application/json").
		Request()
	ctx := srv.newContext(w, r, nil)
	a.NotNil(ctx)

	a.Equal(ctx.LanguageTag(), ctx.Server().LanguageTag())

	cmnHant := language.MustParse("cmn-hant")
	ctx.SetLanguage(cmnHant)
	a.Equal(ctx.LanguageTag(), cmnHant)
}

func TestContext_IsXHR(t *testing.T) {
	a := assert.New(t, false)

	srv := newServer(a, nil)
	router := srv.Routers().New("router", nil, mux.URLDomain("https://example.com"))
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

	srv := newServer(a, &Options{LanguageTag: language.Afrikaans})
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
	ctx := newServer(a, nil).newContext(w, r, nil)
	a.Equal(ctx.ClientIP(), r.RemoteAddr)

	r = rest.Post(a, "/path", nil).Header("x-real-ip", "192.168.1.1:8080").Request()
	ctx = newServer(a, nil).newContext(w, r, nil)
	a.Equal(ctx.ClientIP(), "192.168.1.1:8080")

	r = rest.Post(a, "/path", nil).
		Header("x-forwarded-for", "192.168.2.1:8080,192.168.2.2:111").
		Request()
	ctx = newServer(a, nil).newContext(w, r, nil)
	a.Equal(ctx.ClientIP(), "192.168.2.1:8080")

	// 测试获取 IP 报头的优先级
	r = rest.Post(a, "/path", nil).
		Header("x-forwarded-for", "192.168.2.1:8080,192.168.2.2:111").
		Header("x-real-ip", "192.168.2.2").
		Request()
	ctx = newServer(a, nil).newContext(w, r, nil)
	a.Equal(ctx.ClientIP(), "192.168.2.1:8080")

	// 测试获取 IP 报头的优先级
	r = rest.Post(a, "/path", nil).
		Header("Remote-Addr", "192.168.2.0").
		Header("x-forwarded-for", "192.168.2.1:8080,192.168.2.2:111").
		Header("x-real-ip", "192.168.2.2").
		Request()
	ctx = newServer(a, nil).newContext(w, r, nil)
	a.Equal(ctx.ClientIP(), "192.168.2.1:8080")
}
