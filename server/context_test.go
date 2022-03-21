// SPDX-License-Identifier: MIT

package server

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
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
	"github.com/issue9/logs/v3"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/language"

	"github.com/issue9/web/serialization"
	"github.com/issue9/web/serialization/gob"
	"github.com/issue9/web/serialization/text"
	"github.com/issue9/web/serialization/text/testobject"
)

const (
	gbkString1 = "中文1,11"
	gbkString2 = "中文2,22"
)

var (
	gbkBytes1 = []byte{214, 208, 206, 196, 49, 44, 49, 49}
	gbkBytes2 = []byte{214, 208, 206, 196, 50, 44, 50, 50}
)

var _ http.ResponseWriter = &Context{}

func newServer(a *assert.Assertion, o *Options) *Server {
	if o == nil {
		o = &Options{Port: ":8080"}
	}
	if o.Logs == nil { // 默认重定向到 os.Stderr
		l, err := logs.New(nil)
		a.NotError(err).NotNil(l)

		a.NotError(l.SetOutput(logs.LevelDebug|logs.LevelError|logs.LevelCritical, os.Stderr))
		a.NotError(l.SetOutput(logs.LevelInfo|logs.LevelTrace|logs.LevelWarn, os.Stdout))
		o.Logs = l
	}

	srv, err := New("app", "0.1.0", o)
	a.NotError(err).NotNil(srv)
	a.Equal(srv.Name(), "app").Equal(srv.Version(), "0.1.0")

	// locale
	b := srv.Locale().Builder()
	a.NotError(b.SetString(language.Und, "lang", "und"))
	a.NotError(b.SetString(language.SimplifiedChinese, "lang", "hans"))
	a.NotError(b.SetString(language.TraditionalChinese, "lang", "hant"))

	// mimetype
	a.NotError(srv.Mimetypes().Add(json.Marshal, json.Unmarshal, "application/json"))
	a.NotError(srv.Mimetypes().Add(xml.Marshal, xml.Unmarshal, "application/xml"))
	a.NotError(srv.Mimetypes().Add(gob.Marshal, gob.Unmarshal, DefaultMimetype))
	a.NotError(srv.Mimetypes().Add(text.Marshal, text.Unmarshal, text.Mimetype))
	a.NotError(srv.Mimetypes().Add(nil, nil, "nil"))

	srv.AddResult(411, "41110", localeutil.Phrase("41110"))

	// encoding
	srv.Encodings().Add(map[string]serialization.EncodingWriterFunc{
		"gzip": func(w io.Writer) (serialization.WriteCloseRester, error) {
			return gzip.NewWriter(w), nil
		},
		"deflate": func(w io.Writer) (serialization.WriteCloseRester, error) {
			return flate.NewWriter(&bytes.Buffer{}, flate.DefaultCompression)
		},
	})

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
	ctx := newServer(a, nil).NewContext(w, r)

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

func TestServer_NewContext(t *testing.T) {
	a := assert.New(t, false)
	lw := &bytes.Buffer{}
	srv := newServer(a, &Options{Tag: language.SimplifiedChinese})
	a.NotError(srv.Logs().SetOutput(logs.LevelDebug, lw))

	// 错误的 accept
	w := httptest.NewRecorder()
	r := rest.Get(a, "/path").Header("Accept", "not").Request()
	srv.NewContext(w, r)
	a.Equal(w.Code, http.StatusNotAcceptable)
	a.Contains(lw.String(), localeutil.Error("not found serialization for %s", "not").Error())

	// 错误的 accept-charset
	lw.Reset()
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").Header("Accept", "not").
		Header("Accept", text.Mimetype).
		Header("Accept-Charset", "unknown").
		Request()
	srv.NewContext(w, r)
	a.Equal(w.Code, http.StatusNotAcceptable)
	a.Contains(lw.String(), localeutil.Error("not found charset for %s", "unknown").Error())

	// 错误的 content-type,无输入内容
	lw.Reset()
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").Header("Content-Type", ";charset=utf-8").Request()
	srv.NewContext(w, r)
	a.Equal(w.Code, http.StatusUnsupportedMediaType)
	a.NotEmpty(lw.String())

	// 错误的 content-type,有输入内容
	lw.Reset()
	w = httptest.NewRecorder()

	r = rest.Post(a, "/path", []byte("[]")).Header("Content-Type", ";charset=utf-8").Request()
	srv.NewContext(w, r)
	a.Equal(w.Code, http.StatusUnsupportedMediaType)
	a.NotEmpty(lw.String())

	// 错误的 content-type，且有输入内容
	lw.Reset()
	w = httptest.NewRecorder()
	r = rest.Post(a, "/path", []byte("123")).Header("Content-Type", buildContentType(text.Mimetype, "utf-")).Request()
	r.Header.Set("content-type", buildContentType(text.Mimetype, "utf-"))
	srv.NewContext(w, r)
	a.Equal(w.Code, http.StatusUnsupportedMediaType)
	a.NotEmpty(lw.String())

	// 部分错误的 Accept-Language
	lw.Reset()
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").
		Header("Accept-Language", "zh-hans;q=0.9,zh-Hant;q=xxx").
		Header("content-type", buildContentType(text.Mimetype, DefaultCharset)).
		Request()
	ctx := srv.NewContext(w, r)
	a.NotNil(ctx)
	a.Equal(ctx.languageTag, language.MustParse("zh-hans"))
	a.Empty(lw.String())

	// 正常，指定 Accept-Language，采用默认的 accept
	lw.Reset()
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").
		Header("Accept-Language", "zh-hans;q=0.9,zh-Hant;q=0.7").
		Header("content-type", buildContentType(text.Mimetype, DefaultCharset)).
		Request()
	ctx = srv.NewContext(w, r)
	a.NotNil(ctx)
	a.Empty(lw.String())
	a.True(charsetIsNop(ctx.inputCharset)).
		Equal(ctx.contentType, "application/json; charset=utf-8").
		Equal(ctx.inputMimetype, serialization.UnmarshalFunc(text.Unmarshal)).
		Equal(ctx.languageTag, language.SimplifiedChinese).
		NotNil(ctx.LocalePrinter())

	// 正常，未指定 Accept-Language 和 Accept-Charset 等不是必须的报头
	lw.Reset()
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").
		Header("content-type", buildContentType(text.Mimetype, DefaultCharset)).
		Header("accept", "application/json;q=0.2,text/plain;q=0.9").
		Request()
	ctx = srv.NewContext(w, r)
	a.Empty(lw.String())
	a.NotNil(ctx).
		True(charsetIsNop(ctx.inputCharset)).
		Equal(ctx.contentType, buildContentType(text.Mimetype, DefaultCharset))

	// 正常，未指定 Accept-Language 和 Accept-Charset 等不是必须的报头，且有输入内容
	lw.Reset()
	w = httptest.NewRecorder()
	r = rest.Post(a, "/path", []byte("123")).
		Header("content-type", buildContentType(text.Mimetype, DefaultCharset)).
		Header("accept", "application/json;q=0.2,text/*;q=0.9").
		Request()
	ctx = srv.NewContext(w, r)
	a.Empty(lw.String())
	a.NotNil(ctx).
		True(charsetIsNop(ctx.inputCharset)).
		Equal(ctx.contentType, buildContentType(text.Mimetype, DefaultCharset))
}

func TestContext_Body(t *testing.T) {
	a := assert.New(t, false)
	srv := newServer(a, &Options{Tag: language.SimplifiedChinese})

	// 未缓存
	r := rest.Post(a, "/path", []byte("123")).Request()
	w := httptest.NewRecorder()
	ctx := srv.NewContext(w, r)
	a.Nil(ctx.body)
	data, err := ctx.Body()
	a.NotError(err).Equal(data, []byte("123"))
	a.Equal(ctx.body, data)

	// 读取缓存内容
	data, err = ctx.Body()
	a.NotError(err).Equal(data, []byte("123"))
	a.Equal(ctx.body, data)

	// 采用 Nop 即 utf-8 编码
	r = rest.Post(a, "/path", []byte("123")).Request()
	ctx = srv.NewContext(w, r)
	ctx.inputCharset = encoding.Nop
	data, err = ctx.Body()
	a.NotError(err).Equal(data, []byte("123"))
	a.Equal(ctx.body, data)

	// 采用不同的编码
	r = rest.Post(a, "/path", gbkBytes1).
		Header("Content-type", "text/plain;charset=gb18030").
		Request()
	w = httptest.NewRecorder()
	ctx = srv.NewContext(w, r)
	data, err = ctx.Body()
	a.NotError(err).Equal(string(data), gbkString1)
	a.Equal(ctx.body, data)

	// 采用不同的编码
	w = httptest.NewRecorder()
	r = rest.Post(a, "/path", gbkBytes1).
		Header("Accept", "*/*").
		Header("Content-Type", buildContentType(text.Mimetype, " gb18030")).
		Request()
	ctx = srv.NewContext(w, r)
	a.NotNil(ctx)
	data, err = ctx.Body()
	a.NotError(err).Equal(string(data), gbkString1)
	a.Equal(ctx.body, data)
}

func TestContext_Unmarshal(t *testing.T) {
	a := assert.New(t, false)
	srv := newServer(a, nil)

	r := rest.Post(a, "/path", []byte("test,123")).
		Header("content-type", text.Mimetype).
		Request()
	w := httptest.NewRecorder()
	ctx := srv.NewContext(w, r)

	obj := &testobject.TextObject{}
	a.NotError(ctx.Unmarshal(obj))
	a.Equal(obj.Name, "test").Equal(obj.Age, 123)

	// 无法转换
	o := &struct{}{}
	a.Error(ctx.Unmarshal(o))

	// 空提交
	r = rest.Post(a, "/path", nil).
		Header("content-type", text.Mimetype).
		Request()
	w = httptest.NewRecorder()
	ctx = srv.NewContext(w, r)
	obj = &testobject.TextObject{}
	a.NotError(ctx.Unmarshal(obj))
	a.Equal(obj.Name, "").Equal(obj.Age, 0)
}

func TestContext_Marshal(t *testing.T) {
	a := assert.New(t, false)
	srv := newServer(a, &Options{Tag: language.SimplifiedChinese})

	// 自定义报头
	w := httptest.NewRecorder()
	r := rest.Post(a, "/path", nil).
		Header("Content-Type", text.Mimetype).
		Header("Accept", text.Mimetype).
		Request()
	ctx := srv.NewContext(w, r)
	a.NotNil(ctx)
	obj := &testobject.TextObject{Name: "test", Age: 123}
	a.NotError(ctx.Marshal(http.StatusCreated, obj, map[string]string{"contEnt-type": "json", "content-lanGuage": "zh-hant", "content-encoding": "123"}))
	a.Equal(w.Code, http.StatusCreated)
	a.Equal(w.Body.String(), "test,123")
	a.Equal(w.Header().Get("content-type"), buildContentType(text.Mimetype, "utf-8"))
	a.Equal(w.Header().Get("content-language"), "zh-Hans")
	a.Equal(w.Header().Get("content-encoding"), "123") // 未指定，所有采用 headers 参数的值

	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").
		Header("accept", text.Mimetype).
		Header("accept-language", "").
		Request()
	ctx = srv.NewContext(w, r)
	obj = &testobject.TextObject{Name: "test", Age: 1234}
	a.NotError(ctx.Marshal(http.StatusCreated, obj, nil))
	a.Equal(w.Code, http.StatusCreated)
	a.Equal(w.Body.String(), "test,1234")
	a.Equal(w.Header().Get("content-language"), language.SimplifiedChinese.String()) // 未指定，采用默认值

	// 输出 nil，content-type 和 content-language 均为空
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").
		Header("Accept", text.Mimetype).
		Header("Accept-language", "zh-hans").
		Request()
	ctx = srv.NewContext(w, r)
	a.NotError(ctx.Marshal(http.StatusCreated, nil, nil))
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
	ctx = srv.NewContext(w, r)
	a.NotError(ctx.Marshal(http.StatusCreated, gbkString2, nil))
	a.Equal(w.Code, http.StatusCreated)
	a.Equal(w.Body.Bytes(), gbkBytes2)

	// 同时指定了 accept,accept-language,accept-charset 和 accept-encoding
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").
		Header("Accept", text.Mimetype).
		Header("Accept-Language", "zh-Hans").
		Header("Accept-Charset", "gbk").
		Header("Accept-Encoding", "gzip;q=0.9,deflate").
		Request()
	ctx = srv.NewContext(w, r)
	a.NotError(ctx.Marshal(http.StatusCreated, gbkString2, map[string]string{"content-encoding": "123"}))
	a.NotError(ctx.destroy())
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
	ctx = srv.NewContext(w, r)
	_, err = ctx.Write([]byte("123"))
	a.NotError(err)
	a.NotError(ctx.Marshal(http.StatusCreated, "456", nil))
	a.NotError(ctx.destroy())
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
	ctx = srv.NewContext(w, r)
	_, err = ctx.Write([]byte(gbkString1))
	a.NotError(err)
	a.NotError(ctx.Marshal(http.StatusCreated, gbkString2, nil))
	a.NotError(ctx.destroy())
	a.Equal(w.Code, http.StatusOK) // 未指定压缩，WriteHeader 直接发送
	data, err = io.ReadAll(w.Body)
	a.NotError(err)
	bs := make([]byte, 0, len(gbkBytes1)+len(gbkBytes2))
	bs = append(append(bs, gbkBytes1...), gbkBytes2...)
	a.Equal(data, bs)

	// outputMimetype == nil
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").Header("Accept", "nil").Request()
	ctx = srv.NewContext(w, r)
	a.Nil(ctx.outputMimetype).Equal(ctx.contentType, buildContentType("nil", DefaultCharset))
	a.NotError(ctx.Marshal(http.StatusCreated, "val", nil))
	a.Equal(w.Code, http.StatusNotAcceptable)

	// outputMimetype 返回 serialization.ErrUnsupported
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").Header("Accept", text.Mimetype).Request()
	ctx = srv.NewContext(w, r)
	a.ErrorIs(ctx.Marshal(http.StatusCreated, &struct{}{}, nil), serialization.ErrUnsupported)
	a.Equal(w.Code, http.StatusNotAcceptable)

	// outputMimetype 返回错误
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").Header("Accept", text.Mimetype).Request()
	ctx = srv.NewContext(w, r)
	a.Error(ctx.Marshal(http.StatusCreated, errors.New("error"), nil))
	a.Equal(w.Code, http.StatusInternalServerError)
}

func TestContext_IsXHR(t *testing.T) {
	a := assert.New(t, false)

	srv := newServer(a, nil)
	router := srv.Routers().New("router", nil, &RouterOptions{URLDomain: "https://example.com"})
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

	srv := newServer(a, &Options{Tag: language.Afrikaans})
	b := srv.Locale().Builder()
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

func TestServer_contentType(t *testing.T) {
	a := assert.New(t, false)

	srv := newServer(a, &Options{Tag: language.SimplifiedChinese})
	a.NotNil(srv)

	f, e, err := srv.conentType(";;;")
	a.Error(err).Nil(f).Nil(e)

	// 不存在的 mimetype
	f, e, err = srv.conentType(buildContentType("not-exists", DefaultCharset))
	a.Equal(err, localeutil.Error("not found serialization function for %s", "not-exists")).Nil(f).Nil(e)

	f, e, err = srv.conentType(buildContentType(DefaultMimetype, DefaultCharset))
	a.NotError(err).NotNil(f).NotNil(e)

	// 无效的字符集名称
	f, e, err = srv.conentType(buildContentType(DefaultMimetype, "invalid-charset"))
	a.Error(err).Nil(f).Nil(e)
}

func TestServer_Location(t *testing.T) {
	a := assert.New(t, false)
	srv := newServer(a, nil)

	w := httptest.NewRecorder()
	r := rest.Get(a, "/test").Request()
	ctx := srv.NewContext(w, r)
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

func TestContext_Read(t *testing.T) {
	a := assert.New(t, false)

	w := httptest.NewRecorder()
	r := rest.Post(a, "/path", []byte("test,123")).
		Header("Content-Type", buildContentType(text.Mimetype, "utf-8")).
		Request()
	ctx := newServer(a, nil).NewContext(w, r)
	obj := &testobject.TextObject{}
	a.Nil(ctx.Read(obj, "41110"))
	a.Equal(obj.Name, "test").Equal(obj.Age, 123)

	w = httptest.NewRecorder()
	r = rest.Post(a, "/path", []byte("test,123")).
		Header("Content-Type", buildContentType(text.Mimetype, "utf-8")).
		Request()
	ctx = newServer(a, nil).NewContext(w, r)
	o := &struct{}{}
	resp := ctx.Read(o, "41110")
	a.NotNil(resp)
	a.NotError(resp.Apply(ctx))
	a.Equal(w.Code, http.StatusUnprocessableEntity)
}

func TestContext_ClientIP(t *testing.T) {
	a := assert.New(t, false)
	w := httptest.NewRecorder()

	r := rest.Post(a, "/path", nil).Request()
	ctx := newServer(a, nil).NewContext(w, r)
	a.Equal(ctx.ClientIP(), r.RemoteAddr)

	r = rest.Post(a, "/path", nil).Header("x-real-ip", "192.168.1.1:8080").Request()
	ctx = newServer(a, nil).NewContext(w, r)
	a.Equal(ctx.ClientIP(), "192.168.1.1:8080")

	r = rest.Post(a, "/path", nil).
		Header("x-forwarded-for", "192.168.2.1:8080,192.168.2.2:111").
		Request()
	ctx = newServer(a, nil).NewContext(w, r)
	a.Equal(ctx.ClientIP(), "192.168.2.1:8080")

	// 测试获取 IP 报头的优先级
	r = rest.Post(a, "/path", nil).
		Header("x-forwarded-for", "192.168.2.1:8080,192.168.2.2:111").
		Header("x-real-ip", "192.168.2.2").
		Request()
	ctx = newServer(a, nil).NewContext(w, r)
	a.Equal(ctx.ClientIP(), "192.168.2.1:8080")

	// 测试获取 IP 报头的优先级
	r = rest.Post(a, "/path", nil).
		Header("Remote-Addr", "192.168.2.0").
		Header("x-forwarded-for", "192.168.2.1:8080,192.168.2.2:111").
		Header("x-real-ip", "192.168.2.2").
		Request()
	ctx = newServer(a, nil).NewContext(w, r)
	a.Equal(ctx.ClientIP(), "192.168.2.1:8080")
}

func TestBuildContentType(t *testing.T) {
	a := assert.New(t, false)

	a.Equal("application/xml; charset=utf16", buildContentType("application/xml", "utf16"))
	a.Equal("application/xml; charset="+DefaultCharset, buildContentType("application/xml", ""))
	a.Equal(DefaultMimetype+"; charset="+DefaultCharset, buildContentType("", ""))
	a.Equal("application/xml; charset="+DefaultCharset, buildContentType("application/xml", ""))
}

func TestContext_LocalePrinter(t *testing.T) {
	a := assert.New(t, false)
	srv := newServer(a, &Options{Tag: language.SimplifiedChinese})

	b := srv.Locale().Builder()
	a.NotError(b.SetString(language.MustParse("cmn-hans"), "test", "测试"))
	a.NotError(b.SetString(language.MustParse("cmn-hant"), "test", "測試"))

	w := httptest.NewRecorder()
	r := rest.Get(a, "/path").
		Header("accept-language", "cmn-hant").
		Header("accept", text.Mimetype).
		Request()
	ctx := srv.NewContext(w, r)
	a.NotNil(ctx)
	a.NotError(ctx.Marshal(http.StatusOK, ctx.Sprintf("test"), nil))
	a.Equal(w.Body.String(), "測試")

	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").
		Header("accept-language", "cmn-hans").
		Header("accept", text.Mimetype).
		Request()
	ctx = srv.NewContext(w, r)
	a.NotNil(ctx)
	n, err := ctx.LocalePrinter().Fprintf(ctx, "test")
	a.NotError(err).Equal(n, len("测试"))
	a.Equal(w.Body.String(), "测试")
}

func TestAcceptCharset(t *testing.T) {
	a := assert.New(t, false)

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
