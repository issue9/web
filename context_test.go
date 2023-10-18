// SPDX-License-Identifier: MIT

package web

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/issue9/assert/v3"
	"github.com/issue9/assert/v3/rest"
	"github.com/issue9/mux/v7"
	"golang.org/x/text/language"

	"github.com/issue9/web/internal/header"
	"github.com/issue9/web/servertest"
)

var _ http.ResponseWriter = &Context{}

func newContext(a *assert.Assertion, w http.ResponseWriter, r *http.Request) *Context {
	if w == nil {
		w = httptest.NewRecorder()
	}
	if r == nil {
		r := httptest.NewRequest(http.MethodGet, "/path", bytes.NewBufferString("123"))
		r.Header.Set(header.Accept, "*/*")
	}

	return newTestServer(a).NewContext(w, r)
}

func TestContext_KeepAlive(t *testing.T) {
	a := assert.New(t, false)
	srv := newTestServer(a)
	router := srv.NewRouter("def", nil)
	router.Get("/path", func(ctx *Context) Responser {
		ctx.Header().Set(header.ContentLength, "0")
		ctx.Header().Set("Cache-Control", "no-cache")
		ctx.Header().Set("Connection", "keep-alive")
		ctx.WriteHeader(http.StatusCreated)
		go func() {
			time.Sleep(500 * time.Microsecond) // 等待路由函数返回
			ctx.Write([]byte("123"))
		}()

		c, cancel := context.WithCancel(context.Background())
		time.AfterFunc(500*time.Millisecond, func() { cancel() })
		return KeepAlive(c)
	})

	defer servertest.Run(a, srv)()
	defer srv.Close(0)

	servertest.Get(a, "http://localhost:8080/path").
		Header("Accept", "application/json").
		Do(nil).
		Status(http.StatusCreated).
		StringBody(`123`)
}

// TODO 部分功能移至 server
func testServer_NewContext(t *testing.T) {
	a := assert.New(t, false)
	srv := newTestServer(a)
	router := srv.NewRouter("def", nil)
	router.Get("/path", func(ctx *Context) Responser { ctx.Render(http.StatusOK, nil); return nil })

	defer servertest.Run(a, srv)()
	defer srv.Close(0)

	// 正常，指定 Accept-Language，采用默认的 accept
	var b1 time.Time
	srv.logBuf.Reset()
	router.Get("/p3", func(ctx *Context) Responser {
		a.NotNil(ctx).NotEmpty(ctx.ID())
		a.True(header.CharsetIsNop(ctx.inputCharset)).
			Equal(ctx.Mimetype(false), "application/json").
			Equal(ctx.outputCharsetName, "utf-8").
			Equal(ctx.inputMimetype, UnmarshalFunc(json.Unmarshal)).
			Equal(ctx.LanguageTag(), language.SimplifiedChinese).
			NotNil(ctx.LocalePrinter())
		b1 = ctx.Begin()
		return nil
	})
	servertest.Get(a, "http://localhost:8080/p3").
		Header(header.AcceptLang, "cmn-hans").
		Header(header.Accept, "application/json").
		Do(nil).
		Success()
	a.Empty(srv.logBuf.String())

	// 正常，未指定 Accept-Language 和 Accept-Charset 等不是必须的报头
	srv.logBuf.Reset()
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
	a.Empty(srv.logBuf.String())

	// 正常，未指定 Accept-Language 和 Accept-Charset 等不是必须的报头，且有输入内容
	srv.logBuf.Reset()
	router.Post("/p5", func(ctx *Context) Responser {
		a.NotNil(ctx).
			True(header.CharsetIsNop(ctx.inputCharset)).
			Equal(ctx.Mimetype(false), "application/json").
			Equal(ctx.outputCharsetName, header.UTF8Name)
		ctx.WriteHeader(http.StatusCreated)
		return nil
	})
	servertest.Post(a, "http://localhost:8080/p5", []byte("123")).
		Header("accept", "application/json").
		Do(nil).
		Status(http.StatusCreated)
	a.Empty(srv.logBuf.String())
}

func TestContext_SetMimetype(t *testing.T) {
	a := assert.New(t, false)
	srv := newTestServer(a)

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
	srv := newTestServer(a)

	w := httptest.NewRecorder()
	r := rest.Post(a, "/path", []byte("123")).
		Header("accept", "application/json").
		Request()
	ctx := srv.NewContext(w, r)
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
	srv := newTestServer(a)

	w := httptest.NewRecorder()
	r := rest.Post(a, "/path", []byte("123")).
		Header("accept", "application/json").
		Request()
	ctx := srv.NewContext(w, r)
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
	srv := newTestServer(a)

	w := httptest.NewRecorder()
	r := rest.Post(a, "/path", []byte("123")).
		Header("accept", "application/json").
		Request()
	ctx := srv.NewContext(w, r)
	a.NotNil(ctx)

	a.Equal(ctx.LanguageTag(), ctx.Server().Language())

	cmnHant := language.MustParse("cmn-hant")
	ctx.SetLanguage(cmnHant)
	a.Equal(ctx.LanguageTag(), cmnHant)
}

func TestContext_IsXHR(t *testing.T) {
	a := assert.New(t, false)

	srv := newTestServer(a)
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

	srv := newTestServer(a)
	b := srv.Catalog()
	a.NotError(b.SetString(language.Und, "lang", "und"))
	a.NotError(b.SetString(language.SimplifiedChinese, "lang", "hans"))
	a.NotError(b.SetString(language.TraditionalChinese, "lang", "hant"))
	a.NotError(b.SetString(language.AmericanEnglish, "lang", "en_US"))

	tag := acceptLanguage(srv, "")
	a.Equal(tag, srv.language, "v1:%s, v2:%s", tag.String(), language.Und.String())

	tag = acceptLanguage(srv, "zh") // 匹配 zh-hans
	a.Equal(tag, language.SimplifiedChinese, "v1:%s, v2:%s", tag.String(), language.SimplifiedChinese.String())

	tag = acceptLanguage(srv, "zh-Hant")
	a.Equal(tag, language.TraditionalChinese, "v1:%s, v2:%s", tag.String(), language.TraditionalChinese.String())

	tag = acceptLanguage(srv, "zh-Hans")
	a.Equal(tag, language.SimplifiedChinese, "v1:%s, v2:%s", tag.String(), language.SimplifiedChinese.String())

	tag = acceptLanguage(srv, "english") // english 非正确的 tag，但是常用。
	a.Equal(tag, language.AmericanEnglish, "v1:%s, v2:%s", tag.String(), language.AmericanEnglish.String())

	tag = acceptLanguage(srv, "zh-Hans;q=0.1,zh-Hant;q=0.3,en")
	a.Equal(tag, language.AmericanEnglish, "v1:%s, v2:%s", tag.String(), language.AmericanEnglish.String())
}

func TestContext_ClientIP(t *testing.T) {
	a := assert.New(t, false)
	r := rest.Post(a, "/path", nil).Request()
	ctx := newContext(a, nil, r)
	a.NotNil(ctx)
	a.Equal(ctx.ClientIP(), r.RemoteAddr)
}

// 检测 204 是否存在 http: request method or response status code does not allow body
func TestContext_NoContent(t *testing.T) {
	a := assert.New(t, false)
	s := newTestServer(a)

	s.NewRouter("def", nil).Get("/204", func(ctx *Context) Responser {
		return ResponserFunc(func(ctx *Context) Problem {
			ctx.WriteHeader(http.StatusNoContent)
			return nil
		})
	})

	defer servertest.Run(a, s)()

	servertest.Get(a, "http://localhost:8080/204").
		Header("Accept-Encoding", "gzip"). // 服务端不应该构建压缩对象
		Header("Accept", "application/json").
		Do(nil).
		Status(http.StatusNoContent)

	s.Close(0)

	a.NotContains(s.logBuf.String(), "request method or response status code does not allow body")
}
