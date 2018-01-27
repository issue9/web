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
	charset "golang.org/x/text/encoding"

	"github.com/issue9/web/encoding"
	"github.com/issue9/web/encoding/test"
)

func newContext(w http.ResponseWriter, r *http.Request, outputEncoding encoding.MarshalFunc, outputCharset charset.Encoding) *Context {
	return &Context{
		Response:       w,
		Request:        r,
		OutputCharset:  outputCharset,
		OutputEncoding: outputEncoding,

		InputEncoding: encoding.TextUnmarshal,
		InputCharset:  nil,
	}
}

func TestContext_Body(t *testing.T) {
	a := assert.New(t)
	r := httptest.NewRequest(http.MethodGet, "/path", bytes.NewBufferString("123"))
	r.Header.Set("Accept", "*/*")
	w := httptest.NewRecorder()
	ctx := newContext(w, r, encoding.TextMarshal, nil)

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
	ctx := newContext(w, r, encoding.TextMarshal, nil)

	obj := &test.TextObject{}
	a.NotError(ctx.Unmarshal(obj))
	a.Equal(obj.Name, "test").Equal(obj.Age, 123)

	o := &struct{}{}
	a.Error(ctx.Unmarshal(o))
}

func TestContext_Marshal(t *testing.T) {
	a := assert.New(t)
	r := httptest.NewRequest(http.MethodPost, "/path", nil)
	w := httptest.NewRecorder()
	ctx := newContext(w, r, encoding.TextMarshal, nil)

	obj := &test.TextObject{Name: "test", Age: 123}
	a.NotError(ctx.Marshal(http.StatusCreated, obj, nil))
	a.Equal(w.Code, http.StatusCreated)
	a.Equal(w.Body.String(), "test,123")
}

func TestContext_RenderStatus(t *testing.T) {
	a := assert.New(t)
	r := httptest.NewRequest(http.MethodGet, "/path", nil)
	w := httptest.NewRecorder()
	ctx := newContext(w, r, encoding.TextMarshal, nil)

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
