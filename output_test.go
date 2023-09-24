// SPDX-License-Identifier: MIT

package web

import (
	"bytes"
	"compress/flate"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/issue9/assert/v3"
	"golang.org/x/text/language"

	"github.com/issue9/web/internal/header"
	"github.com/issue9/web/internal/testdata"
	"github.com/issue9/web/logs"
	"github.com/issue9/web/servertest"
)

type response struct {
	http.ResponseWriter
	w io.Writer
}

func (r *response) Write(data []byte) (int, error) { return r.w.Write(data) }

func TestContext_Render(t *testing.T) {
	a := assert.New(t, false)
	buf := new(bytes.Buffer)
	srv := newTestServer(a, &Options{
		Language:   language.SimplifiedChinese,
		HTTPServer: &http.Server{Addr: ":8080"},
		Logs:       &logs.Options{Handler: logs.NewTextHandler(logs.MicroLayout, buf), Levels: logs.AllLevels()},
	})
	defer servertest.Run(a, srv)()
	defer srv.Close(0)

	r := srv.NewRouter("def", nil)

	// 自定义报头
	buf.Reset()
	r.Post("/p1", func(ctx *Context) Responser {
		ctx.Render(http.StatusCreated, testdata.ObjectInst)
		return nil
	})
	servertest.Post(a, "http://localhost:8080/p1", nil).
		Header("Content-Type", "application/json").
		Header("Accept", "application/json").
		Do(nil).
		Status(http.StatusCreated).
		StringBody(testdata.ObjectJSONString).
		Header("content-type", header.BuildContentType("application/json", "utf-8")).
		Header("content-language", "zh-Hans")
	a.Zero(buf.Len())

	buf.Reset()
	r.Get("/p2", func(ctx *Context) Responser {
		ctx.Render(http.StatusCreated, testdata.ObjectInst)
		return nil
	})
	servertest.Get(a, "http://localhost:8080/p2").
		Header("accept", "application/json").
		Header("accept-language", "").
		Do(nil).
		Status(http.StatusCreated).
		StringBody(testdata.ObjectJSONString).
		Header("content-language", language.SimplifiedChinese.String()) // 未指定，采用默认值
	a.Zero(buf.Len())

	// 输出 nil，content-type 和 content-language 均为空
	buf.Reset()
	r.Get("/p3", func(ctx *Context) Responser {
		ctx.Render(http.StatusCreated, nil)
		return nil
	})
	servertest.Get(a, "http://localhost:8080/p3").
		Header("Accept", "application/json").
		Header("Accept-language", "zh-hans").
		Do(nil).
		Status(http.StatusCreated).
		StringBody("").
		Header("content-language", ""). // 指定了输出语言，也返回空。
		Header("content-Type", "")
	a.Zero(buf.Len())

	// accept,accept-language,accept-charset
	buf.Reset()
	r.Get("/p4", func(ctx *Context) Responser {
		ctx.Render(http.StatusCreated, testdata.ObjectInst)
		return nil
	})
	servertest.Get(a, "http://localhost:8080/p4").
		Header("Accept", "application/json").
		Header("Accept-Language", "zh-Hans").
		Header("Accept-Charset", "gbk").
		Do(nil).
		Status(http.StatusCreated).
		BodyFunc(func(a *assert.Assertion, body []byte) {
			a.Equal(body, testdata.ObjectGBKBytes)
		})
	a.Zero(buf.Len())

	// 同时指定了 accept,accept-language,accept-charset 和 accept-encoding
	buf.Reset()
	r.Get("/p7", func(ctx *Context) Responser {
		ctx.Render(http.StatusCreated, testdata.ObjectInst)
		return nil
	})
	servertest.Get(a, "http://localhost:8080/p7").
		Header("Accept", "application/json").
		Header("Accept-Language", "zh-Hans").
		Header("Accept-Charset", "gbk").
		Header("Accept-Encoding", "gzip;q=0.9,deflate").
		Do(nil).
		Status(http.StatusCreated).
		Header("content-encoding", "deflate").
		BodyFunc(func(a *assert.Assertion, body []byte) {
			data, err := io.ReadAll(flate.NewReader(bytes.NewBuffer(body)))
			a.NotError(err).Equal(data, testdata.ObjectGBKBytes)
		})
	a.Zero(buf.Len())

	// 同时通过 ctx.Write 和 ctx.Marshal 输出内容
	buf.Reset()
	r.Get("/p8", func(ctx *Context) Responser {
		_, err := ctx.Write([]byte("123"))
		a.NotError(err)
		a.PanicString(func() {
			ctx.Render(http.StatusCreated, "456")
		}, "已有状态码 200，再次设置无效 201")
		return nil
	})
	servertest.Get(a, "http://localhost:8080/p8").Header("Accept", "application/json").Do(nil)

	// ctx.Write 在 ctx.Marshal 之后可以正常调用。
	buf.Reset()
	r.Get("/p9", func(ctx *Context) Responser {
		ctx.Render(http.StatusCreated, "123")
		_, err := ctx.Write([]byte("456"))
		a.NotError(err)
		return nil
	})
	servertest.Get(a, "http://localhost:8080/p9").
		Header("Accept", "application/json").
		Header("Accept-Encoding", "gzip;q=0.9,deflate").
		Do(nil).
		Status(http.StatusCreated). // 压缩对象缓存了 WriteHeader 的发送
		BodyFunc(func(a *assert.Assertion, body []byte) {
			data, err := io.ReadAll(flate.NewReader(bytes.NewBuffer(body)))
			a.NotError(err).Equal(string(data), `"123"456`)
		})
	a.Zero(buf.Len())

	// outputMimetype == nil
	buf.Reset()
	r.Get("/p10", func(ctx *Context) Responser {
		a.Nil(ctx.outputMimetype.Marshal).
			Equal(ctx.Mimetype(false), "nil").
			Equal(ctx.Charset(), header.UTF8Name)
		ctx.Render(http.StatusCreated, "val")
		return nil
	})
	servertest.Get(a, "http://localhost:8080/p10").Header("Accept", "nil").
		Do(nil).Status(http.StatusNotAcceptable)

	// outputMimetype 返回 ErrUnsupported
	buf.Reset()
	r.Get("/p11", func(ctx *Context) Responser {
		ctx.Render(http.StatusCreated, "任意值")
		return nil
	})
	servertest.Get(a, "http://localhost:8080/p11").Header("Accept", "application/test").
		Do(nil).Status(http.StatusNotAcceptable)
	a.NotZero(buf.Len())

	// outputMimetype 返回错误
	buf.Reset()
	r.Get("/p12", func(ctx *Context) Responser {
		ctx.Render(http.StatusCreated, errors.New("error"))
		return nil
	})
	servertest.Get(a, "http://localhost:8080/p12").Header("Accept", "application/test").
		Do(nil).Status(http.StatusNotAcceptable)
	a.NotZero(buf.Len())

	// 103
	buf.Reset()
	r.Get("/p13", func(ctx *Context) Responser {
		ctx.WriteHeader(http.StatusEarlyHints)
		_, err := ctx.Write([]byte(`123`))
		if err != nil {
			return ctx.Error(err, ProblemInternalServerError)
		}
		return nil
	})
	servertest.Get(a, "http://localhost:8080/p13").
		Header("Accept", "application/json").
		Do(nil).
		Status(http.StatusOK).
		BodyFunc(func(a *assert.Assertion, body []byte) {
			a.Equal(body, []byte(`123`))
		})
	a.Zero(buf.Len())
}

func TestContext_SetWriter(t *testing.T) {
	a := assert.New(t, false)
	srv := newTestServer(a, &Options{
		Language:   language.SimplifiedChinese,
		HTTPServer: &http.Server{Addr: ":8080"},
	})
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
	srv := newTestServer(a, &Options{
		Language:   language.SimplifiedChinese,
		HTTPServer: &http.Server{Addr: ":8080"},
	})
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
	s := newTestServer(a, nil)

	r := s.NewRouter("def", nil)
	r.Any("/string", func(ctx *Context) Responser {
		const body = "string"
		return NotModified(
			func() (string, bool) { return body, true },
			func() (any, error) { return body, nil },
		)
	})
	r.Get("/bytes", func(ctx *Context) Responser {
		const body = "bytes"
		return NotModified(
			func() (string, bool) { return body, false },
			func() (any, error) { return []byte(body), nil },
		)
	})
	r.Get("/errors-500", func(ctx *Context) Responser {
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
	s := newTestServer(a, nil)
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
	s := newTestServer(a, nil)
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
