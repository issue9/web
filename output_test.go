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
	"golang.org/x/text/language"

	"github.com/issue9/web/internal/header"
	"github.com/issue9/web/internal/testdata"
	"github.com/issue9/web/servertest"
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
	ctx := NewContext(srv, w, r, nil, header.RequestIDKey)
	a.NotNil(ctx)
	ctx.Render(http.StatusCreated, testdata.ObjectInst)
	a.Equal(w.Result().StatusCode, http.StatusCreated).
		Equal(w.Header().Get(header.ContentType), header.BuildContentType("application/json", "utf-8")).
		Equal(w.Header().Get(header.ContentLang), "zh-Hans")

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/p1", nil)
	r.Header.Set(header.Accept, "application/json")
	r.Header.Set(header.AcceptLang, "")
	ctx = NewContext(srv, w, r, nil, header.RequestIDKey)
	a.NotNil(ctx)
	ctx.Render(http.StatusCreated, testdata.ObjectInst)
	a.Equal(w.Result().StatusCode, http.StatusCreated).
		Equal(w.Header().Get(header.ContentLang), language.SimplifiedChinese.String()).
		Equal(w.Body.String(), testdata.ObjectJSONString)

	// 输出 nil，content-type 和 content-language 均为空
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/p1", nil)
	r.Header.Set(header.Accept, "application/json")
	r.Header.Set(header.AcceptLang, "zh-hans")
	ctx = NewContext(srv, w, r, nil, header.RequestIDKey)
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
	ctx = NewContext(srv, w, r, nil, header.RequestIDKey)
	a.NotNil(ctx)
	ctx.Render(http.StatusCreated, testdata.ObjectInst)
	a.Equal(w.Body.Bytes(), testdata.ObjectGBKBytes)

	// 同时指定了 accept,accept-language,accept-charset 和 accept-encoding
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/p1", nil)
	r.Header.Set(header.Accept, "application/json")
	r.Header.Set(header.AcceptLang, "zh-hans")
	r.Header.Set(header.AcceptCharset, "gbk")
	r.Header.Set(header.AcceptEncoding, "deflate")
	ctx = NewContext(srv, w, r, nil, header.RequestIDKey)
	a.NotNil(ctx)
	ctx.Render(http.StatusCreated, testdata.ObjectInst)
	ctx.Free()
	a.Equal(w.Result().StatusCode, http.StatusCreated).
		Equal(w.Header().Get(header.ContentEncoding), "deflate")
	data, err := io.ReadAll(flate.NewReader(w.Body))
	a.NotError(err).Equal(data, testdata.ObjectGBKBytes)

	// 同时通过 ctx.Write 和 ctx.Marshal 输出内容
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/p1", nil)
	r.Header.Set(header.Accept, "application/json")
	ctx = NewContext(srv, w, r, nil, header.RequestIDKey)
	a.NotNil(ctx)
	n, err := ctx.Write([]byte("123"))
	a.NotError(err).True(n > 0)
	a.PanicString(func() {
		ctx.Render(http.StatusCreated, "456")
	}, "已有状态码 200，再次设置无效 201")
	ctx.Free()
	a.Equal(w.Result().StatusCode, http.StatusOK)

	// ctx.Write 在 ctx.Marshal 之后可以正常调用。
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/p1", nil)
	r.Header.Set(header.Accept, "application/json")
	ctx = NewContext(srv, w, r, nil, header.RequestIDKey)
	ctx.Render(http.StatusCreated, "123")
	n, err = ctx.Write([]byte("123"))
	a.NotError(err)
	ctx.Free()
	a.True(n > 0).
		Equal(w.Body.String(), `"123"123`)

	// outputMimetype.MarshalBuilder() == nil
	srv.logBuf.Reset()
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/p1", nil)
	r.Header.Set(header.Accept, "nil")
	ctx = NewContext(srv, w, r, nil, header.RequestIDKey)
	a.NotNil(ctx, srv.logBuf.String()).
		Equal(ctx.Mimetype(false), "nil").
		Equal(ctx.Charset(), header.UTF8Name)
	ctx.Render(http.StatusCreated, "val")
	ctx.Free()
	a.Equal(w.Result().StatusCode, http.StatusNotAcceptable)

	// outputMimetype.MarshalBuiler()() 返回 ErrUnsupported
	srv.logBuf.Reset()
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/p1", nil)
	r.Header.Set(header.Accept, "application/test")
	ctx = NewContext(srv, w, r, nil, header.RequestIDKey)
	a.NotNil(ctx, srv.logBuf.String()).
		Equal(ctx.Mimetype(false), "application/test").
		Equal(ctx.Charset(), header.UTF8Name)
	ctx.Render(http.StatusCreated, "任意值")
	ctx.Free()
	a.Equal(w.Result().StatusCode, http.StatusNotAcceptable)

	// outputMimetype 返回错误
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/p1", nil)
	r.Header.Set(header.Accept, "application/test")
	ctx = NewContext(srv, w, r, nil, header.RequestIDKey)
	a.NotNil(ctx).
		Equal(ctx.Mimetype(false), "application/test").
		Equal(ctx.Charset(), header.UTF8Name)
	ctx.Render(http.StatusCreated, errors.New("error"))
	ctx.Free()
	a.Equal(w.Result().StatusCode, http.StatusNotAcceptable)

	// 103
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/p1", nil)
	r.Header.Set(header.Accept, "application/json")
	ctx = NewContext(srv, w, r, nil, header.RequestIDKey)
	a.NotNil(ctx)
	ctx.WriteHeader(http.StatusEarlyHints) // 之后可再输出
	_, err = ctx.Write([]byte(`123`))
	a.NotError(err)
	ctx.Free()
	a.Equal(w.Body.String(), "123")
}

func TestContext_SetWriter(t *testing.T) {
	a := assert.New(t, false)
	srv := newTestServer(a)
	r := srv.NewRouter("def", nil)

	defer servertest.Run(a, srv)()
	defer srv.Close(0)

	r.Get("/p1", func(ctx *Context) Responser {
		ctx.Write([]byte("abc"))

		a.PanicString(func() {
			buf := &bytes.Buffer{}
			ctx.SetWriter(func(w http.ResponseWriter) http.ResponseWriter { return &response{w: buf, ResponseWriter: w} })
		}, "已有内容输出，不可再更改！")
		return nil
	})
	servertest.Get(a, "http://localhost:8080/p1").
		Header("accept-language", "cmn-hant").
		Header("accept", "application/json").
		Do(nil).
		StringBody("abc")

	// setWriter
	r.Get("/p2", func(ctx *Context) Responser {
		ctx.Header().Set("h1", "v1")
		buf := &bytes.Buffer{}
		ctx.SetWriter(func(w http.ResponseWriter) http.ResponseWriter { return &response{w: buf, ResponseWriter: w} })
		ctx.Header().Set("h2", "v2")
		ctx.Write([]byte("abc"))
		a.Equal(buf.String(), "abc")
		return nil
	})
	servertest.Get(a, "http://localhost:8080/p2").
		Header("accept-language", "cmn-hant").
		Header("accept", "application/json").
		Header("accept-encoding", "").
		Do(nil).
		Status(http.StatusOK).
		BodyEmpty().
		Header("h1", "v1").
		Header("h2", "v2")

	// 多次调用 setWriter
	r.Get("/p3", func(ctx *Context) Responser {
		a.PanicString(func() { // setWriter(nil)
			ctx.SetWriter(nil)
		}, "参数 w 不能为空")

		buf1 := &bytes.Buffer{}
		buf2 := &bytes.Buffer{}
		ctx.SetWriter(func(w http.ResponseWriter) http.ResponseWriter { return &response{w: buf1, ResponseWriter: w} })
		ctx.SetWriter(func(w http.ResponseWriter) http.ResponseWriter { return &response{w: buf2, ResponseWriter: w} })
		ctx.Write([]byte("abc"))
		a.Equal(buf2.String(), "abc").Empty(buf1.String())

		return nil
	})
	servertest.Get(a, "http://localhost:8080/p3").
		Header("accept-language", "cmn-hant").
		Header("accept", "application/json").
		Header("accept-encoding", "").
		Do(nil).
		BodyEmpty().
		Success()
}

func TestContext_LocalePrinter(t *testing.T) {
	a := assert.New(t, false)
	srv := newTestServer(a)
	r := srv.NewRouter("def", nil)

	b := srv.Catalog()
	a.NotError(b.SetString(language.MustParse("cmn-hans"), "test", "测试"))
	a.NotError(b.SetString(language.MustParse("cmn-hant"), "test", "測試"))

	defer servertest.Run(a, srv)()
	defer srv.Close(0)

	r.Get("/p1", func(ctx *Context) Responser {
		ctx.Render(http.StatusOK, ctx.Sprintf("test"))
		return nil
	})
	servertest.Get(a, "http://localhost:8080/p1").
		Header("accept-language", "cmn-hant").
		Header("accept", "application/json").
		Do(nil).
		StringBody(`"測試"`)

	r.Get("/p2", func(ctx *Context) Responser {
		n, err := ctx.LocalePrinter().Fprintf(ctx, "test")
		a.NotError(err).Equal(n, len("测试"))
		return nil
	})
	servertest.Get(a, "http://localhost:8080/p2").
		Header("accept-language", "cmn-hans").
		Header("accept", "application/json").
		Do(nil).
		StringBody("测试")
}

func TestNotModified(t *testing.T) {
	a := assert.New(t, false)
	s := newTestServer(a)

	r := s.NewRouter("def", nil)
	r.Any("/string", func(*Context) Responser {
		const body = "string"
		return NotModified(
			func() (string, bool) { return body, true },
			func() (any, error) { return body, nil },
		)
	})
	r.Get("/bytes", func(*Context) Responser {
		const body = "bytes"
		return NotModified(
			func() (string, bool) { return body, false },
			func() (any, error) { return []byte(body), nil },
		)
	})
	r.Get("/errors-500", func(*Context) Responser {
		const body = "500"
		return NotModified(
			func() (string, bool) { return body, false },
			func() (any, error) { return nil, errors.New("500") },
		)
	})

	defer servertest.Run(a, s)()
	defer s.Close(0)

	// get /string

	resp := servertest.Get(a, "http://localhost:8080/string").
		Do(nil).
		Status(http.StatusOK).
		Resp()
	tag := resp.Header.Get(header.ETag)
	servertest.Get(a, "http://localhost:8080/string").
		Header(header.IfNoneMatch, tag).
		Do(nil).
		Status(http.StatusNotModified)

	// post /string

	resp = servertest.Post(a, "http://localhost:8080/string", nil).
		Do(nil).
		Status(http.StatusOK).
		Resp()
	tag = resp.Header.Get(header.ETag)
	servertest.Post(a, "http://localhost:8080/string", nil).
		Header(header.IfNoneMatch, tag).
		Do(nil).
		Status(http.StatusOK)

	// get /bytes

	resp = servertest.Get(a, "http://localhost:8080/bytes").
		Do(nil).
		Status(http.StatusOK).
		Resp()
	tag = resp.Header.Get(header.ETag)
	servertest.Get(a, "http://localhost:8080/bytes").
		Header(header.IfNoneMatch, tag).
		Do(nil).
		Status(http.StatusNotModified)

		// get /errors-500

	servertest.Get(a, "http://localhost:8080/errors-500").
		Do(nil).
		Status(http.StatusInternalServerError).
		Resp()
	servertest.Get(a, "http://localhost:8080/errors-500").
		Do(nil).
		Status(http.StatusInternalServerError)
}

func TestCreated(t *testing.T) {
	a := assert.New(t, false)
	s := newTestServer(a)
	r := s.NewRouter("def", nil)

	// Location == ""
	r.Get("/created", func(ctx *Context) Responser {
		return Created(testdata.ObjectInst, "")
	})
	// Location == "/test"
	r.Get("/created/location", func(ctx *Context) Responser {
		return Created(testdata.ObjectInst, "/test")
	})

	defer servertest.Run(a, s)()
	defer s.Close(0)

	servertest.Get(a, "http://localhost:8080/created").Header("accept", "application/json").Do(nil).
		Status(http.StatusCreated).
		StringBody(testdata.ObjectJSONString)

	servertest.Get(a, "http://localhost:8080/created/location").Header("accept", "application/json").Do(nil).
		Status(http.StatusCreated).
		StringBody(testdata.ObjectJSONString).
		Header("Location", "/test")
}

func TestRedirect(t *testing.T) {
	a := assert.New(t, false)
	s := newTestServer(a)
	r := s.NewRouter("def", nil)

	r.Get("/not-implement", func(ctx *Context) Responser {
		return ctx.NotImplemented()
	})
	r.Get("/ok", func(ctx *Context) Responser {
		return Created(nil, "")
	})
	r.Get("/redirect", func(ctx *Context) Responser {
		return Redirect(http.StatusMovedPermanently, "http://localhost:8080/ok")
	})

	defer servertest.Run(a, s)()
	defer s.Close(0)

	servertest.Get(a, "http://localhost:8080/not-implement").Do(nil).Status(http.StatusNotImplemented)

	servertest.Get(a, "http://localhost:8080/redirect").Do(nil).
		Status(http.StatusCreated). // http.Client.Do 会自动重定向并请求
		Header("Location", "")
}
