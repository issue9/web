// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package web

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/issue9/assert/v6"
	"github.com/issue9/assert/v6/rest"
	"github.com/issue9/mux/v10/header"
	"github.com/issue9/mux/v10/types"
	"golang.org/x/text/language"
)

var (
	_ http.ResponseWriter = &Context{}
	_ context.Context     = &Context{}
)

func (ctx *Context) apply(r Responser) { r.Apply(ctx) }

func newContext(a *assert.Assertion, w http.ResponseWriter, r *http.Request) *Context {
	if w == nil {
		w = httptest.NewRecorder()
	}
	if r == nil {
		r := httptest.NewRequest(http.MethodGet, "/path", bytes.NewBufferString("123"))
		r.Header.Set(header.Accept, "*/*")
	}

	return newTestServer(a, false).NewContext(w, r, types.NewRoute())
}

func TestContext_KeepAlive(t *testing.T) {
	a := assert.New(t, false)
	s := newTestServer(a, false)
	ctx := s.NewContext(httptest.NewRecorder(), rest.Get(a, "/p").Request(), types.NewRoute())
	dur := 500 * time.Millisecond
	begin := time.Now()

	c, cancel := context.WithCancel(context.Background())
	time.AfterFunc(dur, func() { cancel() })
	KeepAlive(c).Apply(ctx)
	a.True(time.Since(begin) >= 500*time.Millisecond)
}

func TestNewContext(t *testing.T) {
	a := assert.New(t, false)
	s := newTestServer(a, false)

	t.Run("unset request id key", func(*testing.T) {
		w := httptest.NewRecorder()
		r := rest.Get(a, "/path").Request()

		s.NewContext(w, r, types.NewRoute())
		a.NotEmpty(w.Header().Get(header.XRequestID))
	})

	t.Run("set request id key", func(*testing.T) {
		w := httptest.NewRecorder()
		r := rest.Get(a, "/path").Header(header.XRequestID, "111").Request()

		s.NewContext(w, r, types.NewRoute())
		a.Equal(w.Header().Get(header.XRequestID), "111")
	})

	t.Run("accept", func(*testing.T) {
		w := httptest.NewRecorder()
		r := rest.Get(a, "/path").Header(header.Accept, "111").Request()

		s.NewContext(w, r, types.NewRoute())
		a.Equal(w.Result().StatusCode, http.StatusNotAcceptable)
	})

	t.Run("accept-charset", func(*testing.T) {
		w := httptest.NewRecorder()
		r := rest.Get(a, "/path").Header(header.AcceptCharset, "111").Request()

		s.NewContext(w, r, types.NewRoute())
		a.Equal(w.Result().StatusCode, http.StatusNotAcceptable)
	})

	t.Run("accept-encoding", func(*testing.T) {
		w := httptest.NewRecorder()
		r := rest.Get(a, "/path").Header(header.AcceptEncoding, "*;q=0").Request()

		s.NewContext(w, r, types.NewRoute())
		a.Equal(w.Result().StatusCode, http.StatusOK)
	})

	t.Run("content-type", func(*testing.T) {
		w := httptest.NewRecorder()
		r := rest.Get(a, "/path").Header(header.ContentType, "111").Request()

		s.NewContext(w, r, types.NewRoute())
		a.Equal(w.Result().StatusCode, http.StatusUnsupportedMediaType)
	})
}

func TestContext_SetMediaType(t *testing.T) {
	a := assert.New(t, false)
	srv := newTestServer(a, false)

	w := httptest.NewRecorder()
	r := rest.Post(a, "/path", []byte("123")).Header(header.Accept, header.JSON).Request()
	ctx := srv.NewContext(w, r, types.NewRoute())
	a.NotNil(ctx)

	a.PanicString(func() {
		ctx.SetMediaType("not-exists")
	}, "指定的编码 not-exists 不存在").
		Equal(ctx.MediaType(false), header.JSON) // 不改变原有的值

	ctx.SetMediaType(header.XML)
	a.Equal(ctx.MediaType(false), header.XML)

	ctx.Render(200, 200) // 输出内容
	a.PanicString(func() {
		ctx.SetMediaType(header.JSON)
	}, "已有内容输出，不可再更改！")
}

func TestContext_SetCharset(t *testing.T) {
	a := assert.New(t, false)
	srv := newTestServer(a, false)

	w := httptest.NewRecorder()
	r := rest.Post(a, "/path", []byte("123")).Header(header.Accept, header.JSON).Request()
	ctx := srv.NewContext(w, r, types.NewRoute())
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
	srv := newTestServer(a, false)

	w := httptest.NewRecorder()
	r := rest.Post(a, "/path", bytes.NewBufferString("123").Bytes()).Header(header.Accept, header.JSON).Request()
	ctx := srv.NewContext(w, r, types.NewRoute())
	a.NotNil(ctx)

	t.Run("出错", func(t *testing.T) {
		a := assert.New(t, false)
		a.PanicString(func() {
			ctx.SetEncoding("*;q=x.0")
		}, "解析参数时发生了错误")

		_, err := ctx.Write([]byte("200"))
		a.NotError(err).
			PanicString(func() {
				ctx.SetEncoding("br")
			}, "已有内容输出，不可再更改！")
	})

	t.Run("setEncoding()", func(t *testing.T) {
		a := a.New(t)
		w := httptest.NewRecorder()
		r = rest.Get(a, "/path").Header(header.Accept, header.JSON).Header(header.AcceptEncoding, "gzip").Request()
		ctx := srv.NewContext(w, r, types.NewRoute())

		ctx.SetEncoding("") // 相当于 *
		a.Empty(w.Header().Get(header.ContentEncoding))
		_, err := ctx.Write([]byte("plain"))
		a.NotError(err)
		ctx.freeContext()

		reader, err := gzip.NewReader(bytes.NewReader(w.Body.Bytes()))
		if err != nil {
			t.Fatal(err)
		}
		data, err := io.ReadAll(reader)
		a.NotError(err).Equal(string(data), "plain")
		a.NotError(reader.Close())
	})

	t.Run("setEncoding(identity)", func(t *testing.T) {
		a := a.New(t)
		w := httptest.NewRecorder()
		r = rest.Get(a, "/path").Header(header.Accept, header.JSON).Header(header.AcceptEncoding, "gzip").Request()
		ctx := srv.NewContext(w, r, types.NewRoute())

		ctx.SetEncoding(header.Identity)
		a.Empty(w.Header().Get(header.ContentEncoding))
		_, err := ctx.Write([]byte("plain"))
		a.NotError(err)
		ctx.freeContext()
		a.Equal(w.Body.String(), "plain")
	})

	t.Run("setEncoding(compress)", func(t *testing.T) {
		a := assert.New(t, false)
		w := httptest.NewRecorder()
		r := rest.Get(a, "/path").Header(header.Accept, header.JSON).Request()
		ctx := srv.NewContext(w, r, types.NewRoute())

		ctx.SetEncoding("deflate")
		a.Equal(ctx.Encoding(), "deflate")
		_, err := ctx.Write([]byte("plain"))
		a.NotError(err)
		ctx.freeContext()

		reader := flate.NewReader(bytes.NewReader(w.Body.Bytes()))
		data, err := io.ReadAll(reader)
		a.NotError(err).NotNil(data).Equal("plain", string(data))
	})
}

func TestContext_SetLanguage(t *testing.T) {
	a := assert.New(t, false)
	srv := newTestServer(a, false)

	w := httptest.NewRecorder()
	r := rest.Post(a, "/path", []byte("123")).Header(header.Accept, header.JSON).Request()
	ctx := srv.NewContext(w, r, types.NewRoute())
	a.NotNil(ctx)

	a.Equal(ctx.LanguageTag(), ctx.Server().Locale().ID())

	cmnHant := language.MustParse("cmn-hant")
	ctx.SetLanguage(cmnHant)
	a.Equal(ctx.LanguageTag(), cmnHant)
}

func TestContext_IsXHR(t *testing.T) {
	a := assert.New(t, false)
	s := newTestServer(a, false)

	w := httptest.NewRecorder()
	r := rest.Get(a, "/p").Request()
	ctx := s.NewContext(w, r, types.NewRoute())
	a.False(ctx.IsXHR())

	w = httptest.NewRecorder()
	r = rest.Get(a, "/p").Request()
	r.Header.Set(header.XRequestedWith, "XMLHttpRequest")
	ctx = s.NewContext(w, r, types.NewRoute())
	a.True(ctx.IsXHR())
}

func TestContext_IsIdempotent(t *testing.T) {
	a := assert.New(t, false)
	s := newTestServer(a, false)

	w := httptest.NewRecorder()
	r := rest.Get(a, "/p").Request()
	ctx := s.NewContext(w, r, types.NewRoute())
	a.True(ctx.IsIdempotent())

	w = httptest.NewRecorder()
	r = rest.Patch(a, "/p", nil).Request()
	ctx = s.NewContext(w, r, types.NewRoute())
	a.False(ctx.IsIdempotent())
}

func TestServer_acceptLanguage(t *testing.T) {
	a := assert.New(t, false)

	srv := newTestServer(a, false)
	b := srv.Locale()
	a.NotError(b.SetString(language.Und, "lang", "und"))
	a.NotError(b.SetString(language.SimplifiedChinese, "lang", "hans"))
	a.NotError(b.SetString(language.TraditionalChinese, "lang", "hant"))
	a.NotError(b.SetString(language.AmericanEnglish, "lang", "en_US"))

	tag := acceptLanguage(srv.Server, "")
	a.Equal(tag, srv.Locale().ID(), "v1:%s, v2:%s", tag.String(), language.Und.String())

	tag = acceptLanguage(srv.Server, "zh") // 匹配 zh-hans
	a.Equal(tag, language.SimplifiedChinese, "v1:%s, v2:%s", tag.String(), language.SimplifiedChinese.String())

	tag = acceptLanguage(srv.Server, "zh-Hant")
	a.Equal(tag, language.TraditionalChinese, "v1:%s, v2:%s", tag.String(), language.TraditionalChinese.String())

	tag = acceptLanguage(srv.Server, "zh-Hans")
	a.Equal(tag, language.SimplifiedChinese, "v1:%s, v2:%s", tag.String(), language.SimplifiedChinese.String())

	tag = acceptLanguage(srv.Server, "english") // english 非正确的 tag，但是常用。
	a.Equal(tag, language.AmericanEnglish, "v1:%s, v2:%s", tag.String(), language.AmericanEnglish.String())

	tag = acceptLanguage(srv.Server, "zh-Hans;q=0.1,zh-Hant;q=0.3,en")
	a.Equal(tag, language.AmericanEnglish, "v1:%s, v2:%s", tag.String(), language.AmericanEnglish.String())
}

func TestContext_ClientIP(t *testing.T) {
	a := assert.New(t, false)
	r := rest.Post(a, "/path", []byte("123")).Request()
	ctx := newContext(a, nil, r)
	a.NotNil(ctx).Equal(ctx.ClientIP(), r.RemoteAddr)
}

func TestContext_time(t *testing.T) {
	a := assert.New(t, false)

	r := rest.Post(a, "/path", []byte("123")).Request()
	ctx := newContext(a, nil, r)

	a.Equal(ctx.Location(), ctx.Server().Location()).
		Equal(ctx.Now().Location(), ctx.Location()).
		Equal(ctx.Begin().Location(), ctx.Location())
}
