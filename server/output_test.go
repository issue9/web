// SPDX-License-Identifier: MIT

package server

import (
	"bytes"
	"compress/flate"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/assert/v3/rest"
	"golang.org/x/text/language"

	"github.com/issue9/web/internal/header"
	"github.com/issue9/web/internal/testdata"
	"github.com/issue9/web/logs"
)

type response struct {
	http.ResponseWriter
	w io.Writer
}

func (r *response) Write(data []byte) (int, error) { return r.w.Write(data) }

func TestContext_Marshal(t *testing.T) {
	a := assert.New(t, false)
	buf := new(bytes.Buffer)
	o := &logs.Options{Writer: logs.NewTextWriter(logs.MicroLayout, buf), Levels: logs.AllLevels()}
	srv := newServer(a, &Options{LanguageTag: language.SimplifiedChinese, Logs: o})

	// 自定义报头
	buf.Reset()
	w := httptest.NewRecorder()
	r := rest.Post(a, "/path", nil).
		Header("Content-Type", "application/json").
		Header("Accept", "application/json").
		Request()
	ctx := srv.newContext(w, r, nil)
	a.NotNil(ctx)
	ctx.Marshal(http.StatusCreated, testdata.ObjectInst, false)
	a.Zero(buf.Len()).
		Equal(w.Code, http.StatusCreated).
		Equal(w.Body.String(), testdata.ObjectJSONString).
		Equal(w.Header().Get("content-type"), header.BuildContentType("application/json", "utf-8")).
		Equal(w.Header().Get("content-language"), "zh-Hans")

	buf.Reset()
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").
		Header("accept", "application/json").
		Header("accept-language", "").
		Request()
	ctx = srv.newContext(w, r, nil)
	ctx.Marshal(http.StatusCreated, testdata.ObjectInst, false)
	a.Zero(buf.Len()).
		Equal(w.Code, http.StatusCreated).
		Equal(w.Body.String(), testdata.ObjectJSONString).
		Equal(w.Header().Get("content-language"), language.SimplifiedChinese.String()) // 未指定，采用默认值

	// 输出 nil，content-type 和 content-language 均为空
	buf.Reset()
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").
		Header("Accept", "application/json").
		Header("Accept-language", "zh-hans").
		Request()
	ctx = srv.newContext(w, r, nil)
	ctx.Marshal(http.StatusCreated, nil, false)
	a.Zero(buf.Len()).
		Equal(w.Code, http.StatusCreated).
		Equal(w.Body.String(), "").
		Equal(w.Header().Get("content-language"), ""). // 指定了输出语言，也返回空。
		Equal(w.Header().Get("content-Type"), "")

	// accept,accept-language,accept-charset
	buf.Reset()
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").
		Header("Accept", "application/json").
		Header("Accept-Language", "zh-Hans").
		Header("Accept-Charset", "gbk").
		Request()
	ctx = srv.newContext(w, r, nil)
	ctx.Marshal(http.StatusCreated, testdata.ObjectInst, false)
	a.Zero(buf.Len()).
		Equal(w.Code, http.StatusCreated).
		Equal(w.Body.Bytes(), testdata.ObjectGBKBytes)

	// problem
	buf.Reset()
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").Header("Accept", "application/json").Request()
	ctx = srv.newContext(w, r, nil)
	ctx.Marshal(http.StatusCreated, "abc", true)
	a.Zero(buf.Len()).
		Equal(w.Code, http.StatusCreated).
		Equal(w.Body.Bytes(), `"abc"`).Equal(w.Header().Get("content-type"), "application/problem+json; charset=utf-8")

	// problem, 未指定
	buf.Reset()
	srv.Mimetypes().Set("application/json", marshalJSON, json.Unmarshal, "")
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").Header("Accept", "application/json").Request()
	ctx = srv.newContext(w, r, nil)
	ctx.Marshal(http.StatusCreated, "abc", true)
	a.Zero(buf.Len()).
		Equal(w.Code, http.StatusCreated).
		Equal(w.Body.String(), `"abc"`).
		Equal(w.Header().Get("content-type"), "application/json; charset=utf-8")

	// 同时指定了 accept,accept-language,accept-charset 和 accept-encoding
	buf.Reset()
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").
		Header("Accept", "application/json").
		Header("Accept-Language", "zh-Hans").
		Header("Accept-Charset", "gbk").
		Header("Accept-Encoding", "gzip;q=0.9,deflate").
		Request()
	ctx = srv.newContext(w, r, nil)
	ctx.Marshal(http.StatusCreated, testdata.ObjectInst, false)
	ctx.destroy()
	a.Zero(buf.Len()).
		Equal(w.Code, http.StatusCreated)
	data, err := io.ReadAll(flate.NewReader(w.Body))
	a.NotError(err).Equal(data, testdata.ObjectGBKBytes).
		Equal(w.Header().Get("content-encoding"), "deflate")

	// 同时通过 ctx.Write 和 ctx.Marshal 输出内容
	buf.Reset()
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").
		Header("Accept", "application/json").
		Header("Accept-Encoding", "gzip;q=0.9,deflate").
		Request()
	ctx = srv.newContext(w, r, nil)
	_, err = ctx.Write([]byte("123"))
	a.NotError(err)
	ctx.Marshal(http.StatusCreated, "456", false)
	ctx.destroy()
	a.Zero(buf.Len()).Equal(w.Code, http.StatusCreated) // 压缩对象缓存了 WriteHeader 的发送
	data, err = io.ReadAll(flate.NewReader(w.Body))
	a.NotError(err).Equal(string(data), `123"456"`)

	// accept,accept-language,accept-charset 和 accept-encoding，部分 Response.Write 输出
	buf.Reset()
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").
		Header("Accept", "application/json").
		Header("Accept-Language", "zh-Hans").
		Header("Accept-Charset", "gbk").
		Request()
	ctx = srv.newContext(w, r, nil)
	_, err = ctx.Write([]byte(testdata.ObjectJSONString))
	a.NotError(err)
	ctx.Marshal(http.StatusCreated, testdata.ObjectInst, false)
	ctx.destroy()
	a.Zero(buf.Len()).Equal(w.Code, http.StatusOK) // 未指定压缩，WriteHeader 直接发送
	data, err = io.ReadAll(w.Body)
	a.NotError(err)
	bs := make([]byte, 0, 2*len(testdata.ObjectGBKBytes))
	bs = append(append(bs, testdata.ObjectGBKBytes...), testdata.ObjectGBKBytes...)
	a.Equal(data, bs)

	// outputMimetype == nil
	buf.Reset()
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").Header("Accept", "nil").Request()
	ctx = srv.newContext(w, r, nil)
	a.Nil(ctx.outputMimetype.Marshal).
		Equal(ctx.Mimetype(false), "nil").
		Equal(ctx.Charset(), header.UTF8Name)
	a.PanicString(func() {
		ctx.Marshal(http.StatusCreated, "val", false)
	}, "未对 nil 作处理")
	a.Equal(w.Code, http.StatusNotAcceptable)

	// outputMimetype 返回 ErrUnsupported
	buf.Reset()
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").Header("Accept", "application/test").Request()
	ctx = srv.newContext(w, r, nil)
	ctx.Marshal(http.StatusCreated, "任意值", false)
	a.NotZero(buf.Len()).
		Equal(w.Code, http.StatusNotAcceptable)

	// outputMimetype 返回错误
	buf.Reset()
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").Header("Accept", "application/test").Request()
	ctx = srv.newContext(w, r, nil)
	ctx.Marshal(http.StatusCreated, errors.New("error"), false)
	a.NotZero(buf.Len()).
		Equal(w.Code, http.StatusNotAcceptable)
}

func TestContext_SetWriter(t *testing.T) {
	a := assert.New(t, false)
	srv := newServer(a, &Options{LanguageTag: language.SimplifiedChinese})

	w := httptest.NewRecorder()
	r := rest.Get(a, "/path").
		Header("accept-language", "cmn-hant").
		Header("accept", "application/json").
		Request()
	ctx := srv.newContext(w, r, nil)
	a.NotNil(ctx)
	ctx.Write([]byte("abc"))
	a.Equal(w.Body.String(), "abc")

	a.PanicString(func() {
		buf := &bytes.Buffer{}
		ctx.SetWriter(func(w http.ResponseWriter) http.ResponseWriter { return &response{w: buf, ResponseWriter: w} })
	}, "已有内容输出，不可再更改！")

	// setWriter
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").
		Header("accept-language", "cmn-hant").
		Header("accept", "application/json").
		Request()
	ctx = srv.newContext(w, r, nil)
	w.Header().Set("h1", "v1")
	a.NotNil(ctx)
	buf := &bytes.Buffer{}
	ctx.SetWriter(func(w http.ResponseWriter) http.ResponseWriter { return &response{w: buf, ResponseWriter: w} })
	ctx.Header().Set("h2", "v2")
	ctx.Write([]byte("abc"))
	a.Equal(buf.String(), "abc").
		Empty(w.Body.String()).
		Equal(ctx.Header().Get("h1"), "v1").
		Equal(ctx.Header().Get("h2"), "v2")

	// 多次调用 setWriter
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").
		Header("accept-language", "cmn-hant").
		Header("accept", "application/json").
		Request()
	ctx = srv.newContext(w, r, nil)
	a.NotNil(ctx)
	buf1 := &bytes.Buffer{}
	buf2 := &bytes.Buffer{}
	ctx.SetWriter(func(w http.ResponseWriter) http.ResponseWriter { return &response{w: buf1, ResponseWriter: w} })
	ctx.SetWriter(func(w http.ResponseWriter) http.ResponseWriter { return &response{w: buf2, ResponseWriter: w} })
	ctx.Write([]byte("abc"))
	a.Equal(buf2.String(), "abc").Empty(buf1.String())

	// setWriter(nil)
	ctx = srv.newContext(w, r, nil)
	a.NotNil(ctx)
	a.PanicString(func() {
		ctx.SetWriter(nil)
	}, "参数 w 不能为空")
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
		Header("accept", "application/json").
		Request()
	ctx := srv.newContext(w, r, nil)
	a.NotNil(ctx)
	ctx.Marshal(http.StatusOK, ctx.Sprintf("test"), false)
	a.Equal(w.Body.String(), `"測試"`)

	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").
		Header("accept-language", "cmn-hans").
		Header("accept", "application/json").
		Request()
	ctx = srv.newContext(w, r, nil)
	a.NotNil(ctx)
	n, err := ctx.LocalePrinter().Fprintf(ctx, "test")
	a.NotError(err).Equal(n, len("测试"))
	a.Equal(w.Body.String(), "测试")
}
