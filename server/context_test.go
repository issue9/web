// SPDX-License-Identifier: MIT

package server

import (
	"bytes"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/issue9/assert/v2"
	"github.com/issue9/assert/v2/rest"
	"github.com/issue9/localeutil"
	"github.com/issue9/logs/v3"
	"github.com/issue9/mux/v5/group"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/language"

	"github.com/issue9/web/internal/charsetdata"
	"github.com/issue9/web/serialization"
	"github.com/issue9/web/serialization/text"
	"github.com/issue9/web/serialization/text/testobject"
)

func TestContext_Vars(t *testing.T) {
	a := assert.New(t, false)
	r := httptest.NewRequest(http.MethodGet, "/path", bytes.NewBufferString("123"))
	r.Header.Set("Accept", "*/*")
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
	r := httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Set("Accept", "not")
	a.Panic(func() {
		srv.NewContext(w, r)
	})
	a.Equal(w.Code, http.StatusNotAcceptable)
	a.Contains(lw.String(), localeutil.Error("not found serialization for %s", "not").Error())

	// 错误的 accept-charset
	lw.Reset()
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Set("Accept", text.Mimetype)
	r.Header.Set("Accept-Charset", "unknown")
	a.Panic(func() {
		srv.NewContext(w, r)
	})
	a.Equal(w.Code, http.StatusNotAcceptable)
	a.Contains(lw.String(), localeutil.Error("not found charset for %s", "unknown").Error())

	// 错误的 content-type,无输入内容
	lw.Reset()
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Set("Content-Type", ";charset=utf-8")
	a.Panic(func() {
		srv.NewContext(w, r)
	})
	a.Equal(w.Code, http.StatusUnsupportedMediaType)
	a.NotEmpty(lw.String())

	// 错误的 content-type,有输入内容
	lw.Reset()
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodPost, "/path", bytes.NewBufferString("[]"))
	r.Header.Set("Content-Type", ";charset=utf-8")
	a.Panic(func() {
		srv.NewContext(w, r)
	})
	a.Equal(w.Code, http.StatusUnsupportedMediaType)
	a.NotEmpty(lw.String())

	// 错误的 content-type，且有输入内容
	lw.Reset()
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/path", bytes.NewBufferString("123"))
	r.Header.Set("content-type", buildContentType(text.Mimetype, "utf-"))
	a.Panic(func() {
		srv.NewContext(w, r)
	})
	a.Equal(w.Code, http.StatusUnsupportedMediaType)
	a.NotEmpty(lw.String())

	// 部分错误的 Accept-Language
	lw.Reset()
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Set("Accept-Language", "zh-hans;q=0.9,zh-Hant;q=xxx")
	r.Header.Set("content-type", buildContentType(text.Mimetype, DefaultCharset))
	ctx := srv.NewContext(w, r)
	a.NotNil(ctx)
	a.Equal(ctx.OutputTag, language.MustParse("zh-hans"))
	a.Empty(lw.String())

	// 正常，指定 Accept-Language，采用默认的 accept
	lw.Reset()
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Set("Accept-Language", "zh-hans;q=0.9,zh-Hant;q=0.7")
	r.Header.Set("content-type", buildContentType(text.Mimetype, DefaultCharset))
	ctx = srv.NewContext(w, r)
	a.NotNil(ctx)
	a.Empty(lw.String())
	a.True(charsetIsNop(ctx.InputCharset)).
		Equal(ctx.OutputMimetypeName, "application/json").
		Equal(ctx.InputMimetype, serialization.UnmarshalFunc(text.Unmarshal)).
		Equal(ctx.OutputTag, language.SimplifiedChinese).
		NotNil(ctx.LocalePrinter)

	// 正常，未指定 Accept-Language 和 Accept-Charset 等不是必须的报头
	lw.Reset()
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Set("content-type", buildContentType(text.Mimetype, DefaultCharset))
	r.Header.Set("accept", "application/json;q=0.2,text/plain;q=0.9")
	ctx = srv.NewContext(w, r)
	a.Empty(lw.String())
	a.NotNil(ctx).
		True(charsetIsNop(ctx.InputCharset)).
		Equal(ctx.OutputMimetypeName, text.Mimetype)

	// 正常，未指定 Accept-Language 和 Accept-Charset 等不是必须的报头，且有输入内容
	lw.Reset()
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/path", bytes.NewBufferString("123"))
	r.Header.Set("content-type", buildContentType(text.Mimetype, DefaultCharset))
	r.Header.Set("accept", "application/json;q=0.2,text/*;q=0.9")
	ctx = srv.NewContext(w, r)
	a.Empty(lw.String())
	a.NotNil(ctx).
		True(charsetIsNop(ctx.InputCharset)).
		Equal(ctx.OutputMimetypeName, text.Mimetype)
}

func TestContext_Body(t *testing.T) {
	a := assert.New(t, false)
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
	srv := newServer(a, &Options{Tag: language.SimplifiedChinese})
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/path", bytes.NewBuffer(charsetdata.GBKData1))
	r.Header.Set("Accept", "*/*")
	r.Header.Set("Content-Type", buildContentType(text.Mimetype, " gb18030"))
	ctx = srv.NewContext(w, r)
	a.NotNil(ctx)
	data, err = ctx.Body()
	a.NotError(err).Equal(string(data), charsetdata.GBKString1)
	a.Equal(ctx.body, data)
}

func TestContext_Unmarshal(t *testing.T) {
	a := assert.New(t, false)

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
	a := assert.New(t, false)
	srv := newServer(a, &Options{Tag: language.SimplifiedChinese})

	// 自定义报头
	r := httptest.NewRequest(http.MethodPost, "/path", nil)
	w := httptest.NewRecorder()
	r.Header.Set("Content-Type", text.Mimetype)
	r.Header.Set("Accept", text.Mimetype)
	ctx := srv.NewContext(w, r)
	a.NotNil(ctx)
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

func TestContext_IsXHR(t *testing.T) {
	a := assert.New(t, false)

	srv := newServer(a, nil)
	router := srv.NewRouter("router", "https://example.com", group.MatcherFunc(group.Any))
	a.NotNil(router)
	router.Get("/not-xhr", func(ctx *Context) Responser {
		a.False(ctx.IsXHR())
		return nil
	})
	router.Get("/xhr", func(ctx *Context) Responser {
		a.True(ctx.IsXHR())
		return nil
	})

	r := httptest.NewRequest(http.MethodGet, "/not-xhr", nil)
	w := httptest.NewRecorder()
	router.MuxRouter().ServeHTTP(w, r)

	r = httptest.NewRequest(http.MethodGet, "/xhr", nil)
	r.Header.Set("X-Requested-With", "XMLHttpRequest")
	w = httptest.NewRecorder()
	router.MuxRouter().ServeHTTP(w, r)
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
	a.ErrorString(err, "未注册的解码函数").Nil(f).Nil(e)

	f, e, err = srv.conentType(buildContentType(DefaultMimetype, DefaultCharset))
	a.NotError(err).NotNil(f).NotNil(e)

	// 无效的字符集名称
	f, e, err = srv.conentType(buildContentType(DefaultMimetype, "invalid-charset"))
	a.Error(err).Nil(f).Nil(e)
}

func TestServer_Location(t *testing.T) {
	a := assert.New(t, false)
	srv := newServer(a, &Options{Tag: language.SimplifiedChinese})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := srv.NewContext(w, r)

	now := ctx.Now()
	a.Equal(now.Location(), srv.Location())
}

func TestContext_Read(t *testing.T) {
	a := assert.New(t, false)
	r := httptest.NewRequest(http.MethodPost, "/path", bytes.NewBufferString("test,123"))
	w := httptest.NewRecorder()
	r.Header.Set("Content-Type", text.Mimetype+"; charset=utf-8")
	ctx := newServer(a, nil).NewContext(w, r)

	obj := &testobject.TextObject{}
	a.Nil(ctx.Read(obj, "41110"))
	a.Equal(obj.Name, "test").Equal(obj.Age, 123)

	o := &struct{}{}
	resp := ctx.Read(o, "41110")
	a.NotNil(resp)
	a.Equal(resp.Status(), http.StatusUnprocessableEntity)
}

func TestContext_ClientIP(t *testing.T) {
	a := assert.New(t, false)
	w := httptest.NewRecorder()

	r := httptest.NewRequest(http.MethodPost, "/path", nil)
	ctx := newServer(a, nil).NewContext(w, r)
	a.Equal(ctx.ClientIP(), r.RemoteAddr)

	// httptest.NewRequest 会直接将  remote-addr 赋值为 192.0.2.1 无法测试
	r, err := http.NewRequest(http.MethodPost, "/path", nil)
	a.NotError(err).NotNil(r)
	r.Header.Set("x-real-ip", "192.168.1.1:8080")
	ctx = newServer(a, nil).NewContext(w, r)
	a.Equal(ctx.ClientIP(), "192.168.1.1:8080")

	r = httptest.NewRequest(http.MethodPost, "/path", nil)
	r.Header.Set("x-forwarded-for", "192.168.2.1:8080,192.168.2.2:111")
	ctx = newServer(a, nil).NewContext(w, r)
	a.Equal(ctx.ClientIP(), "192.168.2.1:8080")

	// 测试获取 IP 报头的优先级
	r = httptest.NewRequest(http.MethodPost, "/path", nil)
	r.Header.Set("x-forwarded-for", "192.168.2.1:8080,192.168.2.2:111")
	r.Header.Set("x-real-ip", "192.168.2.2")
	ctx = newServer(a, nil).NewContext(w, r)
	a.Equal(ctx.ClientIP(), "192.168.2.1:8080")

	// 测试获取 IP 报头的优先级
	r = httptest.NewRequest(http.MethodPost, "/path", nil)
	r.Header.Set("Remote-Addr", "192.168.2.0")
	r.Header.Set("x-forwarded-for", "192.168.2.1:8080,192.168.2.2:111")
	r.Header.Set("x-real-ip", "192.168.2.2")
	ctx = newServer(a, nil).NewContext(w, r)
	a.Equal(ctx.ClientIP(), "192.168.2.1:8080")
}

func TestContext_ServeFile(t *testing.T) {
	a := assert.New(t, false)
	exit := make(chan bool, 1)

	s := newServer(a, nil)
	defer func() {
		a.NotError(s.Close(0))
		<-exit
	}()
	router := s.NewRouter("default", "http://localhost:8080/root", group.MatcherFunc(group.Any))
	a.NotNil(router)

	a.NotPanic(func() {
		router.Get("/path", func(ctx *Context) Responser {
			return ctx.ServeFile("./testdata/file1.txt", "index.html", map[string]string{"Test": "Test"})
		})

		router.Get("/index", func(ctx *Context) Responser {
			return ctx.ServeFile("./testdata", "file1.txt", map[string]string{"Test": "Test"})
		})

		router.Get("/not-exists", func(ctx *Context) Responser {
			// file1.text 不存在
			return ctx.ServeFile("./testdata/file1.text", "index.html", map[string]string{"Test": "Test"})
		})
	})

	go func() {
		a.Equal(s.Serve(), http.ErrServerClosed)
		exit <- true
	}()
	time.Sleep(500 * time.Millisecond)

	testDownload(a, "http://localhost:8080/path", http.StatusOK)
	testDownload(a, "http://localhost:8080/index", http.StatusOK)
	testDownloadNotFound(a, "http://localhost:8080/not-exists")
}

func TestContext_ServeFile_windows(t *testing.T) {
	if runtime.GOOS != "windows" {
		return
	}

	a := assert.New(t, false)
	exit := make(chan bool, 1)

	s := newServer(a, nil)
	defer func() {
		a.NotError(s.Close(0))
		<-exit
	}()
	router := s.NewRouter("default", "http://localhost:8080/root", group.MatcherFunc(group.Any))
	a.NotNil(router)

	a.NotPanic(func() {
		router.Get("/path", func(ctx *Context) Responser {
			return ctx.ServeFile(".\\testdata\\file1.txt", "index.html", map[string]string{"Test": "Test"})
		})

		router.Get("/index", func(ctx *Context) Responser {
			return ctx.ServeFile(".\\testdata", "file1.txt", map[string]string{"Test": "Test"})
		})

		router.Get("/not-exists", func(ctx *Context) Responser {
			// file1.text 不存在
			return ctx.ServeFile("c:\\not-exists-dir\\file1.text", "index.html", map[string]string{"Test": "Test"})
		})
	})

	go func() {
		a.Equal(s.Serve(), http.ErrServerClosed)
		exit <- true
	}()
	time.Sleep(500 * time.Millisecond)

	testDownload(a, "http://localhost:8080/path", http.StatusOK)
	testDownload(a, "http://localhost:8080/index", http.StatusOK)
	testDownloadNotFound(a, "http://localhost:8080/not-exists")
}

func testDownload(a *assert.Assertion, path string, status int) {
	a.TB().Helper()
	rest.NewRequest(a, nil, http.MethodGet, path).Do().
		Status(status).
		BodyNotEmpty().
		Header("Test", "Test")
}

func testDownloadNotFound(a *assert.Assertion, path string) {
	rest.NewRequest(a, nil, http.MethodGet, path).Do().
		Status(http.StatusNotFound)
}

func TestContext_ServeFS(t *testing.T) {
	a := assert.New(t, false)
	fsys := os.DirFS("./")

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/path", nil)
	ctx := &Context{
		Response: w,
		Request:  r,
	}

	// p = context.go
	w.Body.Reset()
	data, err := fs.ReadFile(fsys, "context.go")
	a.NotError(err).NotNil(data)
	a.NotError(ctx.ServeFS(fsys, "context.go", "", nil))
	a.Equal(w.Result().StatusCode, http.StatusOK).
		Equal(w.Body.Bytes(), data)

	// index = context.go
	w.Body.Reset()
	data, err = fs.ReadFile(fsys, "context.go")
	a.NotError(err).NotNil(data)
	a.NotError(ctx.ServeFS(fsys, "", "context.go", nil))
	a.Equal(w.Result().StatusCode, http.StatusOK).
		Equal(w.Body.Bytes(), data)

	// p=gob, index=gob.go
	w.Body.Reset()
	data, err = fs.ReadFile(fsys, "testdata/key.pem")
	a.NotError(err).NotNil(data)
	a.NotError(ctx.ServeFS(fsys, "testdata", "key.pem", nil))
	a.Equal(w.Result().StatusCode, http.StatusOK).
		Equal(w.Body.Bytes(), data)

	// p=gob, index=gob.go, headers != nil
	w.Body.Reset()
	data, err = fs.ReadFile(fsys, "testdata/key.pem")
	a.NotError(err).NotNil(data)
	a.NotError(ctx.ServeFS(fsys, "testdata", "key.pem", map[string]string{"Test": "ttt"}))
	a.Equal(w.Result().StatusCode, http.StatusOK).
		Equal(w.Body.Bytes(), data).
		Equal(w.Header().Get("Test"), "ttt")

	w.Body.Reset()
	a.ErrorIs(ctx.ServeFS(fsys, "testdata", "", nil), os.ErrNotExist).
		Empty(w.Body.Bytes())

	w.Body.Reset()
	a.ErrorIs(ctx.ServeFS(fsys, "testdata", "not-exists.go", nil), os.ErrNotExist).
		Empty(w.Body.Bytes())

	w.Body.Reset()
	a.ErrorIs(ctx.ServeFS(fsys, "not-exists", "file1.txt", nil), os.ErrNotExist).
		Empty(w.Body.Bytes())
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
	r := httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Set("accept-language", "cmn-hant")
	r.Header.Set("accept", text.Mimetype)
	ctx := srv.NewContext(w, r)
	a.NotNil(ctx)
	a.NotError(ctx.Marshal(http.StatusOK, ctx.LocalePrinter.Sprintf("test"), nil))
	a.Equal(w.Body.String(), "測試")

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Set("accept-language", "cmn-hans")
	r.Header.Set("accept", text.Mimetype)
	ctx = srv.NewContext(w, r)
	a.NotNil(ctx)
	n, err := ctx.LocalePrinter.Fprintf(ctx.Response, "test")
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
