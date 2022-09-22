// SPDX-License-Identifier: MIT

package server

import (
	"compress/flate"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/assert/v3/rest"
	"golang.org/x/text/language"

	"github.com/issue9/web/internal/header"
	"github.com/issue9/web/serializer"
	"github.com/issue9/web/serializer/text"
	"github.com/issue9/web/serializer/text/testobject"
)

const (
	gbkString1 = "中文1,11"
	gbkString2 = "中文2,22"
)

var (
	gbkBytes1 = []byte{214, 208, 206, 196, 49, 44, 49, 49}
	gbkBytes2 = []byte{214, 208, 206, 196, 50, 44, 50, 50}
)

func TestContext_Marshal(t *testing.T) {
	a := assert.New(t, false)
	srv := newServer(a, &Options{LanguageTag: language.SimplifiedChinese})

	// 自定义报头
	w := httptest.NewRecorder()
	r := rest.Post(a, "/path", nil).
		Header("Content-Type", text.Mimetype).
		Header("Accept", text.Mimetype).
		Request()
	ctx := srv.newContext(w, r, nil)
	a.NotNil(ctx)
	obj := &testobject.TextObject{Name: "test", Age: 123}
	a.NotError(ctx.Marshal(http.StatusCreated, obj, false))
	a.Equal(w.Code, http.StatusCreated)
	a.Equal(w.Body.String(), "test,123")
	a.Equal(w.Header().Get("content-type"), header.BuildContentType(text.Mimetype, "utf-8"))
	a.Equal(w.Header().Get("content-language"), "zh-Hans")

	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").
		Header("accept", text.Mimetype).
		Header("accept-language", "").
		Request()
	ctx = srv.newContext(w, r, nil)
	obj = &testobject.TextObject{Name: "test", Age: 1234}
	a.NotError(ctx.Marshal(http.StatusCreated, obj, false))
	a.Equal(w.Code, http.StatusCreated)
	a.Equal(w.Body.String(), "test,1234")
	a.Equal(w.Header().Get("content-language"), language.SimplifiedChinese.String()) // 未指定，采用默认值

	// 输出 nil，content-type 和 content-language 均为空
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").
		Header("Accept", text.Mimetype).
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
		Header("Accept", text.Mimetype).
		Header("Accept-Language", "zh-Hans").
		Header("Accept-Charset", "gbk").
		Request()
	ctx = srv.newContext(w, r, nil)
	a.NotError(ctx.Marshal(http.StatusCreated, gbkString2, false))
	a.Equal(w.Code, http.StatusCreated)
	a.Equal(w.Body.Bytes(), gbkBytes2)

	// problem
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").Header("Accept", "application/json").Request()
	ctx = srv.newContext(w, r, nil)
	a.NotError(ctx.Marshal(http.StatusCreated, "abc", true))
	a.Equal(w.Code, http.StatusCreated)
	a.Equal(w.Body.Bytes(), `"abc"`).Equal(w.Header().Get("content-type"), "application/problem+json; charset=utf-8")

	// problem, 未指定
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").Header("Accept", text.Mimetype).Request()
	ctx = srv.newContext(w, r, nil)
	a.NotError(ctx.Marshal(http.StatusCreated, "abc", true))
	a.Equal(w.Code, http.StatusCreated)
	a.Equal(w.Body.Bytes(), "abc").Equal(w.Header().Get("content-type"), text.Mimetype+"; charset=utf-8")

	// 同时指定了 accept,accept-language,accept-charset 和 accept-encoding
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").
		Header("Accept", text.Mimetype).
		Header("Accept-Language", "zh-Hans").
		Header("Accept-Charset", "gbk").
		Header("Accept-Encoding", "gzip;q=0.9,deflate").
		Request()
	ctx = srv.newContext(w, r, nil)
	a.NotError(ctx.Marshal(http.StatusCreated, gbkString2, false))
	ctx.destroy()
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
	ctx = srv.newContext(w, r, nil)
	_, err = ctx.Write([]byte("123"))
	a.NotError(err)
	a.NotError(ctx.Marshal(http.StatusCreated, "456", false))
	ctx.destroy()
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
	ctx = srv.newContext(w, r, nil)
	_, err = ctx.Write([]byte(gbkString1))
	a.NotError(err)
	a.NotError(ctx.Marshal(http.StatusCreated, gbkString2, false))
	ctx.destroy()
	a.Equal(w.Code, http.StatusOK) // 未指定压缩，WriteHeader 直接发送
	data, err = io.ReadAll(w.Body)
	a.NotError(err)
	bs := make([]byte, 0, len(gbkBytes1)+len(gbkBytes2))
	bs = append(append(bs, gbkBytes1...), gbkBytes2...)
	a.Equal(data, bs)

	// outputMimetype == nil
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").Header("Accept", "nil").Request()
	ctx = srv.newContext(w, r, nil)
	a.Nil(ctx.outputMimetype).
		Equal(ctx.outputMimetypeName, "nil").
		Equal(ctx.outputCharsetName, header.UTF8Name)
	a.PanicString(func() {
		ctx.Marshal(http.StatusCreated, "val", false)
	}, "未对 nil 作处理")
	a.Equal(w.Code, http.StatusNotAcceptable)

	// outputMimetype 返回 serializer.ErrUnsupported
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").Header("Accept", text.Mimetype).Request()
	ctx = srv.newContext(w, r, nil)
	a.ErrorIs(ctx.Marshal(http.StatusCreated, &struct{}{}, false), serializer.ErrUnsupported)
	a.Equal(w.Code, http.StatusNotAcceptable)

	// outputMimetype 返回错误
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").Header("Accept", text.Mimetype).Request()
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
		Header("accept", text.Mimetype).
		Request()
	ctx := srv.newContext(w, r, nil)
	a.NotNil(ctx)
	a.NotError(ctx.Marshal(http.StatusOK, ctx.Sprintf("test"), false))
	a.Equal(w.Body.String(), "測試")

	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").
		Header("accept-language", "cmn-hans").
		Header("accept", text.Mimetype).
		Request()
	ctx = srv.newContext(w, r, nil)
	a.NotNil(ctx)
	n, err := ctx.LocalePrinter().Fprintf(ctx, "test")
	a.NotError(err).Equal(n, len("测试"))
	a.Equal(w.Body.String(), "测试")
}
