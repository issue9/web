// SPDX-License-Identifier: MIT

package web

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/issue9/assert/v3"
	"golang.org/x/text/language"

	"github.com/issue9/web/internal/header"
)

var _ http.ResponseWriter = &Context{}

func (ctx *Context) apply(r Responser) {
	if p := r.Apply(ctx); p != nil {
		p.Apply(ctx) // Problem.Apply 始终返回 nil
	}
}

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
	s := newTestServer(a)
	ctx := s.NewContext(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/p", nil))
	dur := 500 * time.Millisecond
	begin := time.Now()

	c, cancel := context.WithCancel(context.Background())
	time.AfterFunc(dur, func() { cancel() })
	KeepAlive(c).Apply(ctx)
	a.True(time.Since(begin) >= 500*time.Millisecond)
}

func TestNewContext(t *testing.T) {
	a := assert.New(t, false)
	s := newTestServer(a)

	t.Run("unset request id key", func(*testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/path", nil)

		s.NewContext(w, r)
		a.NotEmpty(w.Header().Get(header.RequestIDKey))
	})

	t.Run("set request id key", func(*testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/path", nil)
		r.Header.Set(header.RequestIDKey, "111")

		s.NewContext(w, r)
		a.Equal(w.Header().Get(header.RequestIDKey), "111")
	})

	t.Run("accept", func(*testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/path", nil)
		r.Header.Set(header.Accept, "111")

		s.NewContext(w, r)
		a.Equal(w.Result().StatusCode, http.StatusNotAcceptable)
	})

	t.Run("accept-charset", func(*testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/path", nil)
		r.Header.Set(header.AcceptCharset, "111")

		s.NewContext(w, r)
		a.Equal(w.Result().StatusCode, http.StatusNotAcceptable)
	})

	t.Run("accept-encoding", func(*testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/path", nil)
		r.Header.Set(header.AcceptEncoding, "111")

		s.NewContext(w, r)
		a.Equal(w.Result().StatusCode, http.StatusNotAcceptable)
	})

	t.Run("content-type", func(*testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/path", nil)
		r.Header.Set(header.ContentType, "111")

		s.NewContext(w, r)
		a.Equal(w.Result().StatusCode, http.StatusUnsupportedMediaType)
	})
}

func TestContext_SetMimetype(t *testing.T) {
	a := assert.New(t, false)
	srv := newTestServer(a)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/path", bytes.NewBufferString("123"))
	r.Header.Set(header.Accept, "application/json")
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
	r := httptest.NewRequest(http.MethodPost, "/path", bytes.NewBufferString("123"))
	r.Header.Set(header.Accept, "application/json")
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
	r := httptest.NewRequest(http.MethodPost, "/path", bytes.NewBufferString("123"))
	r.Header.Set(header.Accept, "application/json")
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
	r := httptest.NewRequest(http.MethodPost, "/path", bytes.NewBufferString("123"))
	r.Header.Set(header.Accept, "application/json")
	ctx := srv.NewContext(w, r)
	a.NotNil(ctx)

	a.Equal(ctx.LanguageTag(), ctx.Server().Language())

	cmnHant := language.MustParse("cmn-hant")
	ctx.SetLanguage(cmnHant)
	a.Equal(ctx.LanguageTag(), cmnHant)
}

func TestContext_IsXHR(t *testing.T) {
	a := assert.New(t, false)
	s := newTestServer(a)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/p", nil)
	ctx := s.NewContext(w, r)
	a.False(ctx.IsXHR())

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/p", nil)
	r.Header.Set("X-Requested-With", "XMLHttpRequest")
	ctx = s.NewContext(w, r)
	a.True(ctx.IsXHR())
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
	r := httptest.NewRequest(http.MethodPost, "/path", bytes.NewBufferString("123"))
	ctx := newContext(a, nil, r)
	a.NotNil(ctx)
	a.Equal(ctx.ClientIP(), r.RemoteAddr)
}
