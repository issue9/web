// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package context

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
)

func TestNew(t *testing.T) {
	a := assert.New(t)
	r := httptest.NewRequest(http.MethodGet, "/path", nil)
	w := httptest.NewRecorder()

	ctx, err := New(w, r, "not exists", DefaultCharset, true)
	a.Error(err).Nil(ctx)

	ctx, err = New(w, r, DefaultEncoding, "not exits", true)
	a.Error(err).Nil(ctx)

	// 缺少 Accept 报头
	ctx, err = New(w, r, DefaultEncoding, DefaultCharset, true)
	a.Equal(err, ErrClientNotAcceptable).Nil(ctx)

	// 不检测 Accept 报头
	ctx, err = New(w, r, DefaultEncoding, DefaultCharset, false)
	a.NotError(err).NotNil(ctx)

	// 指定 accept 报头
	r.Header.Set("Accept", DefaultEncoding)
	ctx, err = New(w, r, DefaultEncoding, DefaultCharset, true)
	a.NotError(err).NotNil(ctx)

	// 未指定 content-type，使用默认值。
	r.Method = http.MethodPost
	ctx, err = New(w, r, DefaultEncoding, DefaultCharset, true)
	a.NotError(err).NotNil(ctx)

	// 可用的 content-type
	r.Header.Set("Content-type", buildContentType(DefaultEncoding, DefaultCharset))
	ctx, err = New(w, r, DefaultEncoding, DefaultCharset, true)
	a.NotError(err).NotNil(ctx)

	// 不可用的 content-type.encoding
	r.Header.Set("Content-type", buildContentType("text/unknown", DefaultCharset))
	ctx, err = New(w, r, DefaultEncoding, DefaultCharset, true)
	a.Equal(err, ErrUnsupportedContentType).Nil(ctx)

	// 不可用的 content-type.charset
	r.Header.Set("Content-type", buildContentType(DefaultEncoding, "unknown"))
	ctx, err = New(w, r, DefaultEncoding, DefaultCharset, true)
	a.Equal(err, ErrUnsupportedContentType).Nil(ctx)
}

func TestContext_Body(t *testing.T) {
	a := assert.New(t)
	r := httptest.NewRequest(http.MethodGet, "/path", bytes.NewBufferString("123"))
	r.Header.Set("Accept", "*/*")
	w := httptest.NewRecorder()
	ctx, err := New(w, r, DefaultEncoding, DefaultCharset, true)
	a.NotError(err).NotNil(ctx)

	a.Nil(ctx.body) // 未缓存
	data, err := ctx.Body()
	a.NotError(err).Equal(data, []byte("123"))
	a.Equal(ctx.body, data)

	data, err = ctx.Body()
	a.NotError(err).Equal(data, []byte("123"))
	a.Equal(ctx.body, data)

	// TODO 编码
}

func TestContext_Unmarshal(t *testing.T) {
	a := assert.New(t)
	r := httptest.NewRequest(http.MethodPost, "/path", bytes.NewBufferString("test,123"))
	w := httptest.NewRecorder()
	ctx, err := New(w, r, DefaultEncoding, DefaultCharset, false)
	a.NotError(err).NotNil(ctx)

	obj := &textObject{}
	a.NotError(ctx.Unmarshal(obj))
	a.Equal(obj.Name, "test").Equal(obj.Age, 123)

	o := &struct{}{}
	a.Error(ctx.Unmarshal(o))
}

func TestContext_Marshal(t *testing.T) {
	a := assert.New(t)
	r := httptest.NewRequest(http.MethodPost, "/path", nil)
	w := httptest.NewRecorder()
	ctx, err := New(w, r, DefaultEncoding, DefaultCharset, false)
	a.NotError(err).NotNil(ctx)

	obj := &textObject{Name: "test", Age: 123}
	a.NotError(ctx.Marshal(http.StatusCreated, obj, nil))
	a.Equal(w.Code, http.StatusCreated)
	a.Equal(w.Body.String(), "test,123")
}

func TestContext_RenderStatus(t *testing.T) {
	a := assert.New(t)
	r := httptest.NewRequest(http.MethodGet, "/path", nil)
	w := httptest.NewRecorder()
	ctx, err := New(w, r, DefaultEncoding, DefaultCharset, false)
	a.NotError(err).NotNil(ctx)

	ctx.RenderStatus(http.StatusForbidden)
	a.Equal(w.Code, http.StatusForbidden)
}

func TestRenderStatus(t *testing.T) {
	a := assert.New(t)
	w := httptest.NewRecorder()

	RenderStatus(w, http.StatusOK)
	a.Equal(w.Code, http.StatusOK).
		Equal(w.Header().Get("Content-Type"), "text/plain; charset=utf-8")

	w = httptest.NewRecorder()
	RenderStatus(w, http.StatusInternalServerError)
	a.Equal(w.Code, http.StatusInternalServerError).
		Equal(w.Header().Get("Content-Type"), "text/plain; charset=utf-8")
}
