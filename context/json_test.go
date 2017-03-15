// Copyright 2015 by caixw, All rights reserved.
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

var _ Render = JSONRender

var _ Read = JSONRead

func TestJSONEnvelopeRender(t *testing.T) {
	a := assert.New(t)
	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "/index.php?a=b", nil)
	a.NotError(err).NotNil(r)
	ctx := newDefaultContext(w, r)

	// 缺少 Accept
	jsonEnvelopeRender(ctx, http.StatusOK, nil)
	a.Equal(w.Body.String(), `{"status":415}`)

	// 错误的 Accept
	w = httptest.NewRecorder()
	r.Header.Set("Accept", "test")
	ctx = newDefaultContext(w, r)
	jsonEnvelopeRender(ctx, http.StatusOK, nil)
	a.Equal(w.Body.String(), `{"status":415}`)
}

func TestJSONSetHeader(t *testing.T) {
	a := assert.New(t)

	w := httptest.NewRecorder()
	a.NotNil(w)

	jsonSetHeader(w, nil)
	a.Equal(w.Header().Get("Content-Type"), jsonContentType)

	jsonSetHeader(w, map[string]string{"Content-Type": "123"})
	a.Equal(w.Header().Get("Content-Type"), "123")
}

func TestJSONRender(t *testing.T) {
	a := assert.New(t)
	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "/index.php?a=b", nil)
	a.NotError(err).NotNil(r)
	ctx := newDefaultContext(w, r)

	// 少 accept
	JSONRender(ctx, http.StatusCreated, nil, nil)
	a.Equal(w.Code, http.StatusUnsupportedMediaType).Equal(w.Body.String(), "")

	// 错误的 accept
	w = httptest.NewRecorder()
	r.Header.Set("Accept", "test")
	ctx = newDefaultContext(w, r)
	JSONRender(ctx, http.StatusCreated, map[string]string{"name": "name"}, map[string]string{"h": "h"})
	a.Equal(w.Code, http.StatusUnsupportedMediaType)
	a.Equal(w.Body.String(), "")

	w = httptest.NewRecorder()
	r.Header.Set("Accept", jsonContentType)
	ctx = newDefaultContext(w, r)
	JSONRender(ctx, http.StatusCreated, map[string]string{"name": "name"}, map[string]string{"h": "h"})
	a.Equal(w.Code, http.StatusCreated)
	a.Equal(w.Body.String(), `{"name":"name"}`)
	a.Equal(w.Header().Get("h"), "h")

	// 解析 json 出错，会返回 500 错误
	w = httptest.NewRecorder()
	ctx = newDefaultContext(w, r)
	JSONRender(ctx, http.StatusOK, complex(5, 7), nil)
	a.Equal(w.Code, http.StatusInternalServerError)
	a.Equal(w.Body.String(), "")
}

func TestJSONRead(t *testing.T) {
	a := assert.New(t)

	w := httptest.NewRecorder()
	a.NotNil(w)
	r, err := http.NewRequest("POST", "/index.php?a=b", bytes.NewBufferString(`{"key":"1"}`))
	a.NotError(err).NotNil(r)
	ctx := newDefaultContext(w, r)

	// POST 少 Accept, Content-Type
	val := &struct {
		Key string `json:"key"`
	}{}
	ok := JSONRead(ctx, val)
	a.False(ok).
		Equal(w.Code, http.StatusUnsupportedMediaType).
		Equal(val.Key, "")

	// 正常解析
	w = httptest.NewRecorder()
	a.NotNil(w)
	r, err = http.NewRequest("GET", "/index.php?a=b", bytes.NewBufferString(`{"key":"1"}`))
	a.NotError(err).NotNil(r)
	ctx = newDefaultContext(w, r)
	val = &struct {
		Key string `json:"key"`
	}{}
	ok = JSONRead(ctx, val)
	a.True(ok).
		Equal(w.Code, http.StatusOK).
		Equal(val.Key, "1")

	// JSON 格式不正确，无法解析
	w = httptest.NewRecorder()
	a.NotNil(w)
	r, err = http.NewRequest("GET", "/index.php?a=b", bytes.NewBufferString(`{"key":1}`))
	a.NotError(err).NotNil(r)
	ctx = newDefaultContext(w, r)
	val = &struct {
		Key string `json:"key"`
	}{}
	ok = JSONRead(ctx, val)
	a.False(ok).
		Equal(w.Code, 422).
		Equal(val.Key, "")
}
