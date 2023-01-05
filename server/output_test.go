// SPDX-License-Identifier: MIT

package server

import (
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
	"github.com/issue9/web/serializer"
)

func TestContext_Marshal(t *testing.T) {
	a := assert.New(t, false)
	srv := newServer(a, &Options{LanguageTag: language.SimplifiedChinese})

	// 自定义报头
	w := httptest.NewRecorder()
	r := rest.Post(a, "/path", nil).
		Header("Content-Type", "application/json").
		Header("Accept", "application/json").
		Request()
	ctx := srv.newContext(w, r, nil)
	a.NotNil(ctx)
	a.NotError(ctx.Marshal(http.StatusCreated, testdata.ObjectInst, false))
	a.Equal(w.Code, http.StatusCreated)
	a.Equal(w.Body.String(), testdata.ObjectJSONString)
	a.Equal(w.Header().Get("content-type"), header.BuildContentType("application/json", "utf-8"))
	a.Equal(w.Header().Get("content-language"), "zh-Hans")

	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").
		Header("accept", "application/json").
		Header("accept-language", "").
		Request()
	ctx = srv.newContext(w, r, nil)
	a.NotError(ctx.Marshal(http.StatusCreated, testdata.ObjectInst, false))
	a.Equal(w.Code, http.StatusCreated)
	a.Equal(w.Body.String(), testdata.ObjectJSONString)
	a.Equal(w.Header().Get("content-language"), language.SimplifiedChinese.String()) // 未指定，采用默认值

	// 输出 nil，content-type 和 content-language 均为空
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").
		Header("Accept", "application/json").
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
		Header("Accept", "application/json").
		Header("Accept-Language", "zh-Hans").
		Header("Accept-Charset", "gbk").
		Request()
	ctx = srv.newContext(w, r, nil)
	a.NotError(ctx.Marshal(http.StatusCreated, testdata.ObjectInst, false))
	a.Equal(w.Code, http.StatusCreated)
	a.Equal(w.Body.Bytes(), testdata.ObjectGBKBytes)

	// problem
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").Header("Accept", "application/json").Request()
	ctx = srv.newContext(w, r, nil)
	a.NotError(ctx.Marshal(http.StatusCreated, "abc", true))
	a.Equal(w.Code, http.StatusCreated)
	a.Equal(w.Body.Bytes(), `"abc"`).Equal(w.Header().Get("content-type"), "application/problem+json; charset=utf-8")

	// problem, 未指定
	srv.Mimetypes().Set("application/json", marshalJSON, json.Unmarshal, "")
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").Header("Accept", "application/json").Request()
	ctx = srv.newContext(w, r, nil)
	a.NotError(ctx.Marshal(http.StatusCreated, "abc", true))
	a.Equal(w.Code, http.StatusCreated)
	a.Equal(w.Body.String(), `"abc"`).
		Equal(w.Header().Get("content-type"), "application/json; charset=utf-8")

	// 同时指定了 accept,accept-language,accept-charset 和 accept-encoding
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").
		Header("Accept", "application/json").
		Header("Accept-Language", "zh-Hans").
		Header("Accept-Charset", "gbk").
		Header("Accept-Encoding", "gzip;q=0.9,deflate").
		Request()
	ctx = srv.newContext(w, r, nil)
	a.NotError(ctx.Marshal(http.StatusCreated, testdata.ObjectInst, false))
	ctx.destroy()
	a.Equal(w.Code, http.StatusCreated)
	data, err := io.ReadAll(flate.NewReader(w.Body))
	a.NotError(err).Equal(data, testdata.ObjectGBKBytes)
	a.Equal(w.Header().Get("content-encoding"), "deflate")

	// 同时通过 ctx.Write 和 ctx.Marshal 输出内容
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").
		Header("Accept", "application/json").
		Header("Accept-Encoding", "gzip;q=0.9,deflate").
		Request()
	ctx = srv.newContext(w, r, nil)
	_, err = ctx.Write([]byte("123"))
	a.NotError(err)
	a.NotError(ctx.Marshal(http.StatusCreated, "456", false))
	ctx.destroy()
	a.Equal(w.Code, http.StatusCreated) // 压缩对象缓存了 WriteHeader 的发送
	data, err = io.ReadAll(flate.NewReader(w.Body))
	a.NotError(err).Equal(string(data), `123"456"`)

	// accept,accept-language,accept-charset 和 accept-encoding，部分 Response.Write 输出
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").
		Header("Accept", "application/json").
		Header("Accept-Language", "zh-Hans").
		Header("Accept-Charset", "gbk").
		Request()
	ctx = srv.newContext(w, r, nil)
	_, err = ctx.Write([]byte(testdata.ObjectJSONString))
	a.NotError(err)
	a.NotError(ctx.Marshal(http.StatusCreated, testdata.ObjectInst, false))
	ctx.destroy()
	a.Equal(w.Code, http.StatusOK) // 未指定压缩，WriteHeader 直接发送
	data, err = io.ReadAll(w.Body)
	a.NotError(err)
	bs := make([]byte, 0, 2*len(testdata.ObjectGBKBytes))
	bs = append(append(bs, testdata.ObjectGBKBytes...), testdata.ObjectGBKBytes...)
	a.Equal(data, bs)

	// outputMimetype == nil
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
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").Header("Accept", "application/test").Request()
	ctx = srv.newContext(w, r, nil)
	a.ErrorIs(ctx.Marshal(http.StatusCreated, "任意值", false), serializer.ErrUnsupported())
	a.Equal(w.Code, http.StatusNotAcceptable)

	// outputMimetype 返回错误
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").Header("Accept", "application/test").Request()
	ctx = srv.newContext(w, r, nil)
	a.Error(ctx.Marshal(http.StatusCreated, errors.New("error"), false))
	a.Equal(w.Code, http.StatusInternalServerError)
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
	a.NotError(ctx.Marshal(http.StatusOK, ctx.Sprintf("test"), false))
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
