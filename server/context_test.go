// SPDX-License-Identifier: MIT

package server

import (
	"bytes"
	"compress/flate"
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/issue9/assert/v2"
	"github.com/issue9/assert/v2/rest"
	"github.com/issue9/localeutil"
	"github.com/issue9/logs/v4"
	"github.com/issue9/mux/v7"
	"github.com/issue9/term/v3/colors"
	"golang.org/x/text/language"

	"github.com/issue9/web/internal/header"
	"github.com/issue9/web/serializer"
	"github.com/issue9/web/serializer/text"
	"github.com/issue9/web/serializer/text/testobject"
)

const (
	gbkString1 = "中文1,11"
	gbkString2 = "中文2,22"
)

var (
	gbkBytes1 = []byte{214, 208, 206, 196, 49, 44, 49, 49}
	gbkBytes2 = []byte{214, 208, 206, 196, 50, 44, 50, 50}
)

func newServer(a *assert.Assertion, o *Options) *Server {
	if o == nil {
		o = &Options{HTTPServer: &http.Server{Addr: ":8080"}}
	}
	if o.Logs == nil { // 默认重定向到 os.Stderr
		o.Logs = logs.New(logs.NewTermWriter("[15:04:05]", colors.Red, os.Stderr), logs.Caller, logs.Created)
	}

	srv, err := New("app", "0.1.0", o)
	a.NotError(err).NotNil(srv)
	a.Equal(srv.Name(), "app").Equal(srv.Version(), "0.1.0")

	// mimetype
	mimetype := srv.Mimetypes()
	a.NotError(mimetype.Add(json.Marshal, json.Unmarshal, "application/json"))
	a.NotError(mimetype.Add(xml.Marshal, xml.Unmarshal, "application/xml"))
	a.NotError(mimetype.Add(text.Marshal, text.Unmarshal, text.Mimetype))
	a.NotError(mimetype.Add(nil, nil, "nil"))

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

	srv.Problems().Add("41110", 411, localeutil.Phrase("lang"), localeutil.Phrase("41110"))
	srv.Problems().AddMimetype("application/json", "application/problem+json")

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
	srv := newServer(a, &Options{LanguageTag: language.SimplifiedChinese})
	srv.Logs().SetOutput(logs.NewTextWriter("2006-01-02", lw))
	srv.Logs().Enable(logs.LevelDebug)

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
		Header("Accept", text.Mimetype).
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
	r = rest.Post(a, "/path", []byte("123")).Header("Content-Type", header.BuildContentType(text.Mimetype, "utf-")).Request()
	r.Header.Set("content-type", header.BuildContentType(text.Mimetype, "utf-"))
	srv.newContext(w, r, nil)
	a.Equal(w.Code, http.StatusUnsupportedMediaType)
	a.NotEmpty(lw.String())

	// 部分错误的 Accept-Language
	lw.Reset()
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").
		Header("Accept-Language", "zh-hans;q=0.9,zh-Hant;q=xxx").
		Header("content-type", header.BuildContentType(text.Mimetype, header.UTF8Name)).
		Request()
	ctx := srv.newContext(w, r, nil)
	a.NotNil(ctx)
	a.Equal(ctx.LanguageTag(), language.MustParse("zh-hans"))
	a.Empty(lw.String())

	// 正常，指定 Accept-Language，采用默认的 accept
	lw.Reset()
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").
		Header("Accept-Language", "zh-hans;q=0.9,zh-Hant;q=0.7").
		Header("content-type", header.BuildContentType(text.Mimetype, header.UTF8Name)).
		Request()
	ctx = srv.newContext(w, r, nil)
	a.NotNil(ctx)
	a.Empty(lw.String())
	a.True(header.CharsetIsNop(ctx.inputCharset)).
		Equal(ctx.outputMimetypeName, "application/json").
		Equal(ctx.outputCharsetName, "utf-8").
		Equal(ctx.inputMimetype, serializer.UnmarshalFunc(text.Unmarshal)).
		Equal(ctx.LanguageTag(), language.SimplifiedChinese).
		NotNil(ctx.LocalePrinter())

	// 正常，未指定 Accept-Language 和 Accept-Charset 等不是必须的报头
	lw.Reset()
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").
		Header("content-type", header.BuildContentType(text.Mimetype, header.UTF8Name)).
		Header("accept", "application/json;q=0.2,text/plain;q=0.9").
		Request()
	ctx = srv.newContext(w, r, nil)
	a.Empty(lw.String())
	a.NotNil(ctx).
		True(header.CharsetIsNop(ctx.inputCharset)).
		Equal(ctx.outputMimetypeName, text.Mimetype).
		Equal(ctx.outputCharsetName, header.UTF8Name)

	// 正常，未指定 Accept-Language 和 Accept-Charset 等不是必须的报头，且有输入内容
	lw.Reset()
	w = httptest.NewRecorder()
	r = rest.Post(a, "/path", []byte("123")).
		Header("content-type", header.BuildContentType(text.Mimetype, header.UTF8Name)).
		Header("accept", "application/json;q=0.2,text/*;q=0.9").
		Request()
	ctx = srv.newContext(w, r, nil)
	a.Empty(lw.String())
	a.NotNil(ctx).
		True(header.CharsetIsNop(ctx.inputCharset)).
		Equal(ctx.outputMimetypeName, text.Mimetype).
		Equal(ctx.outputCharsetName, header.UTF8Name)
}

func TestContext_Marshal(t *testing.T) {
	a := assert.New(t, false)
	srv := newServer(a, &Options{LanguageTag: language.SimplifiedChinese})

	// 自定义报头
	w := httptest.NewRecorder()
	r := rest.Post(a, "/path", nil).
		Header("Content-Type", text.Mimetype).
		Header("Accept", text.Mimetype).
		Request()
	ctx := srv.newContext(w, r, nil)
	a.NotNil(ctx)
	obj := &testobject.TextObject{Name: "test", Age: 123}
	a.NotError(ctx.Marshal(http.StatusCreated, obj, false))
	a.Equal(w.Code, http.StatusCreated)
	a.Equal(w.Body.String(), "test,123")
	a.Equal(w.Header().Get("content-type"), header.BuildContentType(text.Mimetype, "utf-8"))
	a.Equal(w.Header().Get("content-language"), "zh-Hans")

	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").
		Header("accept", text.Mimetype).
		Header("accept-language", "").
		Request()
	ctx = srv.newContext(w, r, nil)
	obj = &testobject.TextObject{Name: "test", Age: 1234}
	a.NotError(ctx.Marshal(http.StatusCreated, obj, false))
	a.Equal(w.Code, http.StatusCreated)
	a.Equal(w.Body.String(), "test,1234")
	a.Equal(w.Header().Get("content-language"), language.SimplifiedChinese.String()) // 未指定，采用默认值

	// 输出 nil，content-type 和 content-language 均为空
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").
		Header("Accept", text.Mimetype).
		Header("Accept-language", "zh-hans").
		Request()
	ctx = srv.newContext(w, r, nil)
	a.NotError(ctx.Marshal(http.StatusCreated, nil, false))
	a.Equal(w.Code, http.StatusCreated)
	a.Equal(w.Body.String(), "")
	a.Equal(w.Header().Get("content-language"), "") // 指定了输出语言，也返回空。
	a.Equal(w.Header().Get("content-Type"), "")

	// accept,accept-language,accept-charset
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").
		Header("Accept", text.Mimetype).
		Header("Accept-Language", "zh-Hans").
		Header("Accept-Charset", "gbk").
		Request()
	ctx = srv.newContext(w, r, nil)
	a.NotError(ctx.Marshal(http.StatusCreated, gbkString2, false))
	a.Equal(w.Code, http.StatusCreated)
	a.Equal(w.Body.Bytes(), gbkBytes2)

	// problem
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").Header("Accept", "application/json").Request()
	ctx = srv.newContext(w, r, nil)
	a.NotError(ctx.Marshal(http.StatusCreated, "abc", true))
	a.Equal(w.Code, http.StatusCreated)
	a.Equal(w.Body.Bytes(), `"abc"`).Equal(w.Header().Get("content-type"), "application/problem+json; charset=utf-8")

	// problem, 未指定
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").Header("Accept", text.Mimetype).Request()
	ctx = srv.newContext(w, r, nil)
	a.NotError(ctx.Marshal(http.StatusCreated, "abc", true))
	a.Equal(w.Code, http.StatusCreated)
	a.Equal(w.Body.Bytes(), "abc").Equal(w.Header().Get("content-type"), text.Mimetype+"; charset=utf-8")

	// 同时指定了 accept,accept-language,accept-charset 和 accept-encoding
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").
		Header("Accept", text.Mimetype).
		Header("Accept-Language", "zh-Hans").
		Header("Accept-Charset", "gbk").
		Header("Accept-Encoding", "gzip;q=0.9,deflate").
		Request()
	ctx = srv.newContext(w, r, nil)
	a.NotError(ctx.Marshal(http.StatusCreated, gbkString2, false))
	ctx.destroy()
	a.Equal(w.Code, http.StatusCreated)
	data, err := io.ReadAll(flate.NewReader(w.Body))
	a.NotError(err).Equal(data, gbkBytes2)
	a.Equal(w.Header().Get("content-encoding"), "deflate")

	// 同时通过 ctx.Write 和 ctx.Marshal 输出内容
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").
		Header("Accept", text.Mimetype).
		Header("Accept-Encoding", "gzip;q=0.9,deflate").
		Request()
	ctx = srv.newContext(w, r, nil)
	_, err = ctx.Write([]byte("123"))
	a.NotError(err)
	a.NotError(ctx.Marshal(http.StatusCreated, "456", false))
	ctx.destroy()
	a.Equal(w.Code, http.StatusCreated) // 压缩对象缓存了 WriteHeader 的发送
	data, err = io.ReadAll(flate.NewReader(w.Body))
	a.NotError(err).Equal(string(data), "123456")

	// accept,accept-language,accept-charset 和 accept-encoding，部分 Response.Write 输出
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").
		Header("Accept", text.Mimetype).
		Header("Accept-Language", "zh-Hans").
		Header("Accept-Charset", "gbk").
		Request()
	ctx = srv.newContext(w, r, nil)
	_, err = ctx.Write([]byte(gbkString1))
	a.NotError(err)
	a.NotError(ctx.Marshal(http.StatusCreated, gbkString2, false))
	ctx.destroy()
	a.Equal(w.Code, http.StatusOK) // 未指定压缩，WriteHeader 直接发送
	data, err = io.ReadAll(w.Body)
	a.NotError(err)
	bs := make([]byte, 0, len(gbkBytes1)+len(gbkBytes2))
	bs = append(append(bs, gbkBytes1...), gbkBytes2...)
	a.Equal(data, bs)

	// outputMimetype == nil
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").Header("Accept", "nil").Request()
	ctx = srv.newContext(w, r, nil)
	a.Nil(ctx.outputMimetype).
		Equal(ctx.outputMimetypeName, "nil").
		Equal(ctx.outputCharsetName, header.UTF8Name)
	a.PanicString(func() {
		ctx.Marshal(http.StatusCreated, "val", false)
	}, "未对 nil 作处理")
	a.Equal(w.Code, http.StatusNotAcceptable)

	// outputMimetype 返回 serializer.ErrUnsupported
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").Header("Accept", text.Mimetype).Request()
	ctx = srv.newContext(w, r, nil)
	a.ErrorIs(ctx.Marshal(http.StatusCreated, &struct{}{}, false), serializer.ErrUnsupported)
	a.Equal(w.Code, http.StatusNotAcceptable)

	// outputMimetype 返回错误
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").Header("Accept", text.Mimetype).Request()
	ctx = srv.newContext(w, r, nil)
	a.Error(ctx.Marshal(http.StatusCreated, errors.New("error"), false))
	a.Equal(w.Code, http.StatusInternalServerError)
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

func TestServer_Location(t *testing.T) {
	a := assert.New(t, false)
	srv := newServer(a, nil)

	w := httptest.NewRecorder()
	r := rest.Get(a, "/test").Request()
	ctx := srv.newContext(w, r, nil)
	now := ctx.Now()
	a.Equal(now.Location(), srv.Location()).
		Equal(now.Location(), ctx.Location())

	a.NotError(ctx.SetLocation("UTC"))
	now2 := ctx.Now()
	a.Equal(now2.Location(), ctx.Location())
	if now2.Location() != srv.Location() {
		a.NotEqual(ctx.Location(), srv.Location())
	}
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

func TestContext_LocalePrinter(t *testing.T) {
	a := assert.New(t, false)
	srv := newServer(a, &Options{LanguageTag: language.SimplifiedChinese})

	b := srv.CatalogBuilder()
	a.NotError(b.SetString(language.MustParse("cmn-hans"), "test", "测试"))
	a.NotError(b.SetString(language.MustParse("cmn-hant"), "test", "測試"))

	w := httptest.NewRecorder()
	r := rest.Get(a, "/path").
		Header("accept-language", "cmn-hant").
		Header("accept", text.Mimetype).
		Request()
	ctx := srv.newContext(w, r, nil)
	a.NotNil(ctx)
	a.NotError(ctx.Marshal(http.StatusOK, ctx.Sprintf("test"), false))
	a.Equal(w.Body.String(), "測試")

	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").
		Header("accept-language", "cmn-hans").
		Header("accept", text.Mimetype).
		Request()
	ctx = srv.newContext(w, r, nil)
	a.NotNil(ctx)
	n, err := ctx.LocalePrinter().Fprintf(ctx, "test")
	a.NotError(err).Equal(n, len("测试"))
	a.Equal(w.Body.String(), "测试")
}

func TestContext_Problem(t *testing.T) {
	a := assert.New(t, false)
	srv := newServer(a, nil)
	a.NotError(srv.CatalogBuilder().SetString(language.Und, "lang", "und"))
	a.NotError(srv.CatalogBuilder().SetString(language.SimplifiedChinese, "lang", "hans"))
	srv.Problems().Add("40000", 400, localeutil.Phrase("lang"), localeutil.Phrase("lang")) // lang 有翻译

	// 能正常翻译错误信息
	w := httptest.NewRecorder()
	r := rest.Get(a, "/path").
		Header("accept-language", language.SimplifiedChinese.String()).
		Header("accept", "application/json").
		Request()
	ctx := srv.newContext(w, r, nil)
	resp := ctx.Problem("40000")
	resp.Apply(ctx)
	a.Equal(w.Body.String(), `{"type":"40000","title":"hans","status":400}`)

	// 未指定 accept-language，采用默认的 und
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").
		Header("accept", "application/json").
		Request()
	ctx = srv.newContext(w, r, nil)
	resp = ctx.Problem("40000")
	resp.Apply(ctx)
	a.Equal(w.Body.String(), `{"type":"40000","title":"und","status":400}`)

	// 不存在的本地化信息，采用默认的 und
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").
		Header("accept-language", "en-US").
		Header("accept", "application/json").
		Request()
	ctx = srv.newContext(w, r, nil)
	resp = ctx.Problem("40000").With("with", "abc")
	resp.Apply(ctx)
	a.Equal(w.Body.String(), `{"type":"40000","title":"und","status":400,"with":"abc"}`)

	// 不存在
	a.Panic(func() { ctx.Problem("400") })
	a.Panic(func() { ctx.Problem("50000") })

	// with field

	r = rest.Post(a, "/path", []byte("123")).
		Header("Accept", "application/json").
		Header("Content-Type", "application/json").
		Request()
	w = httptest.NewRecorder()
	ctx = newServer(a, nil).newContext(w, r, nil)
	ctx.Server().Problems().Add("40010", http.StatusBadRequest, localeutil.Phrase("40010"), localeutil.Phrase("40010"))
	ctx.Server().Problems().Add("40011", http.StatusBadRequest, localeutil.Phrase("40011"), localeutil.Phrase("40011"))

	resp = ctx.Problem("40010").With("detail", "40010")
	resp.AddParam("k1", "v1")

	resp.Apply(ctx)
	a.Equal(w.Body.String(), `{"type":"40010","title":"40010","status":400,"params":[{"name":"k1","reason":"v1"}],"detail":"40010"}`)
}
