// SPDX-License-Identifier: MIT

package web

import (
	"bytes"
	"compress/flate"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/mux/v7/types"
	"golang.org/x/text/language"

	"github.com/issue9/web/internal/header"
)

type response struct {
	http.ResponseWriter
	w io.Writer
}

func (r *response) Write(data []byte) (int, error) { return r.w.Write(data) }

func TestContext_Render(t *testing.T) {
	a := assert.New(t, false)
	srv := newTestServer(a)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/p1", nil)
	r.Header.Set(header.ContentType, "application/json")
	r.Header.Set(header.Accept, "application/json")
	ctx := srv.NewContext(w, r, types.NewContext())
	a.NotNil(ctx)
	ctx.Render(http.StatusCreated, objectInst)
	a.Equal(w.Result().StatusCode, http.StatusCreated).
		Equal(w.Header().Get(header.ContentType), header.BuildContentType("application/json", "utf-8")).
		Equal(w.Header().Get(header.ContentLang), "zh-Hans")

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/p1", nil)
	r.Header.Set(header.Accept, "application/json")
	r.Header.Set(header.AcceptLang, "")
	ctx = srv.NewContext(w, r, types.NewContext())
	a.NotNil(ctx)
	ctx.Render(http.StatusCreated, objectInst)
	a.Equal(w.Result().StatusCode, http.StatusCreated).
		Equal(w.Header().Get(header.ContentLang), language.SimplifiedChinese.String()).
		Equal(w.Body.String(), objectJSONString)

	// 输出 nil，content-type 和 content-language 均为空
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/p1", nil)
	r.Header.Set(header.Accept, "application/json")
	r.Header.Set(header.AcceptLang, "zh-hans")
	ctx = srv.NewContext(w, r, types.NewContext())
	a.NotNil(ctx)
	ctx.Render(http.StatusCreated, nil)
	a.Equal(w.Result().StatusCode, http.StatusCreated).
		Equal(w.Header().Get(header.ContentLang), ""). // 指定了输出语言，也返回空。
		Equal(w.Header().Get(header.ContentType), "")

	// accept,accept-language,accept-charset
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/p1", nil)
	r.Header.Set(header.Accept, "application/json")
	r.Header.Set(header.AcceptLang, "zh-hans")
	r.Header.Set(header.AcceptCharset, "gbk")
	ctx = srv.NewContext(w, r, types.NewContext())
	a.NotNil(ctx)
	ctx.Render(http.StatusCreated, objectInst)
	a.Equal(w.Body.Bytes(), objectGBKBytes)

	// 同时指定了 accept,accept-language,accept-charset 和 accept-encoding
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/p1", nil)
	r.Header.Set(header.Accept, "application/json")
	r.Header.Set(header.AcceptLang, "zh-hans")
	r.Header.Set(header.AcceptCharset, "gbk")
	r.Header.Set(header.AcceptEncoding, "deflate")
	ctx = srv.NewContext(w, r, types.NewContext())
	a.NotNil(ctx)
	ctx.Render(http.StatusCreated, objectInst)
	srv.b.FreeContext(ctx)
	a.Equal(w.Result().StatusCode, http.StatusCreated).
		Equal(w.Header().Get(header.ContentEncoding), "deflate")
	data, err := io.ReadAll(flate.NewReader(w.Body))
	a.NotError(err).Equal(data, objectGBKBytes)

	// 同时通过 ctx.Write 和 ctx.Marshal 输出内容
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/p1", nil)
	r.Header.Set(header.Accept, "application/json")
	ctx = srv.NewContext(w, r, types.NewContext())
	a.NotNil(ctx)
	n, err := ctx.Write([]byte("123"))
	a.NotError(err).True(n > 0)
	a.PanicString(func() {
		ctx.Render(http.StatusCreated, "456")
	}, "已有状态码 200，再次设置无效 201")
	srv.b.FreeContext(ctx)
	a.Equal(w.Result().StatusCode, http.StatusOK).Equal(w.Body.String(), `123`)

	// ctx.Write 在 ctx.Marshal 之后可以正常调用。
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/p1", nil)
	r.Header.Set(header.Accept, "application/json")
	ctx = srv.NewContext(w, r, types.NewContext())
	ctx.Render(http.StatusCreated, "123")
	n, err = ctx.Write([]byte("123"))
	a.NotError(err)
	srv.b.FreeContext(ctx)
	a.True(n > 0).
		Equal(w.Body.String(), `"123"123`)

	// outputMimetype.Marshal 返回 ErrUnsupported
	srv.logBuf.Reset()
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/p1", nil)
	r.Header.Set(header.Accept, "application/test")
	ctx = srv.NewContext(w, r, types.NewContext())
	a.NotNil(ctx, srv.logBuf.String()).
		Equal(ctx.Mimetype(false), "application/test").
		Equal(ctx.Charset(), header.UTF8Name)
	ctx.Render(http.StatusCreated, "任意值")
	srv.b.FreeContext(ctx)
	a.Equal(w.Result().StatusCode, http.StatusNotAcceptable)

	// outputMimetype.Marshal 返回错误
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/p1", nil)
	r.Header.Set(header.Accept, "application/test")
	ctx = srv.NewContext(w, r, types.NewContext())
	a.NotNil(ctx).
		Equal(ctx.Mimetype(false), "application/test").
		Equal(ctx.Charset(), header.UTF8Name)
	ctx.Render(http.StatusCreated, errors.New("error"))
	srv.b.FreeContext(ctx)
	a.Equal(w.Result().StatusCode, http.StatusNotAcceptable)
}

func TestContext_Wrap(t *testing.T) {
	a := assert.New(t, false)
	s := newTestServer(a)

	// 输出内容之后，不能调用 Wrap

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/p1", nil)
	r.Header.Set(header.Accept, "application/json")
	r.Header.Set(header.AcceptLang, "cmn-hant")
	ctx := s.NewContext(w, r, types.NewContext())
	ctx.Write([]byte("abc"))

	a.PanicString(func() {
		buf := &bytes.Buffer{}
		ctx.Wrap(func(w http.ResponseWriter) http.ResponseWriter { return &response{w: buf, ResponseWriter: w} })
	}, "已有内容输出，不可再更改！")

	a.Equal(w.Body.String(), "abc")

	// 调用 Wrap 之后修改了报头内容

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/p1", nil)
	r.Header.Set(header.Accept, "application/json")
	r.Header.Set(header.AcceptLang, "cmn-hant")
	r.Header.Set(header.AcceptEncoding, "")
	ctx = s.NewContext(w, r, types.NewContext())

	ctx.Header().Set("h1", "v1")
	buf := &bytes.Buffer{}
	ctx.Wrap(func(w http.ResponseWriter) http.ResponseWriter { return &response{w: buf, ResponseWriter: w} })
	ctx.Header().Set("h2", "v2")
	ctx.Write([]byte("abc"))
	a.Equal(buf.String(), "abc").
		Equal(w.Header().Get("h1"), "v1").
		Equal(w.Header().Get("h2"), "v2").
		Empty(w.Body.String())

	// 多次调用 Wrap

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/p1", nil)
	r.Header.Set(header.Accept, "application/json")
	r.Header.Set(header.AcceptLang, "cmn-hant")
	r.Header.Set(header.AcceptEncoding, "")
	ctx = s.NewContext(w, r, types.NewContext())

	a.PanicString(func() { // Wrap(nil)
		ctx.Wrap(nil)
	}, "参数 f 不能为空")

	buf1 := &bytes.Buffer{}
	buf2 := &bytes.Buffer{}
	ctx.Wrap(func(w http.ResponseWriter) http.ResponseWriter { return &response{w: buf1, ResponseWriter: w} })
	ctx.Wrap(func(w http.ResponseWriter) http.ResponseWriter { return &response{w: buf2, ResponseWriter: w} })
	ctx.Write([]byte("abc"))
	a.Equal(buf2.String(), "abc").Empty(buf1.String()).
		True(w.Result().StatusCode > 199).
		Empty(w.Body.String())
}

func TestContext_LocalePrinter(t *testing.T) {
	a := assert.New(t, false)
	srv := newTestServer(a)

	b := srv.Locale()
	a.NotError(b.SetString(language.MustParse("cmn-hans"), "test", "测试"))
	a.NotError(b.SetString(language.MustParse("cmn-hant"), "test", "測試"))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/p", nil)
	r.Header.Set(header.Accept, "application/json")
	r.Header.Set(header.AcceptLang, "cmn-hant")
	ctx := srv.NewContext(w, r, types.NewContext())
	ctx.Render(http.StatusOK, ctx.Sprintf("test"))
	a.Equal(w.Body.String(), `"測試"`)

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/p", nil)
	r.Header.Set(header.Accept, "application/json")
	r.Header.Set(header.AcceptLang, "cmn-hans")
	ctx = srv.NewContext(w, r, types.NewContext())
	n, err := ctx.LocalePrinter().Fprintf(ctx, "test")
	a.NotError(err).Equal(n, len("测试"))
	a.Equal(w.Body.String(), "测试")
}

func TestNotModified(t *testing.T) {
	a := assert.New(t, false)
	s := newTestServer(a)

	// weak

	const body = "string"
	nm := NotModified(
		func() (string, bool) { return body, true },
		func() (any, error) { return body, nil },
	)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/p", nil)
	s.NewContext(w, r, types.NewContext()).apply(nm)
	tag := w.Header().Get(header.ETag)
	a.Equal(w.Result().StatusCode, http.StatusOK).
		NotEmpty(tag)

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/p", nil)
	r.Header.Set(header.IfNoneMatch, tag)
	s.NewContext(w, r, types.NewContext()).apply(nm)
	a.Equal(w.Result().StatusCode, http.StatusNotModified)

	// Post 不启用
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodPost, "/p", nil)
	s.NewContext(w, r, types.NewContext()).apply(nm)
	tag = w.Header().Get(header.ETag)
	a.Equal(w.Result().StatusCode, http.StatusOK).Empty(tag)

	// weak=false

	nm = NotModified(
		func() (string, bool) { return body, false },
		func() (any, error) { return []byte(body), nil },
	)

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/p", nil)
	s.NewContext(w, r, types.NewContext()).apply(nm)
	tag = w.Header().Get(header.ETag)
	a.Equal(w.Result().StatusCode, http.StatusOK).
		NotEmpty(tag)

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/p", nil)
	r.Header.Set(header.IfNoneMatch, tag)
	s.NewContext(w, r, types.NewContext()).apply(nm)
	a.Equal(w.Result().StatusCode, http.StatusNotModified)

	// error

	nm = NotModified(
		func() (string, bool) { return body, false },
		func() (any, error) { return nil, errors.New("500") },
	)

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/p", nil)
	s.NewContext(w, r, types.NewContext()).apply(nm)
	a.Equal(w.Result().StatusCode, http.StatusInternalServerError)
}

func TestCreated(t *testing.T) {
	a := assert.New(t, false)
	s := newTestServer(a)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/p", nil)
	r.Header.Set(header.Accept, "application/json")
	s.NewContext(w, r, types.NewContext()).
		apply(Created(nil, ""))
	a.Equal(w.Result().StatusCode, http.StatusCreated).
		Empty(w.Body.String())

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/p", nil)
	r.Header.Set(header.Accept, "application/json")
	s.NewContext(w, r, types.NewContext()).
		apply(Created(objectInst, ""))
	a.Equal(w.Result().StatusCode, http.StatusCreated).
		Equal(w.Body.String(), objectJSONString).
		Empty(w.Header().Get(header.Location))

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/p", nil)
	r.Header.Set(header.Accept, "application/json")
	s.NewContext(w, r, types.NewContext()).
		apply(Created(objectInst, "/p2"))
	a.Equal(w.Result().StatusCode, http.StatusCreated).
		Equal(w.Body.String(), objectJSONString).
		Equal(w.Header().Get(header.Location), "/p2")
}

func TestRedirect(t *testing.T) {
	a := assert.New(t, false)
	s := newTestServer(a)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/p", nil)
	ctx := s.NewContext(w, r, types.NewContext())
	ctx.apply(ctx.NotImplemented())
	a.Equal(w.Result().StatusCode, http.StatusNotImplemented)

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/p", nil)
	Redirect(http.StatusMovedPermanently, "http://example.com")
	s.NewContext(w, r, types.NewContext()).
		apply(Redirect(http.StatusMovedPermanently, "http://example.com"))
	a.Equal(w.Result().StatusCode, http.StatusMovedPermanently).
		Equal(w.Header().Get(header.Location), "http://example.com")
}

func TestNoContent(t *testing.T) {
	// 检测 204 是否存在 http: request method or response status code does not allow body

	a := assert.New(t, false)
	s := newTestServer(a)
	s.logBuf.Reset()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/p", nil)
	r.Header.Set(header.AcceptEncoding, "gzip") // 服务端不应该构建压缩对象
	r.Header.Set(header.Accept, "application/json")
	s.NewContext(w, r, types.NewContext()).apply(NoContent())
	a.NotContains(s.logBuf.String(), "request method or response status code does not allow body")
}
